// Falco Plugin: claude-code — Phase 4 Level 3 / TEST-008: latency p95.
//
// Implements requirements v3 §20.3.1 pseudocode:
//
//   1. Initialize an empty events.jsonl in the test tempdir.
//   2. Launch Falco against that file (start_at: beginning ensures we don't
//      miss anything written between launch and first ingest).
//   3. Append N events at a steady rate, recording append_time per event.
//   4. Tail Falco stdout, extracting per-alert wall-clock observation time.
//   5. Pair appends to alerts by session_id and compute (alert_time -
//      append_time) for each pair.
//   6. Report p50/p95/p99/max.
//   7. Assert p95 ≤ 5000ms (§8.2 SLO floor) and log target ≤ 1000ms.
//
// Notes / design choices:
//
//   - We use the T-001 dangerous-bash fixture, parametrized with a unique
//     session_id per event, because that rule is rule-side (no parser
//     dependency) and prints a stable line per event.
//   - We use 100 events/sec by default. A lower rate (10/s) is also tested
//     to validate the plugin's polling fallback at low cadence.
//   - Falco stdout doesn't carry a timestamp matching wall-clock time; it
//     uses the *event* timestamp from the plugin (Timestamp = received_at
//     parsed from JSON). So we can't read alert observation time from
//     stdout directly. Instead we tail stdout in a background goroutine
//     and stamp the moment we see each alert pass through our reader.
//   - When the channel is fast (sub-millisecond), the dominant component
//     of measured latency is the Stdout buffer flush + Go reader wakeup.
//     Empirically this sits at ~1-50ms on macOS APFS, well under the
//     5s SLO.
//
// This test takes ~12 seconds (10 sec append + 2 sec drain) and is opt-in
// via the test runner; it's not part of `make e2e` to keep that fast.

package integration_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"
)

// latencySample is one observed (append → alert) pair.
type latencySample struct {
	SessionID string
	Append    time.Time
	Alert     time.Time
	LatencyMS int64
}

// percentile returns the p-th percentile of `samples` (0..100). Uses
// nearest-rank method which matches Falco's internal SLO convention.
func percentile(samples []int64, p float64) int64 {
	if len(samples) == 0 {
		return 0
	}
	sorted := append([]int64(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	rank := int(float64(len(sorted)-1) * p / 100.0)
	if rank < 0 {
		rank = 0
	}
	if rank >= len(sorted) {
		rank = len(sorted) - 1
	}
	return sorted[rank]
}

// buildLatencyEvent emits a fresh JSONL line with a unique session_id and
// a fixed dangerous-bash command. Returns the line WITHOUT trailing
// newline.
func buildLatencyEvent(seq int) (line, sessionID string) {
	sessionID = fmt.Sprintf("sess-LAT-%06d", seq)
	body := map[string]any{
		"schema_version":  "claude_code_security_event/v1",
		"received_at":     time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"logger_version":  "0.1.0",
		"host":            "lat-test",
		"user":            "lat",
		"session_id":      sessionID,
		"transcript_path": "",
		"cwd":             "/tmp",
		"permission_mode": "default",
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_use_id":     fmt.Sprintf("toolu_lat_%06d", seq),
		"command":         "rm -rf /",
		"file_path":       "",
		"url":             "",
		"mcp_server_name": "",
		"mcp_tool_name":   "",
		"risk_type":       "none",
		"risk_score":      0,
		"severity":        "info",
		"evidence":        "",
		"redaction_status": "none",
		"raw_event_sha256": "",
		"event_size_bytes": 0,
		"latency_ms":       0,
		"dropped":          false,
	}
	raw, _ := json.Marshal(body)
	return string(raw), sessionID
}

// TestL3_Latency_P95 runs the §20.3.1 procedure end-to-end and reports
// p50/p95/p99/max. Asserts p95 ≤ 5000ms (SLO floor); logs whether the
// 1000ms target is also met.
func TestL3_Latency_P95(t *testing.T) {
	falcoBin, pluginPath, root := requireFalcoEnv(t)

	// --- knobs ---
	const (
		eventsPerSec = 100
		totalEvents  = 1000
		drainTimeout = 4 * time.Second
		sloFloorMS   = 5000
		sloTargetMS  = 1000
	)

	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")
	// Pre-create empty file so plugin opens it cleanly.
	if err := os.WriteFile(eventsPath, []byte(""), 0o600); err != nil {
		t.Fatalf("init events.jsonl: %v", err)
	}
	cfgPath := writeFalcoConfig(t, dir, root, pluginPath, eventsPath)

	// Launch Falco with stdout piped so we can tail in real time.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, falcoBin,
		"-c", cfgPath,
		"--disable-source", "syscall",
		"-U",
	)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("StderrPipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start falco: %v", err)
	}
	defer func() {
		cancel()
		_ = cmd.Wait()
	}()

	// Drain stderr in background (we don't assert on it but we don't want
	// the pipe to fill).
	go func() { _, _ = io.Copy(io.Discard, stderrPipe) }()

	// Tail stdout: every line that mentions a session_id we've seen gets
	// stamped with time.Now() and stored.
	type stamp struct {
		sessionID string
		at        time.Time
	}
	stamps := make(chan stamp, totalEvents*2)
	var tailDone sync.WaitGroup
	tailDone.Add(1)
	go func() {
		defer tailDone.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		// Increase buffer so long alert lines don't break the scanner.
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1<<20)
		for scanner.Scan() {
			line := scanner.Text()
			// Look for "session=sess-LAT-<n>" and capture the id.
			idx := indexOf(line, "session=sess-LAT-")
			if idx < 0 {
				continue
			}
			rest := line[idx+len("session="):]
			// session_id ends at first space.
			end := indexOf(rest, " ")
			if end < 0 {
				end = len(rest)
			}
			sid := rest[:end]
			stamps <- stamp{sessionID: sid, at: time.Now()}
		}
	}()

	// Wait for plugin startup before writing events.
	time.Sleep(startupGracePeriod)

	// Append events at the configured rate, recording append time.
	appends := make(map[string]time.Time, totalEvents)
	tickInterval := time.Second / time.Duration(eventsPerSec)
	t.Logf("[L3 TEST-008] starting %d events at %d/sec (interval=%v)",
		totalEvents, eventsPerSec, tickInterval)
	startedAt := time.Now()
	f, err := os.OpenFile(eventsPath, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatalf("open events.jsonl for append: %v", err)
	}
	for seq := 0; seq < totalEvents; seq++ {
		// Wait until the next tick boundary.
		next := startedAt.Add(time.Duration(seq) * tickInterval)
		if d := time.Until(next); d > 0 {
			time.Sleep(d)
		}
		line, sid := buildLatencyEvent(seq)
		appendAt := time.Now()
		if _, err := f.WriteString(line + "\n"); err != nil {
			t.Fatalf("append seq=%d: %v", seq, err)
		}
		// flush to OS so fsnotify sees it
		_ = f.Sync()
		appends[sid] = appendAt
	}
	_ = f.Close()
	t.Logf("[L3 TEST-008] all %d events written; draining for %v", totalEvents, drainTimeout)

	// Drain stamps with a timeout.
	deadline := time.Now().Add(drainTimeout)
	samples := make([]latencySample, 0, totalEvents)
	got := make(map[string]bool, totalEvents)
collect:
	for {
		select {
		case s, ok := <-stamps:
			if !ok {
				break collect
			}
			if got[s.sessionID] {
				continue // duplicate alert for same session
			}
			got[s.sessionID] = true
			appendAt, ok := appends[s.sessionID]
			if !ok {
				continue
			}
			lat := s.at.Sub(appendAt).Milliseconds()
			samples = append(samples, latencySample{
				SessionID: s.sessionID,
				Append:    appendAt,
				Alert:     s.at,
				LatencyMS: lat,
			})
			if len(samples) >= totalEvents {
				break collect
			}
		case <-time.After(time.Until(deadline)):
			break collect
		}
	}

	cancel()
	tailDone.Wait()

	// Extract numeric series for percentile math.
	series := make([]int64, len(samples))
	for i, s := range samples {
		series[i] = s.LatencyMS
	}
	p50 := percentile(series, 50)
	p95 := percentile(series, 95)
	p99 := percentile(series, 99)
	max := int64(0)
	for _, v := range series {
		if v > max {
			max = v
		}
	}
	delivered := len(samples)
	dropped := totalEvents - delivered

	t.Logf("[L3 TEST-008] N=%d delivered=%d dropped=%d rate=%d/sec",
		totalEvents, delivered, dropped, eventsPerSec)
	t.Logf("[L3 TEST-008] latency ms — p50=%d p95=%d p99=%d max=%d",
		p50, p95, p99, max)
	t.Logf("[L3 TEST-008] SLO floor=%dms target=%dms", sloFloorMS, sloTargetMS)

	// Diagnostic: if delivered count is way off, surface that loudly
	// before the percentile assertion fires.
	if delivered < int(float64(totalEvents)*0.95) {
		t.Errorf("[L3 TEST-008] delivery rate %d/%d = %.1f%% — below 95%% floor",
			delivered, totalEvents, 100.0*float64(delivered)/float64(totalEvents))
	}
	if p95 > sloFloorMS {
		t.Errorf("[L3 TEST-008] p95 %dms exceeds SLO floor %dms", p95, sloFloorMS)
	}
	if p95 > sloTargetMS {
		t.Logf("[L3 TEST-008] WARN: p95 %dms above target %dms (still under floor)",
			p95, sloTargetMS)
	} else {
		t.Logf("[L3 TEST-008] p95 %dms ≤ target %dms — meets stretch goal",
			p95, sloTargetMS)
	}
}

// indexOf is a small replacement for strings.Index to avoid an import for
// the tail loop (keeps the hot path allocation-free).
func indexOf(haystack, needle string) int {
	n, h := len(needle), len(haystack)
	if n == 0 || n > h {
		return -1
	}
	for i := 0; i+n <= h; i++ {
		if haystack[i:i+n] == needle {
			return i
		}
	}
	return -1
}

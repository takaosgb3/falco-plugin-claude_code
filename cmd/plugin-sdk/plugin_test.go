package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/takaosgb3/falco-plugin-claude_code/pkg/parser"
)

// validJSONLLine returns one schema-valid claude_code JSONL line carrying
// `marker` in the command field so the test can identify which append the
// event came from.
func validJSONLLine(marker string) string {
	return `{` +
		`"schema_version":"claude_code_security_event/v1",` +
		`"received_at":"2026-04-26T12:00:00Z",` +
		`"session_id":"s",` +
		`"hook_event_name":"PreToolUse",` +
		`"tool_name":"Bash",` +
		`"command":"echo ` + marker + `"` +
		`}` + "\n"
}

// writeNLines appends N JSONL lines to path with the given marker prefix.
func writeNLines(t *testing.T, path string, n int, prefix string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	require.NoError(t, err)
	defer f.Close()
	for i := 0; i < n; i++ {
		_, err := f.WriteString(validJSONLLine(fmt.Sprintf("%s-%d", prefix, i)))
		require.NoError(t, err)
	}
	require.NoError(t, f.Sync())
}

// initPlugin creates a configured ClaudeCodePlugin tailing the given paths.
func initPlugin(t *testing.T, logPaths []string) *ClaudeCodePlugin {
	t.Helper()
	p := &ClaudeCodePlugin{}
	cfg := ClaudeCodeConfig{
		LogPaths:        logPaths,
		EventBufferSize: 1024,
		PollIntervalMs:  50, // tight poll to keep tests fast
	}
	p.config = cfg
	p.parser = parser.New(parser.Config{
		LogFormat:        "json",
		SecurityPatterns: true,
	})
	return p
}

// openAndCleanup opens an instance and registers the matching Close().
func openAndCleanup(t *testing.T, p *ClaudeCodePlugin) *ClaudeCodeInstance {
	t.Helper()
	src, err := p.Open("")
	require.NoError(t, err)
	inst := src.(*ClaudeCodeInstance)
	t.Cleanup(func() { inst.Close() })
	return inst
}

// drainEvents collects up to `want` events from the eventCh with a deadline.
func drainEvents(t *testing.T, inst *ClaudeCodeInstance, want int, timeout time.Duration) []*ClaudeCodeEvent {
	t.Helper()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	out := make([]*ClaudeCodeEvent, 0, want)
	for len(out) < want {
		select {
		case ev, ok := <-inst.eventCh:
			if !ok {
				return out
			}
			out = append(out, ev)
		case <-deadline.C:
			return out
		}
	}
	return out
}

// --- TestPipeline_FsNotify ---

// Append → fsnotify Write → parseLine → eventCh delivery.
func TestPipeline_FsNotify(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)

	// Allow the watcher to settle (P021 fsnotify timing).
	time.Sleep(100 * time.Millisecond)

	writeNLines(t, path, 3, "fsnotify")

	got := drainEvents(t, inst, 3, 5*time.Second)
	require.Len(t, got, 3)
	for _, ev := range got {
		assert.Contains(t, ev.Command, "fsnotify")
	}
}

// --- TestPipeline_Polling ---

// Disable fsnotify by stopping the watcher; ensure pollLoop catches up.
func TestPipeline_Polling(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)

	// Cripple fsnotify by closing the watcher; pollLoop must still deliver.
	inst.watcher.Close()
	time.Sleep(50 * time.Millisecond)

	writeNLines(t, path, 4, "polling")

	got := drainEvents(t, inst, 4, 5*time.Second)
	require.Len(t, got, 4)
	for _, ev := range got {
		assert.Contains(t, ev.Command, "polling")
	}
}

// --- TestPipeline_Rotation_Rename (§20.2.1) ---

// 1) write 10 lines pre-rotation → SeekEnd skips them
// 2) rename the file
// 3) create a fresh file at the same path
// 4) write 5 post-rotation lines
// 5) only post-rotation events appear on eventCh.
func TestPipeline_Rotation_Rename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	writeNLines(t, path, 10, "rotation-pre")

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)

	// Settle.
	time.Sleep(100 * time.Millisecond)

	// rename rotation
	require.NoError(t, os.Rename(path, path+".1"))

	// fresh file at original path (new inode)
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	time.Sleep(100 * time.Millisecond) // P021 fsnotify timing
	writeNLines(t, path, 5, "rotation-post")

	got := drainEvents(t, inst, 5, 5*time.Second)
	require.Len(t, got, 5, "expected 5 post-rotation events, got %d", len(got))
	for _, ev := range got {
		assert.Contains(t, ev.Command, "rotation-post",
			"pre-rotation events leaked through; SeekEnd is not honoured")
	}
	assert.Greater(t, atomic.LoadUint64(&inst.rotationEvents), uint64(0))
	assert.Greater(t, atomic.LoadUint64(&inst.reopenEvents), uint64(0))
}

// --- TestPipeline_Rotation_Truncate ---

// truncate-in-place (`> events.jsonl`) keeps the same inode; the parser
// must detect via offset-decrease and reset.
func TestPipeline_Rotation_Truncate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(100 * time.Millisecond)

	writeNLines(t, path, 5, "pre-trunc")
	// drain
	_ = drainEvents(t, inst, 5, 3*time.Second)

	// truncate in place
	require.NoError(t, os.Truncate(path, 0))
	time.Sleep(150 * time.Millisecond) // P021 fsnotify timing for poll cycle

	writeNLines(t, path, 3, "post-trunc")

	got := drainEvents(t, inst, 3, 5*time.Second)
	require.Len(t, got, 3)
	for _, ev := range got {
		assert.Contains(t, ev.Command, "post-trunc")
	}
}

// --- TestPipeline_Backpressure ---

// EventBufferSize=2 → producer overruns the channel → drop counter increments.
func TestPipeline_Backpressure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := &ClaudeCodePlugin{}
	p.config = ClaudeCodeConfig{
		LogPaths:        []string{path},
		EventBufferSize: 2,
		PollIntervalMs:  50,
	}
	p.parser = parser.New(parser.Config{LogFormat: "json"})

	src, err := p.Open("")
	require.NoError(t, err)
	inst := src.(*ClaudeCodeInstance)
	t.Cleanup(func() { inst.Close() })

	time.Sleep(100 * time.Millisecond)

	// burst-write 50 lines; only ~2 should fit in the channel and the rest
	// must be counted as dropped.
	writeNLines(t, path, 50, "burst")
	time.Sleep(500 * time.Millisecond)

	dropped := atomic.LoadUint64(&inst.droppedEvents)
	assert.Greater(t, dropped, uint64(0),
		"expected drop counter to fire under backpressure; got %d", dropped)
}

// --- TestPipeline_Close_NoLeak (P011) ---

func TestPipeline_Close_NoLeak(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(50 * time.Millisecond)

	// Close should not deadlock and should be idempotent.
	inst.Close()
	inst.Close() // idempotent via closeOnce

	// After Close, events ch is closed; further sends are gated.
	// (We rely on -race / wg.Wait to catch goroutine leaks.)
}

// --- TestPipeline_GOBRoundTrip ---

// Encode → decode the ClaudeCodeEvent using gob to confirm the Headers map
// initialization (P004) holds end-to-end.
func TestPipeline_GOBRoundTrip(t *testing.T) {
	src := &ClaudeCodeEvent{
		LogPath:       "/tmp/x",
		Raw:           "raw",
		Headers:       map[string]string{"trace": "abc"},
		Command:       "echo ok",
		SessionID:     "s",
		EventName:     "PreToolUse",
		SchemaVersion: "claude_code_security_event/v1",
	}
	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(src))

	var dst ClaudeCodeEvent
	require.NoError(t, gob.NewDecoder(&buf).Decode(&dst))
	assert.Equal(t, src.Raw, dst.Raw)
	assert.Equal(t, src.Headers["trace"], dst.Headers["trace"])
}

// --- TestPipeline_PathTraversalRejected ---

func TestPipeline_PathTraversalRejected(t *testing.T) {
	p := &ClaudeCodePlugin{}
	p.config = ClaudeCodeConfig{LogPaths: []string{"../etc/passwd"}, EventBufferSize: 16}
	p.parser = parser.New(parser.Config{LogFormat: "json"})
	_, err := p.Open("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal")
}

// --- TestPipeline_StartAtBeginning ---

// When start_at=beginning is configured, the pre-existing lines should be
// drained on Open() (replay mode, §14.1 P-007).
func TestPipeline_StartAtBeginning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	writeNLines(t, path, 4, "replay")

	p := &ClaudeCodePlugin{}
	p.config = ClaudeCodeConfig{
		LogPaths:        []string{path},
		EventBufferSize: 1024,
		PollIntervalMs:  50,
		StartAt:         "beginning",
	}
	p.parser = parser.New(parser.Config{LogFormat: "json"})

	src, err := p.Open("")
	require.NoError(t, err)
	inst := src.(*ClaudeCodeInstance)
	t.Cleanup(func() { inst.Close() })

	got := drainEvents(t, inst, 4, 5*time.Second)
	require.Len(t, got, 4, "replay mode must read pre-existing lines")
}

// --- TestPipeline_LargeLine (P012) ---

// A 60KB JSON line must be readable (logger caps raw at 64KB).
func TestPipeline_LargeLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(100 * time.Millisecond)

	bigCmd := strings.Repeat("a", 60*1024)
	line := `{"schema_version":"claude_code_security_event/v1","received_at":"2026-04-26T12:00:00Z","session_id":"s","hook_event_name":"PreToolUse","command":"` + bigCmd + `"}` + "\n"
	require.NoError(t, appendString(path, line))

	got := drainEvents(t, inst, 1, 5*time.Second)
	require.Len(t, got, 1, "large line must be readable end-to-end (P012)")
	// Parser caps at MaxFieldLength (default 10KB), so the redelivered
	// command field is shorter than the input.
	assert.LessOrEqual(t, len(got[0].Command), 64*1024)
}

func appendString(path, s string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(f, s); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

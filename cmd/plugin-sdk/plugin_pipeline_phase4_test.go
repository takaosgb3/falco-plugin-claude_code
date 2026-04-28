// Phase 4 Level 2 — Pipeline test extensions
//
// These tests complement the original `plugin_test.go` by exercising the
// full Open() → fsnotify → parser → eventCh → NextBatch() pipeline against
// real fixtures from `test/fixtures/hook_events/`. They also assert on the
// observability counters (rotation, dropped, redacted) that production
// SLOs depend on.
//
// References:
//   - Requirements v3 §20.1 (3-layer E2E)
//   - Requirements v3 §20.2.1 (rotation_scenario 5 steps)
//   - Requirements v3 §20.3 TEST-005 (pipeline test assertions)
//   - Requirements v3 §17.1 (redaction)
//   - PROBLEM_PATTERNS P004, P006, P011, P014, P015, P020, P021
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/takaosgb3/falco-plugin-claude_code/pkg/parser"
	"github.com/takaosgb3/falco-plugin-claude_code/pkg/testutil"
)

// repoRoot returns the path to the repository root from this test file.
// `cmd/plugin-sdk/plugin_pipeline_phase4_test.go` is two levels deep.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return root
}

// fixtureLine returns the JSONL representation of a fixture (with `_meta`
// stripped) suitable for appending to an events.jsonl file.
func fixtureLine(t *testing.T, path string) string {
	t.Helper()
	c, err := testutil.LoadFixture(path)
	require.NoError(t, err, "load fixture %s", path)
	return c.JSONLine + "\n"
}

// appendFixture writes the given fixture's JSONL line to the given path.
func appendFixture(t *testing.T, eventsPath, fixturePath string) {
	t.Helper()
	line := fixtureLine(t, fixturePath)
	f, err := os.OpenFile(eventsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	require.NoError(t, err)
	defer f.Close()
	_, err = io.WriteString(f, line)
	require.NoError(t, err)
	require.NoError(t, f.Sync())
}

// --- TestPipeline_RotationScenario_FromFixtures ---
//
// §20.2.1 rotation_scenario, but driven by real fixtures so we exercise the
// schema validator end-to-end as well. Rename rotation:
//
//	1) write 10 pre-rotation lines (parser_baseline marker)
//	2) plugin Open() — SeekEnd skips the 10 baseline lines
//	3) rename events.jsonl → events.jsonl.1 (rotation event)
//	4) recreate events.jsonl, append 5 fixture lines (post-rotation)
//	5) drain eventCh → only the 5 post-rotation fixtures appear
func TestPipeline_RotationScenario_FromFixtures(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")

	// 1) pre-rotation seeded lines
	writeNLines(t, path, 10, "rotation-baseline")

	// 2) Open — must SeekEnd
	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(150 * time.Millisecond) // P021 fsnotify settle

	// 3) rename rotation
	require.NoError(t, os.Rename(path, path+".1"))

	// 4) fresh inode + 5 fixture-backed events
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	time.Sleep(150 * time.Millisecond)

	root := testutil.FixturesRoot(repoRoot(t))
	postRotationFixtures := []string{
		filepath.Join(root, "PreToolUse", "T-002-secret-exfil-curl.json"),
		filepath.Join(root, "PreToolUse", "T-003-bypass-mode.json"),
		filepath.Join(root, "ConfigChange", "T-005-settings-json.json"),
		filepath.Join(root, "ConfigChange", "T-006-hook-disabled.json"),
		filepath.Join(root, "PreToolUse", "T-009-sensitive-env-read.json"),
	}
	for _, fx := range postRotationFixtures {
		appendFixture(t, path, fx)
	}

	// 5) drain post-rotation fixtures only
	got := drainEvents(t, inst, 5, 5*time.Second)
	require.Len(t, got, 5,
		"expected exactly 5 post-rotation fixture events, got %d", len(got))

	// pre-rotation marker must NOT appear in the post-rotation drain.
	for _, ev := range got {
		assert.NotContains(t, ev.Command, "rotation-baseline",
			"SeekEnd contract violated: pre-rotation lines leaked through")
	}

	// Verify counter telemetry (P006 / FP-011): rotation + reopen events.
	assert.Greater(t, atomic.LoadUint64(&inst.rotationEvents), uint64(0),
		"rotation counter must increment on rename")
	assert.Greater(t, atomic.LoadUint64(&inst.reopenEvents), uint64(0),
		"reopen counter must increment when new inode is opened")
}

// --- TestPipeline_PollingFallback_FromFixtures ---
//
// fsnotify is the primary event source but the polling fallback is what
// keeps the plugin alive on Spotlight-indexed dirs / NFS / SMB. Cripple
// fsnotify mid-flight and verify pollLoop catches up *real* fixture lines
// (not just synthetic ones). This covers P021 fsnotify timing.
func TestPipeline_PollingFallback_FromFixtures(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(100 * time.Millisecond)

	// Take fsnotify out of the picture: the watcher Close() returns and
	// the read loop's select will fall through to poll-only.
	inst.watcher.Close()
	time.Sleep(80 * time.Millisecond)

	root := testutil.FixturesRoot(repoRoot(t))
	fixtures := []string{
		filepath.Join(root, "PreToolUse", "T-011-git-force-push.json"),
		filepath.Join(root, "PermissionRequest", "T-004-allow-userSettings.json"),
		filepath.Join(root, "ConfigChange", "T-007-mcp-config.json"),
	}
	for _, fx := range fixtures {
		appendFixture(t, path, fx)
	}

	// Polling tick is 50ms (initPlugin); 5 seconds is plenty.
	got := drainEvents(t, inst, len(fixtures), 5*time.Second)
	require.Len(t, got, len(fixtures),
		"polling fallback failed to deliver %d fixture events; got %d",
		len(fixtures), len(got))

	// Each event must carry the parser-side classification — proving the
	// detector ran on the polled-in line.
	risks := map[string]bool{}
	for _, ev := range got {
		risks[ev.RiskType] = true
	}
	assert.True(t, risks["git_destructive"], "T-011 risk_type missing")
	assert.True(t, risks["permission_update"], "T-004 risk_type missing")
	assert.True(t, risks["mcp_config_changed"], "T-007 risk_type missing")
}

// --- TestPipeline_RedactionEndToEnd ---
//
// §17.1 redaction must run at the parser layer (defense in depth) so that
// any secret-like token the upstream logger missed is masked before the
// event leaves the plugin. We inject a synthetic line that contains an AWS
// access key in `command`, run it through the full pipeline, and verify
// (a) the leaked token is replaced with the canonical mask, and
// (b) `redaction_status` is upgraded to "redacted".
func TestPipeline_RedactionEndToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(100 * time.Millisecond)

	// Synthetic event: logger forgot to redact, plugin must catch it.
	// AKIA-prefix is the deterministic AWS Access Key shape.
	leakedLine := `{` +
		`"schema_version":"claude_code_security_event/v1",` +
		`"received_at":"2026-04-26T12:00:00Z",` +
		`"session_id":"redact-e2e",` +
		`"hook_event_name":"PreToolUse",` +
		`"tool_name":"Bash",` +
		`"command":"echo AKIAIOSFODNN7EXAMPLE",` +
		`"redaction_status":"none"` +
		`}` + "\n"

	require.NoError(t, appendString(path, leakedLine))

	got := drainEvents(t, inst, 1, 5*time.Second)
	require.Len(t, got, 1, "leaked-secret event was not delivered")

	ev := got[0]
	assert.NotContains(t, ev.Command, "AKIAIOSFODNN7EXAMPLE",
		"raw AWS access key leaked through the parser layer (§17.1 violation)")
	// Mask format is `***REDACTED:<kind>***` (see redaction.go mask()).
	// We check for the canonical prefix to allow `<kind>` variation as the
	// pattern set evolves.
	assert.Contains(t, ev.Command, "***REDACTED:",
		"redaction mask missing — Redact() did not run")
	assert.Contains(t, ev.Command, "aws_access_key_id",
		"redaction kind tag missing for AWS access key")
	assert.Equal(t, "redacted", ev.RedactionStatus,
		"redaction_status must be upgraded to 'redacted'")
}

// --- TestPipeline_MalformedLineSkip ---
//
// The plugin must tolerate broken JSON lines without losing subsequent
// well-formed lines (P-002 / FP-009). Append a broken line followed by
// good fixtures and verify only the good lines reach the channel.
func TestPipeline_MalformedLineSkip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(100 * time.Millisecond)

	// 1) clearly broken JSON
	require.NoError(t, appendString(path, "{this is not json\n"))
	// 2) missing required fields (no schema_version)
	require.NoError(t, appendString(path,
		`{"received_at":"2026-04-26T12:00:00Z","session_id":"x","hook_event_name":"PreToolUse"}`+"\n"))
	// 3) unsupported schema family
	require.NoError(t, appendString(path,
		`{"schema_version":"some_other_schema/v1","received_at":"2026-04-26T12:00:00Z","session_id":"x","hook_event_name":"PreToolUse"}`+"\n"))
	// 4) good fixture line
	root := testutil.FixturesRoot(repoRoot(t))
	appendFixture(t, path, filepath.Join(root, "PreToolUse", "T-002-secret-exfil-curl.json"))
	// 5) another good fixture
	appendFixture(t, path, filepath.Join(root, "ConfigChange", "T-005-settings-json.json"))

	got := drainEvents(t, inst, 2, 5*time.Second)
	require.Len(t, got, 2,
		"expected exactly the 2 well-formed events to be delivered; got %d", len(got))

	// Both must be the parser-classified variants.
	risks := map[string]bool{}
	for _, ev := range got {
		risks[ev.RiskType] = true
	}
	assert.True(t, risks["secret_exfiltration"])
	assert.True(t, risks["settings_modified"])
}

// --- TestPipeline_LatencyBudget ---
//
// Lower bound: plugin-internal latency from append to NextBatch arrival.
// The §8.3 SLO budget is 100 ms p95; we verify the much tighter
// "single-event median latency under 50 ms" because no Falco round-trip is
// in the picture here. testing.Short() bails out (CI may run on slower
// hardware).
func TestPipeline_LatencyBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("latency budget test skipped in -short mode")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(100 * time.Millisecond) // settle fsnotify

	// One fixture line per round; we run 30 rounds and look at the p95 of
	// observed deliveries.
	const N = 30
	root := testutil.FixturesRoot(repoRoot(t))
	fxPath := filepath.Join(root, "PreToolUse", "T-002-secret-exfil-curl.json")
	line := fixtureLine(t, fxPath)

	latencies := make([]time.Duration, 0, N)
	for i := 0; i < N; i++ {
		t0 := time.Now()
		require.NoError(t, appendString(path, line))

		got := drainEvents(t, inst, 1, 2*time.Second)
		require.Len(t, got, 1, "round %d: missing event", i)
		latencies = append(latencies, time.Since(t0))
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p50 := latencies[len(latencies)/2]
	p95 := latencies[int(float64(len(latencies))*0.95)]

	t.Logf("plugin-internal latency: p50=%s p95=%s max=%s",
		p50, p95, latencies[len(latencies)-1])

	// Budget: plugin-internal latency p95 well under the §8.3 100 ms SLO.
	// 50 ms gives plenty of headroom for slower CI runners.
	const budget = 50 * time.Millisecond
	assert.Less(t, p95, budget,
		"plugin-internal p95 latency %s exceeds budget %s", p95, budget)
}

// --- TestPipeline_FixtureIngestion ---
//
// Drives every hook_events fixture through the live pipeline (Open ->
// fsnotify -> parser -> NextBatch) and verifies each fixture's
// `expected_risk_type` round-trips end to end. This is the strongest
// "fixtures match production wiring" guarantee we have short of Level 3.
func TestPipeline_FixtureIngestion(t *testing.T) {
	root := testutil.FixturesRoot(repoRoot(t))
	cases := testutil.LoadAllFixtures(t, root)
	require.GreaterOrEqual(t, len(cases), 20,
		"L2 ingestion needs at least 20 fixtures")

	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	// EventBufferSize is generous: we want zero drops on any fixture.
	p := &ClaudeCodePlugin{}
	p.config = ClaudeCodeConfig{
		LogPaths:        []string{path},
		EventBufferSize: 1024,
		PollIntervalMs:  50,
	}
	p.parser = parser.New(parser.Config{LogFormat: "json", SecurityPatterns: true})
	src, err := p.Open("")
	require.NoError(t, err)
	inst := src.(*ClaudeCodeInstance)
	t.Cleanup(func() { inst.Close() })

	time.Sleep(100 * time.Millisecond)

	for _, c := range cases {
		require.NoError(t, appendString(path, c.JSONLine+"\n"))
	}

	got := drainEvents(t, inst, len(cases), 10*time.Second)
	require.Len(t, got, len(cases),
		"ingested %d events but received %d at the channel", len(cases), len(got))

	// Cross-reference each delivered event against the fixture index.
	bySession := make(map[string]*ClaudeCodeEvent, len(got))
	for _, ev := range got {
		bySession[ev.SessionID] = ev
	}

	for _, c := range cases {
		// Re-extract the session id from the JSON line for matching.
		sid := extractStringField(c.JSONLine, "session_id")
		require.NotEmpty(t, sid, "fixture %s has empty session_id", c.Path)

		ev, ok := bySession[sid]
		require.True(t, ok, "fixture %s did not produce an event (session_id=%s)", c.Path, sid)

		if c.Meta.ExpectedRiskType != "" {
			assert.Equal(t, c.Meta.ExpectedRiskType, ev.RiskType,
				"fixture %s: pipeline risk_type mismatch", c.Path)
		}
		if c.Meta.ExpectedEventName != "" {
			assert.Equal(t, c.Meta.ExpectedEventName, ev.EventName,
				"fixture %s: event_name mismatch", c.Path)
		}
	}

	// Drop counter must remain zero on this generous buffer.
	assert.Equal(t, uint64(0), atomic.LoadUint64(&inst.droppedEvents),
		"unexpected drops during fixture ingestion")
}

// extractStringField is a tiny JSON-string field grabber used by the
// fixture-ingestion test. We avoid importing encoding/json again here and
// keep the helper local for clarity.
func extractStringField(line, field string) string {
	key := `"` + field + `":"`
	i := strings.Index(line, key)
	if i < 0 {
		return ""
	}
	rest := line[i+len(key):]
	j := strings.IndexByte(rest, '"')
	if j < 0 {
		return ""
	}
	return rest[:j]
}

// --- TestPipeline_MultipleLogPaths ---
//
// Production deployments may watch several events.jsonl files at once
// (per-user logger fan-out). Verify open / append / drain across two
// files.
func TestPipeline_MultipleLogPaths(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "events-a.jsonl")
	pathB := filepath.Join(dir, "events-b.jsonl")
	require.NoError(t, os.WriteFile(pathA, nil, 0o600))
	require.NoError(t, os.WriteFile(pathB, nil, 0o600))

	p := initPlugin(t, []string{pathA, pathB})
	inst := openAndCleanup(t, p)
	time.Sleep(150 * time.Millisecond)

	for i := 0; i < 3; i++ {
		require.NoError(t, appendString(pathA, validJSONLLine(fmt.Sprintf("a-%d", i))))
	}
	for i := 0; i < 4; i++ {
		require.NoError(t, appendString(pathB, validJSONLLine(fmt.Sprintf("b-%d", i))))
	}

	got := drainEvents(t, inst, 7, 5*time.Second)
	require.Len(t, got, 7, "expected events from both log files")

	pathSet := map[string]bool{}
	for _, ev := range got {
		pathSet[ev.LogPath] = true
	}
	assert.True(t, pathSet[pathA], "events from %s missing", pathA)
	assert.True(t, pathSet[pathB], "events from %s missing", pathB)
}

// --- TestPipeline_CountersSnapshot ---
//
// The parser counters (Counters API in pkg/parser) are observability
// surface for §6.3 FP-011. Push a malformed line + a redactable line and
// verify the counters move.
func TestPipeline_CountersSnapshot(t *testing.T) {
	p := parser.New(parser.Config{
		LogFormat:        "json",
		SecurityPatterns: true,
		MaxFieldLength:   1024,
	})

	before := p.CountersSnapshot()

	// 1) malformed line (broken JSON)
	_, err := p.Parse("{not json")
	require.Error(t, err)

	// 2) redactable line (AWS access key in command)
	good := `{` +
		`"schema_version":"claude_code_security_event/v1",` +
		`"received_at":"2026-04-26T12:00:00Z",` +
		`"session_id":"counter-test",` +
		`"hook_event_name":"PreToolUse",` +
		`"tool_name":"Bash",` +
		`"command":"echo AKIAIOSFODNN7EXAMPLE secret"` +
		`}`
	entry, err := p.Parse(good)
	require.NoError(t, err)
	require.NotNil(t, entry)

	after := p.CountersSnapshot()

	assert.Greater(t, after.Malformed, before.Malformed,
		"malformed counter must increment on broken JSON")
	assert.Greater(t, after.Redacted, before.Redacted,
		"redacted counter must increment when secret is masked")

	// JSON encode/decode round-trip the snapshot — the Counters API exposes
	// these counters for the doctor CLI / health check (§22.4 OPS-005).
	// Field names are Go-default capitalized (Malformed/Redacted/Detected);
	// the doctor CLI may transform the case at the boundary if needed.
	buf, err := json.Marshal(after)
	require.NoError(t, err)
	assert.Contains(t, string(buf), `"Redacted"`,
		"Counters JSON must include the Redacted counter")
	assert.Contains(t, string(buf), `"Malformed"`,
		"Counters JSON must include the Malformed counter")
}

// --- TestPipeline_HeartbeatPassesThrough ---
//
// The hook logger emits `_heartbeat_` events at heartbeat intervals
// (§22.4 OPS-002). The plugin must accept them and forward them through
// the pipeline; the rule layer is what decides whether to alert.
func TestPipeline_HeartbeatPassesThrough(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	p := initPlugin(t, []string{path})
	inst := openAndCleanup(t, p)
	time.Sleep(100 * time.Millisecond)

	root := testutil.FixturesRoot(repoRoot(t))
	appendFixture(t, path,
		filepath.Join(root, "_heartbeat_", "heartbeat-ok.json"))

	got := drainEvents(t, inst, 1, 5*time.Second)
	require.Len(t, got, 1)
	assert.Equal(t, "_heartbeat_", got[0].EventName)
	// Heartbeat must not be misclassified.
	assert.Contains(t, []string{"", "none"}, got[0].RiskType,
		"heartbeat must not produce a risk classification")
}

// Package integration_test provides Phase 4 Level 3 (Falco-in-the-loop)
// integration tests. These tests require a real `falco` binary on PATH (or
// at ~/bin/falco) and the plugin's compiled .dylib next to the repo root.
//
// Reference:
//   - Requirements v3 §20.1   (3-layer E2E split)
//   - Requirements v3 §20.3   (TEST-006 rule-firing, TEST-007 native smoke,
//                              TEST-008 latency p95)
//   - Requirements v3 §31     (AT-1..AT-5 acceptance)
//   - PROBLEM_PATTERNS P003 (source: required), P008 (load_plugins),
//                       P014 (SeekEnd → start_at: beginning), P017/P018
//                       (macOS Falco quirks).
//
// Test layout:
//
//   test/integration/
//     helpers.go               — config templating, fixture stripping, falco runner
//     falco_alerts_test.go     — TEST-006 (all 18 categories)
//     falco_smoke_test.go      — TEST-007 (config struct check)
//     latency_test.go          — TEST-008 (p95 / p99 / max)
//     acceptance_test.go       — AT-1..AT-5 acceptance summary
//
// All tests skip cleanly if the prerequisites are missing, so this package
// is safe to include in `go test ./...` on bare developer machines.
package integration_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	// fixtureMetaKey is the key the fixture loader strips before writing to
	// the JSONL stream consumed by Falco. Production payloads do not have
	// this key.
	fixtureMetaKey = "_meta"

	// alertGracePeriod is the time we wait for Falco to ingest a freshly
	// appended fixture and emit (or not emit) an alert. 1.5s is empirically
	// safe — fsnotify on macOS APFS observes writes within ~50ms, and the
	// plugin's 250ms poll fallback ensures even fsnotify drops are caught.
	alertGracePeriod = 1500 * time.Millisecond

	// startupGracePeriod is the time we wait after launching Falco before
	// considering the plugin "ready". Falco prints `Loading rules from file`
	// then `Starting health webserver` ... we just wait a fixed budget.
	startupGracePeriod = 800 * time.Millisecond
)

// repoRoot returns the absolute path to the repository root (two levels up
// from this package: test/integration/ → test/ → repo).
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repoRoot: %v", err)
	}
	return root
}

// findFalcoBinary returns the absolute path to a usable falco binary or ""
// if none is available. Order: ~/bin/falco → falco on PATH.
func findFalcoBinary() string {
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, "bin", "falco")
		if st, err := os.Stat(candidate); err == nil && st.Mode()&0o111 != 0 {
			return candidate
		}
	}
	if path, err := exec.LookPath("falco"); err == nil {
		return path
	}
	return ""
}

// pluginBinaryPath returns the absolute path to the plugin shared library.
// Mirrors the OS/arch suffixing logic in the Makefile.
func pluginBinaryPath(repoRoot string) string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return filepath.Join(repoRoot, "libclaude-code-plugin-darwin-arm64.dylib")
		}
		return filepath.Join(repoRoot, "libclaude-code-plugin-darwin-amd64.dylib")
	case "linux":
		if runtime.GOARCH == "arm64" {
			return filepath.Join(repoRoot, "libclaude-code-plugin-linux-arm64.so")
		}
		return filepath.Join(repoRoot, "libclaude-code-plugin-linux-amd64.so")
	}
	return ""
}

// requireFalcoEnv aborts the calling test with t.Skip() unless Falco and the
// plugin .dylib are both present. Use this at the top of every integration
// test so `go test ./...` on a bare laptop is still green.
func requireFalcoEnv(t *testing.T) (falcoBin, pluginPath, root string) {
	t.Helper()
	root = repoRoot(t)
	falcoBin = findFalcoBinary()
	if falcoBin == "" {
		t.Skip("falco binary not found; skipping Level 3 integration test")
	}
	pluginPath = pluginBinaryPath(root)
	if st, err := os.Stat(pluginPath); err != nil || st.Size() == 0 {
		t.Skipf("plugin binary missing at %s; run `make build` first", pluginPath)
	}
	return falcoBin, pluginPath, root
}

// stripMeta loads a fixture JSON file, removes its `_meta` block, and
// returns (jsonLine, meta). The returned line is suitable for direct append
// to a JSONL stream that Falco / the plugin will consume.
func stripMeta(t *testing.T, path string) (string, map[string]any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("parse fixture %s: %v", path, err)
	}
	var meta map[string]any
	if metaRaw, ok := generic[fixtureMetaKey]; ok {
		_ = json.Unmarshal(metaRaw, &meta)
		delete(generic, fixtureMetaKey)
	}
	cleaned, err := json.Marshal(generic)
	if err != nil {
		t.Fatalf("re-marshal fixture %s: %v", path, err)
	}
	return string(cleaned), meta
}

// writeFalcoConfig renders falco-test.yaml.tmpl with the supplied paths and
// writes it under `dir`. Returns the absolute path to the rendered config.
func writeFalcoConfig(t *testing.T, dir, repoRoot, pluginPath, eventsPath string) string {
	t.Helper()
	tmplPath := filepath.Join(repoRoot, "test", "integration", "falco-test.yaml.tmpl")
	tmpl, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("read falco-test.yaml.tmpl: %v", err)
	}
	rules := filepath.Join(repoRoot, "rules", "claude-code_rules.yaml")
	health := filepath.Join(repoRoot, "rules", "claude_code_health.yaml")
	rendered := string(tmpl)
	rendered = strings.ReplaceAll(rendered, "__PLUGIN_PATH__", pluginPath)
	rendered = strings.ReplaceAll(rendered, "__EVENTS_PATH__", eventsPath)
	rendered = strings.ReplaceAll(rendered, "__RULES_PATH__", rules)
	rendered = strings.ReplaceAll(rendered, "__HEALTH_PATH__", health)

	out := filepath.Join(dir, "falco-test.yaml")
	if err := os.WriteFile(out, []byte(rendered), 0o600); err != nil {
		t.Fatalf("write falco-test.yaml: %v", err)
	}
	return out
}

// FalcoSession represents a running falco process. Use startFalco() to
// launch and (*FalcoSession).Stop() to terminate. Stdout/stderr is buffered
// in-memory; AlertsContaining helps assert on output substrings.
type FalcoSession struct {
	cmd       *exec.Cmd
	stdoutBuf *bytes.Buffer
	stderrBuf *bytes.Buffer
	cancel    context.CancelFunc
}

// startFalco launches `falco -c <config> --disable-source syscall -U` and
// returns once the process is running. The caller must call Stop() to clean
// up. Falco binds no privileged resources on macOS, so no sudo is needed.
func startFalco(t *testing.T, falcoBin, configPath string) *FalcoSession {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, falcoBin,
		"-c", configPath,
		"--disable-source", "syscall",
		"-U",
	)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start falco: %v", err)
	}
	// Allow the plugin to load and start tailing.
	time.Sleep(startupGracePeriod)
	return &FalcoSession{cmd: cmd, stdoutBuf: stdout, stderrBuf: stderr, cancel: cancel}
}

// Stop terminates the falco process and waits for cleanup. Safe to call
// more than once.
func (s *FalcoSession) Stop() {
	if s == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	// best-effort wait
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Wait()
	}
}

// Stdout returns the accumulated falco stdout as a string. Safe to call
// while falco is still running (bytes.Buffer reads are independent of the
// running cmd).
func (s *FalcoSession) Stdout() string { return s.stdoutBuf.String() }

// Stderr returns the accumulated falco stderr as a string.
func (s *FalcoSession) Stderr() string { return s.stderrBuf.String() }

// AlertsContaining returns the subset of stdout lines that contain `needle`.
// Empty result means the alert was not observed.
func (s *FalcoSession) AlertsContaining(needle string) []string {
	out := []string{}
	scanner := bufio.NewScanner(strings.NewReader(s.Stdout()))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, needle) {
			out = append(out, line)
		}
	}
	return out
}

// appendJSONL appends `line` (without trailing newline) to the file at
// `path`, creating the file if necessary. The returned timestamp is the
// monotonic clock just before the append finishes — useful for latency.
func appendJSONL(t *testing.T, path, line string) time.Time {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatalf("open %s for append: %v", path, err)
	}
	defer f.Close()
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}
	if _, err := f.WriteString(line); err != nil {
		t.Fatalf("append to %s: %v", path, err)
	}
	if err := f.Sync(); err != nil {
		t.Fatalf("sync %s: %v", path, err)
	}
	return time.Now()
}

// runFalcoOnFixtures is the workhorse for TEST-006. It:
//  1. Creates a temp dir, writes the JSONL with all the supplied fixture
//     bodies AT ONCE (so start_at: beginning replays them all in one run).
//  2. Renders falco-test.yaml.
//  3. Launches falco, waits 2s, captures stdout, kills it.
//
// Returns the captured stdout for assertion.
func runFalcoOnFixtures(t *testing.T, falcoBin, pluginPath, repoRoot string, fixtureBodies []string) (stdout, stderr string) {
	t.Helper()
	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")
	body := strings.Join(fixtureBodies, "\n") + "\n"
	if err := os.WriteFile(eventsPath, []byte(body), 0o600); err != nil {
		t.Fatalf("write events.jsonl: %v", err)
	}
	cfgPath := writeFalcoConfig(t, dir, repoRoot, pluginPath, eventsPath)

	sess := startFalco(t, falcoBin, cfgPath)
	// give falco enough time to ingest all fixtures
	time.Sleep(alertGracePeriod)
	sess.Stop()
	return sess.Stdout(), sess.Stderr()
}

// countAlerts counts how many lines in `stdout` contain `needle`. Used to
// answer "did rule X fire at all".
func countAlerts(stdout, needle string) int {
	count := 0
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), needle) {
			count++
		}
	}
	return count
}

// fixturePath returns the absolute path to a fixture under
// test/fixtures/hook_events.
func fixturePath(repoRoot, rel string) string {
	return filepath.Join(repoRoot, "test", "fixtures", "hook_events", rel)
}

// errFmt formats a Falco failure with both stdout and stderr context, since
// rule-load errors land in stderr while alerts land in stdout.
func errFmt(stdout, stderr string) string {
	return fmt.Sprintf("--- falco stdout ---\n%s\n--- falco stderr ---\n%s", stdout, stderr)
}

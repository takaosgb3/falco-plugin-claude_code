package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestParseDuration covers OPS-005 duration validation.
func TestParseDuration(t *testing.T) {
	cases := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"15m", 15 * time.Minute, false},
		{"1h30m", 90 * time.Minute, false},
		{"30s", 30 * time.Second, false},
		{"0", 0, false},                         // explicit zero is allowed
		{"900", 0, true},                        // OPS-005: bare integer must error
		{"abc", 0, true},                        // gibberish
		{"", 0, true},                           // empty
		{"-1h", -time.Hour, false},              // negative duration is parseable; allowed
	}
	for _, c := range cases {
		got, err := parseDuration(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseDuration(%q): expected error, got %v", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseDuration(%q): unexpected error %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseDuration(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestExpandPath covers ~/ expansion and traversal rejection.
func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	if got, err := expandPath("~/foo/bar"); err != nil {
		t.Fatalf("unexpected: %v", err)
	} else if !strings.HasPrefix(got, home) {
		t.Errorf("expand: %s does not start with %s", got, home)
	}
	if _, err := expandPath("../etc/passwd"); err == nil {
		t.Errorf("expected traversal error for ../etc/passwd")
	}
	if got, err := expandPath("/abs/path"); err != nil || got != "/abs/path" {
		t.Errorf("absolute path roundtrip failed: got=%q err=%v", got, err)
	}
}

// TestReadLastJSONLine covers the tail-position core helper for various edge
// cases (single line, multi-line, trailing newline, empty file).
func TestReadLastJSONLine(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{"single", `{"a":1}`, `{"a":1}`, false},
		{"single-newline", "{\"a\":1}\n", `{"a":1}`, false},
		{"two-lines", "{\"a\":1}\n{\"b\":2}\n", `{"b":2}`, false},
		{"three-lines", "{\"a\":1}\n{\"b\":2}\n{\"c\":3}", `{"c":3}`, false},
		{"trailing-blank", "{\"a\":1}\n\n", `{"a":1}`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := filepath.Join(dir, c.name+".jsonl")
			if err := os.WriteFile(p, []byte(c.content), 0o600); err != nil {
				t.Fatalf("write: %v", err)
			}
			got, err := readLastJSONLine(p)
			if c.wantErr {
				if err == nil {
					t.Errorf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected: %v", err)
			}
			if got != c.want {
				t.Errorf("got=%q want=%q", got, c.want)
			}
		})
	}
}

// TestDoEnvSucceeds asserts OPS-001 PASSes when the plugin binary exists.
func TestDoEnvSucceeds(t *testing.T) {
	dir := t.TempDir()
	plug := filepath.Join(dir, "fake-plugin.dylib")
	if err := os.WriteFile(plug, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	gf := globalFlags{pluginPath: plug}
	r := doEnv(gf)
	if r.Status != statusPass || r.ExitCode != exitPass {
		t.Errorf("doEnv: got status=%s code=%d, want PASS/0; msg=%s", r.Status, r.ExitCode, r.Message)
	}
}

// TestDoEnvFailsOnMissingPlugin asserts OPS-001 FAILs when the plugin path is
// supplied but the file is absent.
func TestDoEnvFailsOnMissingPlugin(t *testing.T) {
	gf := globalFlags{pluginPath: "/nonexistent/path/plugin.dylib"}
	r := doEnv(gf)
	if r.Status != statusFail || r.ExitCode != exitFail {
		t.Errorf("doEnv: got status=%s code=%d, want FAIL/1", r.Status, r.ExitCode)
	}
}

// TestDoTailPositionFreshPass writes a synthetic events.jsonl with
// received_at=now and asserts PASS.
func TestDoTailPositionFreshPass(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "events.jsonl")
	rec := map[string]any{
		"schema_version": "claude_code_security_event/v1",
		"received_at":    time.Now().UTC().Format(time.RFC3339Nano),
		"event_name":     "_heartbeat_",
	}
	body, _ := json.Marshal(rec)
	if err := os.WriteFile(p, append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	r := doTailPosition(globalFlags{}, []string{p})
	if r.Status != statusPass {
		t.Errorf("status=%s want PASS; msg=%s", r.Status, r.Message)
	}
	if r.ExitCode != exitPass {
		t.Errorf("exit=%d want 0", r.ExitCode)
	}
}

// TestDoTailPositionStale exercises the STALE branch (>= max-age).
func TestDoTailPositionStale(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "events.jsonl")
	rec := map[string]any{
		"received_at": time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339Nano),
		"event_name":  "_heartbeat_",
	}
	body, _ := json.Marshal(rec)
	if err := os.WriteFile(p, append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	r := doTailPosition(globalFlags{}, []string{p, "--max-age", "15m"})
	if r.Status != statusStale {
		t.Errorf("status=%s want STALE; msg=%s", r.Status, r.Message)
	}
	if r.ExitCode != exitStale {
		t.Errorf("exit=%d want %d", r.ExitCode, exitStale)
	}
}

// TestDoTailPositionMissingFile checks OPS-005 exit=2 path.
func TestDoTailPositionMissingFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "absent.jsonl")
	r := doTailPosition(globalFlags{}, []string{p})
	if r.Status != statusSkip || r.ExitCode != exitSkip {
		t.Errorf("status=%s exit=%d want SKIP/2", r.Status, r.ExitCode)
	}
}

// TestDoTailPositionEmptyFile checks OPS-005 exit=1 path.
func TestDoTailPositionEmptyFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "empty.jsonl")
	if err := os.WriteFile(p, []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}
	r := doTailPosition(globalFlags{}, []string{p})
	if r.Status != statusFail || r.ExitCode != exitFail {
		t.Errorf("status=%s exit=%d want FAIL/1", r.Status, r.ExitCode)
	}
}

// TestDoTailPositionBareIntegerRejected covers OPS-005 unit requirement.
// Note: Go flag package requires flags BEFORE positional args.
func TestDoTailPositionBareIntegerRejected(t *testing.T) {
	r := doTailPosition(globalFlags{}, []string{"--max-age", "900", "/tmp/nonexistent.jsonl"})
	if r.Status != statusFail {
		t.Errorf("expected FAIL for bare integer max-age, got %s msg=%s", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "no unit") && !strings.Contains(r.Message, "max-age") {
		t.Errorf("error message should mention unit requirement, got: %s", r.Message)
	}
}

// TestDoSelfCheckPresent asserts OPS-004 PASSes when fixture+rule are present.
// This depends on the real repo layout, so it skips if go.mod isn't found
// upward.
func TestDoSelfCheckPresent(t *testing.T) {
	if _, ok := repoRoot(); !ok {
		t.Skip("repo root not found from cwd; skipping")
	}
	r := doSelfCheck(globalFlags{})
	if r.Status != statusPass {
		t.Errorf("status=%s want PASS; msg=%s", r.Status, r.Message)
	}
}

// TestDoVerifySignatureSkipsWithoutCosign verifies graceful SKIP when cosign
// is unavailable OR signature/cert files are missing.
func TestDoVerifySignatureSkipsWithoutCosign(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "fake-binary")
	if err := os.WriteFile(bin, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := doVerifySignature(globalFlags{}, []string{bin})
	// expected: SKIP because cosign or sig/cert missing
	if r.Status != statusSkip {
		t.Errorf("status=%s want SKIP; msg=%s", r.Status, r.Message)
	}
}

// TestRunUsage covers the help / unknown subcommand paths via the public
// `run` function (with captured writers).
func TestRunUsage(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"--help"}, &out, &errBuf)
	if code != exitPass {
		t.Errorf("--help exit=%d want 0", code)
	}
	if !strings.Contains(out.String(), "claude-code-doctor") {
		t.Errorf("usage missing program name; got: %s", out.String())
	}

	out.Reset()
	errBuf.Reset()
	code = run([]string{"bogus-subcommand"}, &out, &errBuf)
	if code != exitFail {
		t.Errorf("unknown subcommand exit=%d want 1", code)
	}
}

// TestRunEnvJSONOutput covers the JSON shape of `env --json`.
func TestRunEnvJSONOutput(t *testing.T) {
	var out, errBuf bytes.Buffer
	dir := t.TempDir()
	plug := filepath.Join(dir, "p.dylib")
	if err := os.WriteFile(plug, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := run([]string{"env", "--json", "--plugin", plug}, &out, &errBuf)
	if code != exitPass {
		t.Errorf("env --json exit=%d want 0; stderr=%s", code, errBuf.String())
	}
	var r Result
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatalf("env --json output is not JSON: %v\n%s", err, out.String())
	}
	if r.Subcommand != "env" || r.Status != statusPass {
		t.Errorf("unexpected JSON: %+v", r)
	}
}

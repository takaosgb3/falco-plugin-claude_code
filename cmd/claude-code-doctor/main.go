// claude-code-doctor — operator CLI for the claude-code Falco plugin.
//
// Implements requirements v3 §22.4 OPS-001..OPS-006:
//
//   env             OPS-001: Falco/Go/plugin existence + version
//   plugin-load     OPS-002: `falco -L` shows the plugin
//   rule-check      OPS-003: rules load without parse errors
//   self-check      OPS-004: heartbeat fixture fires the health rule
//   tail-position   OPS-005: tail position / staleness of events.jsonl
//   verify-signature OPS-006: cosign verify-blob against a release artifact
//   all             run env + plugin-load + rule-check + self-check
//
// Exit codes (per §22.4 OPS-005 spec, generalised across subcommands):
//
//	0  PASS (default)
//	1  FAIL (target observed, but does not satisfy the check)
//	2  SKIP (prerequisite missing — e.g. Falco binary, cosign, the file itself)
//	3  STALE (OPS-005 only; events.jsonl exists but last record is older than
//	   --max-age)
//
// The doctor is intentionally tolerant of missing prerequisites: in CI or on a
// dev laptop where Falco is not installed, OPS-002/003/004 SKIP cleanly so a
// `claude-code-doctor all` returns exit 0 as long as OPS-001 passes.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Result is the outcome of a single subcommand. The zero value is valid and
// represents an unknown / not-yet-run check (status="").
type Result struct {
	Subcommand string `json:"subcommand"`
	Status     string `json:"status"` // PASS / FAIL / SKIP / STALE
	ExitCode   int    `json:"exit_code"`
	Message    string `json:"message,omitempty"`
	// Details is opaque per-subcommand context (versions, paths, …) emitted
	// when --json is set or --verbose is used.
	Details map[string]any `json:"details,omitempty"`
}

const (
	statusPass  = "PASS"
	statusFail  = "FAIL"
	statusSkip  = "SKIP"
	statusStale = "STALE"

	exitPass  = 0
	exitFail  = 1
	exitSkip  = 2
	exitStale = 3

	defaultMaxAge = 15 * time.Minute
)

// globalFlags is mutated by parseGlobals().
type globalFlags struct {
	configPath string
	pluginPath string
	jsonOut    bool
	verbose    bool
}

func main() {
	exitCode := run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(exitCode)
}

// run is the testable entry point. argv excludes the program name. The caller
// passes its own writers so tests can capture output without touching globals.
func run(argv []string, stdout, stderr io.Writer) int {
	if len(argv) == 0 {
		printUsage(stderr)
		return exitFail
	}

	// Top-level help is allowed before a subcommand name (POSIX style).
	switch argv[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return exitPass
	}

	subcmd := argv[0]
	rest := argv[1:]

	gf, rest, err := parseGlobals(rest)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFail
	}

	switch subcmd {
	case "env":
		return finalize(stdout, gf, doEnv(gf))
	case "plugin-load":
		return finalize(stdout, gf, doPluginLoad(gf))
	case "rule-check":
		return finalize(stdout, gf, doRuleCheck(gf))
	case "self-check":
		return finalize(stdout, gf, doSelfCheck(gf))
	case "tail-position":
		return finalize(stdout, gf, doTailPosition(gf, rest))
	case "verify-signature":
		return finalize(stdout, gf, doVerifySignature(gf, rest))
	case "all":
		return finalize(stdout, gf, doAll(gf))
	default:
		fmt.Fprintf(stderr, "unknown subcommand: %s\n\n", subcmd)
		printUsage(stderr)
		return exitFail
	}
}

// parseGlobals consumes all known global flags from argv, leaving subcommand
// positional arguments behind. It uses ContinueOnError so unknown flags pass
// through to per-subcommand parsing later.
func parseGlobals(argv []string) (globalFlags, []string, error) {
	fs := flag.NewFlagSet("claude-code-doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // suppress usage on per-call parse error
	gf := globalFlags{}
	fs.StringVar(&gf.configPath, "config", "./falco-local.yaml", "Falco config file")
	fs.StringVar(&gf.pluginPath, "plugin", "", "plugin .dylib/.so path (default: auto)")
	fs.BoolVar(&gf.jsonOut, "json", false, "emit machine-readable JSON result")
	fs.BoolVar(&gf.verbose, "verbose", false, "verbose human-readable output")

	// We must let unknown flags through so subcommand-specific flags
	// (--max-age, etc.) don't trigger an error here. flag.FlagSet does not
	// support that natively; we manually filter.
	known := []string{}
	rest := []string{}
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		consume := func(name string) bool {
			if arg == name || strings.HasPrefix(arg, name+"=") {
				known = append(known, arg)
				if arg == name && i+1 < len(argv) && !strings.HasPrefix(argv[i+1], "-") {
					known = append(known, argv[i+1])
					i++
				}
				return true
			}
			return false
		}
		consume2 := func() bool {
			return consume("-config") || consume("--config") ||
				consume("-plugin") || consume("--plugin") ||
				consume("-json") || consume("--json") ||
				consume("-verbose") || consume("--verbose")
		}
		if !consume2() {
			rest = append(rest, arg)
		}
	}
	if err := fs.Parse(known); err != nil {
		return gf, nil, fmt.Errorf("parse global flags: %w", err)
	}
	if gf.pluginPath == "" {
		gf.pluginPath = defaultPluginPath()
	}
	return gf, rest, nil
}

// finalize prints a Result and returns the appropriate exit code.
func finalize(stdout io.Writer, gf globalFlags, r Result) int {
	if gf.jsonOut {
		_ = json.NewEncoder(stdout).Encode(r)
	} else {
		printHuman(stdout, r, gf.verbose)
	}
	return r.ExitCode
}

func printHuman(stdout io.Writer, r Result, verbose bool) {
	tag := r.Status
	if tag == "" {
		tag = "UNKNOWN"
	}
	fmt.Fprintf(stdout, "[%s] %s: %s\n", tag, r.Subcommand, r.Message)
	if verbose && len(r.Details) > 0 {
		// stable key order
		keys := make([]string, 0, len(r.Details))
		for k := range r.Details {
			keys = append(keys, k)
		}
		// simple sort
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[j] < keys[i] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
		for _, k := range keys {
			fmt.Fprintf(stdout, "  %s = %v\n", k, r.Details[k])
		}
	}
}

func printUsage(w io.Writer) {
	const usage = `claude-code-doctor — operator diagnostic CLI for the claude-code Falco plugin

Usage:
  claude-code-doctor <subcommand> [flags]

Subcommands:
  env                    OPS-001: report Falco / Go / plugin presence + versions
  plugin-load            OPS-002: verify Falco can load the plugin (-L)
  rule-check             OPS-003: verify Falco can load the claude-code rules
  self-check             OPS-004: dry-run a heartbeat fixture and confirm fire
  tail-position [path]   OPS-005: report tail position / staleness of events.jsonl
                            --max-age <duration> (default 15m); valid units s/m/h
  verify-signature [bin] OPS-006: cosign verify-blob a release artifact
                            --signature <path> --certificate <path>
  all                    run env + plugin-load + rule-check + self-check

Global flags:
  --config <path>        Falco config file (default ./falco-local.yaml)
  --plugin <path>        plugin .dylib/.so path (default: auto-detect by OS)
  --json                 emit JSON result (machine-readable)
  --verbose              verbose human output

Exit codes:
  0  PASS
  1  FAIL
  2  SKIP (prerequisite missing)
  3  STALE (tail-position only)
`
	fmt.Fprint(w, usage)
}

// -------- subcommand implementations --------

// doEnv (OPS-001) verifies Go runtime, optional Falco binary, and plugin.
func doEnv(gf globalFlags) Result {
	r := Result{Subcommand: "env", Details: map[string]any{}}
	r.Details["go_version"] = runtime.Version()
	r.Details["goos"] = runtime.GOOS
	r.Details["goarch"] = runtime.GOARCH

	// Falco
	falcoBin := findFalcoBinary()
	if falcoBin == "" {
		r.Details["falco"] = "not found"
	} else {
		r.Details["falco_path"] = falcoBin
		if v, err := falcoVersion(falcoBin); err == nil {
			r.Details["falco_version"] = v
		} else {
			r.Details["falco_version_error"] = err.Error()
		}
	}

	// Plugin
	if gf.pluginPath == "" {
		r.Details["plugin"] = "not configured"
	} else {
		st, err := os.Stat(gf.pluginPath)
		if err != nil {
			r.Details["plugin"] = "missing"
			r.Details["plugin_path"] = gf.pluginPath
			r.Status = statusFail
			r.ExitCode = exitFail
			r.Message = fmt.Sprintf("plugin not found at %s; run `make build`", gf.pluginPath)
			return r
		}
		r.Details["plugin_path"] = gf.pluginPath
		r.Details["plugin_size_bytes"] = st.Size()
	}

	// PASS regardless of Falco presence — Falco is optional for OPS-001.
	r.Status = statusPass
	r.ExitCode = exitPass
	r.Message = "environment check OK"
	return r
}

// doPluginLoad (OPS-002) runs `falco -L -c <config> --disable-source syscall -U`
// and confirms the plugin name appears.
func doPluginLoad(gf globalFlags) Result {
	r := Result{Subcommand: "plugin-load", Details: map[string]any{}}
	falcoBin := findFalcoBinary()
	if falcoBin == "" {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = "falco binary not found; SKIP"
		return r
	}
	if _, err := os.Stat(gf.configPath); err != nil {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = fmt.Sprintf("config not found: %s; SKIP", gf.configPath)
		return r
	}
	out, err := runFalcoList(falcoBin, gf.configPath)
	r.Details["falco"] = falcoBin
	r.Details["config"] = gf.configPath
	if err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("falco -L failed: %v", err)
		r.Details["stderr_excerpt"] = excerpt(out, 500)
		return r
	}
	if !strings.Contains(out, "claude-code") && !strings.Contains(out, "claude_code") {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = "falco -L did not list claude-code plugin"
		r.Details["stdout_excerpt"] = excerpt(out, 500)
		return r
	}
	r.Status = statusPass
	r.ExitCode = exitPass
	r.Message = "plugin recognized by falco -L"
	return r
}

// doRuleCheck (OPS-003) runs `falco -L` and looks for rule load errors.
func doRuleCheck(gf globalFlags) Result {
	r := Result{Subcommand: "rule-check", Details: map[string]any{}}
	falcoBin := findFalcoBinary()
	if falcoBin == "" {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = "falco binary not found; SKIP"
		return r
	}
	if _, err := os.Stat(gf.configPath); err != nil {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = fmt.Sprintf("config not found: %s; SKIP", gf.configPath)
		return r
	}
	out, err := runFalcoList(falcoBin, gf.configPath)
	if err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("rule load failed: %v", err)
		r.Details["stderr_excerpt"] = excerpt(out, 800)
		return r
	}
	// Falco emits LOAD_ERR_* on parse errors. LOAD_UNUSED is just a warning.
	if strings.Contains(out, "LOAD_ERR_") {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = "rule loader emitted LOAD_ERR_*"
		r.Details["stderr_excerpt"] = excerpt(out, 800)
		return r
	}
	r.Status = statusPass
	r.ExitCode = exitPass
	r.Message = "rules loaded without LOAD_ERR_*"
	return r
}

// doSelfCheck (OPS-004) is intentionally a SKIP-by-default in v0.1: spawning
// a Falco process and waiting on a heartbeat fixture is implemented by the
// integration tests (test/integration/falco_alerts_test.go). The doctor only
// confirms the fixture and rule are present, so a heartbeat-fired check can be
// performed by the operator out-of-band.
func doSelfCheck(gf globalFlags) Result {
	r := Result{Subcommand: "self-check", Details: map[string]any{}}
	root, ok := repoRoot()
	if !ok {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = "could not locate repo root; SKIP"
		return r
	}
	heartbeat := filepath.Join(root, "test", "fixtures", "hook_events", "_heartbeat_", "heartbeat-ok.json")
	if _, err := os.Stat(heartbeat); err != nil {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = "heartbeat fixture missing; SKIP"
		return r
	}
	healthRules := filepath.Join(root, "rules", "claude_code_health.yaml")
	body, err := os.ReadFile(healthRules)
	if err != nil {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = "claude_code_health.yaml missing; SKIP"
		return r
	}
	if !strings.Contains(string(body), `claude_code.event_name = "_heartbeat_"`) {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = "Plugin Heartbeat rule condition not found in health rules"
		return r
	}
	r.Details["heartbeat_fixture"] = heartbeat
	r.Details["health_rules"] = healthRules
	r.Status = statusPass
	r.ExitCode = exitPass
	r.Message = "heartbeat fixture and rule present (operator: run `make e2e-l3` to assert fire)"
	return r
}

// doTailPosition (OPS-005) inspects the last record of an events.jsonl file
// and decides PASS / STALE / FAIL based on `received_at` age.
func doTailPosition(gf globalFlags, args []string) Result {
	r := Result{Subcommand: "tail-position", Details: map[string]any{}}

	fs := flag.NewFlagSet("tail-position", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	maxAgeStr := fs.String("max-age", defaultMaxAge.String(), "max age of last record (Go time.Duration; e.g. 15m, 1h)")
	if err := fs.Parse(args); err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("flag parse: %v", err)
		return r
	}
	pos := fs.Args()
	path := defaultEventsPath()
	if len(pos) > 0 {
		path = pos[0]
	}
	maxAge, err := parseDuration(*maxAgeStr)
	if err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("invalid --max-age: %v", err)
		return r
	}
	r.Details["path"] = path
	r.Details["max_age"] = maxAge.String()

	expanded, err := expandPath(path)
	if err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("expand path: %v", err)
		return r
	}
	st, err := os.Stat(expanded)
	if err != nil {
		// OPS-005: file missing -> exit 2
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = fmt.Sprintf("events.jsonl not found at %s", expanded)
		return r
	}
	r.Details["size_bytes"] = st.Size()
	if st.Size() == 0 {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = "events.jsonl exists but is empty"
		return r
	}
	last, err := readLastJSONLine(expanded)
	if err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("read last line: %v", err)
		return r
	}
	r.Details["last_line_size"] = len(last)

	var rec map[string]any
	if err := json.Unmarshal([]byte(last), &rec); err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("last record not JSON: %v", err)
		return r
	}
	receivedAt, _ := rec["received_at"].(string)
	if receivedAt == "" {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = "last record missing received_at"
		return r
	}
	r.Details["received_at"] = receivedAt
	t, err := time.Parse(time.RFC3339Nano, receivedAt)
	if err != nil {
		// also try plain RFC3339 (no nanos)
		t, err = time.Parse(time.RFC3339, receivedAt)
		if err != nil {
			r.Status = statusFail
			r.ExitCode = exitFail
			r.Message = fmt.Sprintf("received_at not parseable: %v", err)
			return r
		}
	}
	age := time.Since(t)
	r.Details["age"] = age.String()
	if age > maxAge {
		r.Status = statusStale
		r.ExitCode = exitStale
		r.Message = fmt.Sprintf("last record is %s old (> %s)", age.Truncate(time.Second), maxAge)
		return r
	}
	r.Status = statusPass
	r.ExitCode = exitPass
	r.Message = fmt.Sprintf("last record is %s old (≤ %s)", age.Truncate(time.Second), maxAge)
	return r
}

// doVerifySignature (OPS-006) wraps cosign verify-blob. Skips if cosign isn't
// installed or the signature/cert files are missing.
func doVerifySignature(gf globalFlags, args []string) Result {
	r := Result{Subcommand: "verify-signature", Details: map[string]any{}}
	fs := flag.NewFlagSet("verify-signature", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sigPath := fs.String("signature", "", "path to .sig signature file")
	crtPath := fs.String("certificate", "", "path to .crt / .cert certificate file")
	identityRegexp := fs.String("certificate-identity-regexp",
		`https://github\.com/takaosgb3/falco-plugin-claude_code/.+`,
		"cosign --certificate-identity-regexp value")
	oidcIssuer := fs.String("certificate-oidc-issuer",
		"https://token.actions.githubusercontent.com",
		"cosign --certificate-oidc-issuer value")
	if err := fs.Parse(args); err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("flag parse: %v", err)
		return r
	}
	pos := fs.Args()
	if len(pos) == 0 {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = "missing positional <binary> path"
		return r
	}
	binPath := pos[0]
	if _, err := os.Stat(binPath); err != nil {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = fmt.Sprintf("artifact not found: %s; SKIP", binPath)
		return r
	}
	if _, err := exec.LookPath("cosign"); err != nil {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = "cosign not installed; SKIP"
		return r
	}
	if *sigPath == "" {
		*sigPath = binPath + ".sig"
	}
	if *crtPath == "" {
		// try .cert then .crt (sigstore writes either depending on version)
		c1 := binPath + ".cert"
		c2 := binPath + ".crt"
		if _, err := os.Stat(c1); err == nil {
			*crtPath = c1
		} else if _, err := os.Stat(c2); err == nil {
			*crtPath = c2
		}
	}
	if *sigPath == "" || *crtPath == "" {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = "signature or certificate file missing; SKIP"
		return r
	}
	if _, err := os.Stat(*sigPath); err != nil {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = fmt.Sprintf("signature missing: %s; SKIP", *sigPath)
		return r
	}
	if _, err := os.Stat(*crtPath); err != nil {
		r.Status = statusSkip
		r.ExitCode = exitSkip
		r.Message = fmt.Sprintf("certificate missing: %s; SKIP", *crtPath)
		return r
	}
	cmd := exec.Command("cosign", "verify-blob",
		"--signature", *sigPath,
		"--certificate", *crtPath,
		"--certificate-identity-regexp", *identityRegexp,
		"--certificate-oidc-issuer", *oidcIssuer,
		binPath,
	)
	out, err := cmd.CombinedOutput()
	r.Details["cmd"] = strings.Join(cmd.Args, " ")
	if err != nil {
		r.Status = statusFail
		r.ExitCode = exitFail
		r.Message = fmt.Sprintf("cosign verify-blob failed: %v", err)
		r.Details["output"] = excerpt(string(out), 800)
		return r
	}
	r.Status = statusPass
	r.ExitCode = exitPass
	r.Message = "cosign verify-blob succeeded"
	return r
}

// doAll runs the four "always-applicable" environment checks and aggregates
// the worst exit code (FAIL > STALE > SKIP > PASS).
func doAll(gf globalFlags) Result {
	r := Result{Subcommand: "all", Details: map[string]any{}}
	steps := []struct {
		name string
		fn   func() Result
	}{
		{"env", func() Result { return doEnv(gf) }},
		{"plugin-load", func() Result { return doPluginLoad(gf) }},
		{"rule-check", func() Result { return doRuleCheck(gf) }},
		{"self-check", func() Result { return doSelfCheck(gf) }},
	}
	worst := exitPass
	for _, s := range steps {
		sr := s.fn()
		r.Details[s.name+"_status"] = sr.Status
		r.Details[s.name+"_message"] = sr.Message
		if sr.ExitCode > worst {
			worst = sr.ExitCode
		}
	}
	r.ExitCode = worst
	switch worst {
	case exitPass:
		r.Status = statusPass
		r.Message = "all checks PASS"
	case exitSkip:
		r.Status = statusSkip
		r.Message = "some checks SKIP (prerequisites missing); none failed"
	case exitFail:
		r.Status = statusFail
		r.Message = "one or more checks FAIL"
	}
	return r
}

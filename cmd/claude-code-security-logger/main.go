// claude-code-security-logger
//
// A small command-line hook handler for Claude Code (NOT a Falco plugin).
//
// Reads a single Claude Code Hook JSON object from stdin, normalizes it into a
// `claude_code_security_event/v1` event (requirements v3 §10.1), redacts
// secrets per §17.1, and appends the event as a single JSONL line to a target
// file (default: ~/.claude/security/events.jsonl).
//
// Phase 1 status: SKELETON ONLY.
//   - HL-001 stdin JSON read         : implemented (tolerant: empty stdin -> exit 0)
//   - HL-002 JSON parse              : implemented (malformed -> stderr + exit 0)
//   - HL-003 schema fields           : partial (schema_version / received_at / logger_version / host / user)
//   - HL-004 common fields           : passthrough (preserved if present)
//   - HL-005 type-specific normalize : TODO (Phase 2)
//   - HL-006 redaction               : TODO (Phase 2; §17.1 patterns)
//   - HL-007 JSONL append            : implemented (atomic append, file mode 0600)
//   - HL-008 permissions             : implemented (dir 0700 / file 0600)
//   - HL-009 stdout quiet            : implemented (no stdout output)
//   - HL-010 fail-open exit policy   : implemented (errors logged to stderr; exit 0)
//   - HL-011 path safety             : partial (~/ expansion + traversal check; symlink race TODO)
//   - HL-012 size limit              : implemented (raw 64KB cap; field-level cap TODO)
//   - HL-013 latency target          : informational (latency_ms emitted; benchmarks TODO)
//   - HL-014 testability             : structure ready (Phase 2 fixtures + tests)
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// LoggerVersion is bumped whenever the schema or redaction logic changes.
const LoggerVersion = "0.1.0"

// SchemaVersion identifies the normalized event schema (requirements v3 §10.3).
const SchemaVersion = "claude_code_security_event/v1"

// Caps (§27.1)
const (
	maxRawBytes        = 64 * 1024 // §27.1 Max event size
	maxEvidenceBytes   = 2 * 1024  // §27.1 Evidence max
	defaultDirMode     = 0o700     // §27.1
	defaultFileMode    = 0o600     // §27.1
	defaultLogPathTmpl = "~/.claude/security/events.jsonl"
)

func main() {
	if err := run(); err != nil {
		// HL-010: fail-open. Log to stderr, never block Claude Code.
		fmt.Fprintf(os.Stderr, "[claude-code-security-logger] %v\n", err)
	}
	os.Exit(0)
}

func run() error {
	outPath := flag.String("out", defaultLogPathTmpl, "JSONL output path")
	flag.Parse()

	resolvedPath, err := resolvePath(*outPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// HL-001: read stdin (bounded)
	rawInput, err := readStdinCapped(maxRawBytes)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if len(rawInput) == 0 {
		// No hook payload -> nothing to log. Exit silently.
		return nil
	}

	receivedAt := time.Now()

	// HL-002: parse JSON
	var hookInput map[string]interface{}
	if err := json.Unmarshal(rawInput, &hookInput); err != nil {
		// Malformed input: emit a malformed marker event (TODO Phase 2: counters)
		_ = appendMalformed(resolvedPath, rawInput, receivedAt, err)
		return nil
	}

	// HL-003 / HL-004 / HL-005: build normalized event (skeleton; TODO Phase 2)
	event := buildNormalizedEvent(hookInput, rawInput, receivedAt)

	// HL-006: redact secrets in evidence/command/file_path/url (TODO Phase 2)
	redactEventInPlace(event)

	// HL-007 / HL-008: atomic JSONL append with mode 0600
	if err := appendJSONL(resolvedPath, event); err != nil {
		return fmt.Errorf("append JSONL: %w", err)
	}

	return nil
}

// resolvePath expands a leading "~/" and rejects ".." segments (HL-011).
// Symlink race detection is a TODO for Phase 2.
func resolvePath(p string) (string, error) {
	if strings.Contains(p, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", p)
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, p[2:])
	}
	return p, nil
}

// readStdinCapped reads up to maxBytes from stdin. Bytes beyond the cap are
// dropped silently (HL-012; we record event_size_bytes from the cap).
func readStdinCapped(maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(os.Stdin, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		data = data[:maxBytes]
	}
	return data, nil
}

// buildNormalizedEvent creates a §10.1-shaped event. This Phase 1 skeleton
// preserves common fields straight from the hook input and computes the
// schema/sha256/host/user fields. Phase 2 adds type-specific normalization.
func buildNormalizedEvent(hookInput map[string]interface{}, rawInput []byte, receivedAt time.Time) map[string]interface{} {
	event := make(map[string]interface{})

	// Logger-controlled fields (HL-003)
	event["schema_version"] = SchemaVersion
	event["received_at"] = receivedAt.Format(time.RFC3339Nano)
	event["logger_version"] = LoggerVersion

	if host, err := os.Hostname(); err == nil {
		event["host"] = host
	}
	if u, err := user.Current(); err == nil {
		event["user"] = u.Username
	}

	// Common Claude Code Hook fields (HL-004): passthrough
	for _, k := range []string{
		"session_id", "transcript_path", "cwd",
		"permission_mode", "hook_event_name",
	} {
		if v, ok := hookInput[k]; ok {
			event[k] = v
		}
	}
	// Normalize hook_event_name -> event_name as well, so downstream parser
	// can read either key (parser already handles both; §10.2 / parser.go).
	if v, ok := hookInput["hook_event_name"]; ok {
		event["event_name"] = v
	}

	// Tool-related passthrough (HL-005 will refine in Phase 2)
	for _, k := range []string{
		"tool_name", "tool_use_id", "tool_input", "tool_response",
		"command", "file_path", "url", "source",
		"mcp_server_name", "mcp_tool_name", "mcp_scope",
		"permission_destination", "permission_behavior",
		"event", "permission_suggestions",
	} {
		if v, ok := hookInput[k]; ok {
			event[k] = v
		}
	}

	// SHA-256 of raw input (FP-011 / ES-007)
	sum := sha256.Sum256(rawInput)
	event["raw_event_sha256"] = hex.EncodeToString(sum[:])
	event["event_size_bytes"] = len(rawInput)

	// Detector fields (set by hook logger heuristics in Phase 2; defaults here)
	if _, ok := event["risk_type"]; !ok {
		event["risk_type"] = "none"
	}
	if _, ok := event["risk_score"]; !ok {
		event["risk_score"] = 0
	}
	if _, ok := event["severity"]; !ok {
		event["severity"] = "info"
	}
	if _, ok := event["redaction_status"]; !ok {
		event["redaction_status"] = "none"
	}
	event["dropped"] = false
	event["latency_ms"] = time.Since(receivedAt).Milliseconds()

	return event
}

// redactEventInPlace masks secrets in evidence/command/file_path/url.
// Phase 1: no-op. Phase 2 implements the §17.1 minimum redaction set
// (AWS / GCP / Slack / GitHub PAT / OAuth / JWT / RSA / .env / Cookie).
func redactEventInPlace(event map[string]interface{}) {
	// TODO (Phase 2): implement §17.1 redaction patterns
	if v, ok := event["evidence"].(string); ok && len(v) > maxEvidenceBytes {
		event["evidence"] = v[:maxEvidenceBytes]
		event["redaction_status"] = "truncated"
	}
}

// appendJSONL atomically appends a single JSON object as a JSONL line.
// Creates the parent directory with mode 0700 and the file with mode 0600.
func appendJSONL(path string, event map[string]interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, defaultDirMode); err != nil {
		return err
	}

	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, defaultFileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(line); err != nil {
		return err
	}
	return nil
}

// appendMalformed appends a minimal record so that operators can see that an
// invalid hook payload arrived, without blocking Claude Code (HL-002).
func appendMalformed(path string, rawInput []byte, receivedAt time.Time, parseErr error) error {
	sum := sha256.Sum256(rawInput)
	record := map[string]interface{}{
		"schema_version":   SchemaVersion,
		"received_at":      receivedAt.Format(time.RFC3339Nano),
		"logger_version":   LoggerVersion,
		"hook_event_name":  "MalformedInput",
		"event_name":       "MalformedInput",
		"raw_event_sha256": hex.EncodeToString(sum[:]),
		"event_size_bytes": len(rawInput),
		"risk_type":        "malformed_input",
		"risk_score":       0,
		"severity":         "notice",
		"redaction_status": "none",
		"evidence":         truncate(parseErr.Error(), maxEvidenceBytes),
		"dropped":          false,
	}
	return appendJSONL(path, record)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

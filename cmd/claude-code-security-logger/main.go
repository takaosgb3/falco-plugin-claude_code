// claude-code-security-logger
//
// A small command-line hook handler for Claude Code (NOT a Falco plugin).
//
// Reads a single Claude Code Hook JSON object from stdin, normalizes it into a
// `claude_code_security_event/v1` event (requirements v3 §10.1), redacts
// secrets per §17.1, and appends the event as a single JSONL line to a target
// file (default: ~/.claude/security/events.jsonl).
//
// Phase 2 status:
//   - HL-001 stdin JSON read         : implemented
//   - HL-002 JSON parse              : implemented (malformed -> stderr + malformed JSONL line)
//   - HL-003 schema fields           : implemented
//   - HL-004 common fields           : passthrough preserved
//   - HL-005 type-specific normalize : implemented (per-event-type promotion of tool_input fields)
//   - HL-006 redaction               : implemented (pkg/parser §17.1 patterns)
//   - HL-007 JSONL append            : implemented (atomic append, file mode 0600)
//   - HL-008 permissions             : implemented (dir 0700 / file 0600)
//   - HL-009 stdout quiet            : implemented
//   - HL-010 fail-open exit policy   : implemented (errors logged to stderr; exit 0)
//   - HL-011 path safety             : implemented (~/ expansion + traversal check + symlink/inode check)
//   - HL-012 size limit              : implemented (raw 64KB cap + per-field 2KB cap on string fields)
//   - HL-013 latency target          : informational (latency_ms emitted; benchmarks TODO Phase 4)
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
	"syscall"
	"time"

	"github.com/takaosgb3/falco-plugin-claude_code/pkg/parser"
)

// LoggerVersion is bumped whenever the schema or redaction logic changes.
const LoggerVersion = "0.1.0"

// SchemaVersion identifies the normalized event schema (requirements v3 §10.3).
const SchemaVersion = "claude_code_security_event/v1"

// Caps (§27.1, HL-012)
const (
	maxRawBytes        = 64 * 1024 // §27.1 Max event size
	maxFieldBytes      = 2 * 1024  // §27.1 / HL-012 individual string field cap (2 KiB)
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
		// Malformed input: emit a malformed marker event (§14.1 P-007 counter).
		_ = appendMalformed(resolvedPath, rawInput, receivedAt, err)
		return nil
	}

	// HL-003 / HL-004 / HL-005: build normalized event
	event := buildNormalizedEvent(hookInput, rawInput, receivedAt)

	// HL-006: redact secrets in evidence/command/file_path/url/raw_excerpt.
	// HL-012: per-field 2KB cap.
	redactEventInPlace(event)

	// HL-007 / HL-008: atomic JSONL append with mode 0600.
	if err := appendJSONL(resolvedPath, event); err != nil {
		return fmt.Errorf("append JSONL: %w", err)
	}

	return nil
}

// resolvePath expands a leading "~/" and rejects ".." segments (HL-011).
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

// buildNormalizedEvent creates a §10.1-shaped event. HL-005 implements
// type-specific normalization: certain hook event names promote sub-fields
// from `tool_input` / `tool_response` into top-level normalized fields so
// downstream rules can write `claude_code.command` rather than digging into
// raw JSON.
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
	if v, ok := hookInput["hook_event_name"]; ok {
		event["event_name"] = v
	}

	// Tool-related passthrough at the top level (HL-005 may overwrite below)
	for _, k := range []string{
		"tool_name", "tool_use_id", "command", "file_path", "url", "source",
		"mcp_server_name", "mcp_tool_name", "mcp_scope",
		"permission_destination", "permission_behavior",
		"event", "permission_suggestions",
	} {
		if v, ok := hookInput[k]; ok {
			event[k] = v
		}
	}

	// HL-005 type-specific normalization. tool_input / tool_response are
	// nested objects whose shape depends on tool_name. We promote known
	// subfields to the canonical top-level claude_code field names so
	// rules don't have to introspect raw JSON.
	normalizeToolFields(hookInput, event)
	normalizeURLDomain(event)
	normalizeFileEvent(hookInput, event)
	normalizeMCP(event)

	// SHA-256 of raw input (FP-011 / ES-007)
	sum := sha256.Sum256(rawInput)
	event["raw_event_sha256"] = hex.EncodeToString(sum[:])
	event["event_size_bytes"] = len(rawInput)

	// command_hash (§14.1 P-005). Computed pre-redaction so the hash
	// identifies the *original* command; redaction afterwards leaves the
	// `command` field masked but the hash stable for SIEM correlation.
	if cmd, ok := event["command"].(string); ok && cmd != "" {
		h := sha256.Sum256([]byte(cmd))
		event["command_hash"] = hex.EncodeToString(h[:])[:16]
	}

	// Detector fields (set by hook logger heuristics in Phase 2 — currently
	// emit defaults; the parser-side detector classifies if these stay at
	// defaults).
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

// normalizeToolFields promotes well-known subfields of tool_input.
//
// PreToolUse / PostToolUse messages carry `tool_input: { ... tool-specific ... }`
// (Claude Code Hook spec). For example:
//
//	Bash       → { "command": "..." }
//	Read/Edit  → { "file_path": "..." }
//	WebFetch   → { "url": "...", "prompt": "..." }
//	Mcp        → { "mcp_server": "...", "mcp_tool": "...", ... }
//
// PostToolUse additionally carries `tool_response` and `duration_ms`.
func normalizeToolFields(hookInput map[string]interface{}, event map[string]interface{}) {
	if ti, ok := hookInput["tool_input"].(map[string]interface{}); ok {
		promote(ti, event, "command", "command")
		promote(ti, event, "file_path", "file_path")
		promote(ti, event, "url", "url")
		promote(ti, event, "mcp_server", "mcp_server_name")
		promote(ti, event, "mcp_tool", "mcp_tool_name")
		promote(ti, event, "mcp_scope", "mcp_scope")
	}
	if tr, ok := hookInput["tool_response"].(map[string]interface{}); ok {
		// We do not promote response bodies (could leak); only meta.
		if v, ok := tr["duration_ms"].(float64); ok {
			if v < 0 {
				event["duration_ms"] = uint64(0)
			} else {
				event["duration_ms"] = uint64(v)
			}
		}
		if v, ok := tr["failure"].(bool); ok && v {
			event["failure_count"] = uint64(1)
		}
	}
	if v, ok := hookInput["duration_ms"].(float64); ok {
		if v < 0 {
			event["duration_ms"] = uint64(0)
		} else {
			event["duration_ms"] = uint64(v)
		}
	}
	if v, ok := hookInput["tool_count"].(float64); ok {
		if v < 0 {
			event["tool_count"] = uint64(0)
		} else {
			event["tool_count"] = uint64(v)
		}
	}
}

// promote copies src[srcKey] (string) into dst[dstKey] if not already set.
func promote(src, dst map[string]interface{}, srcKey, dstKey string) {
	if _, exists := dst[dstKey]; exists {
		return
	}
	if v, ok := src[srcKey].(string); ok && v != "" {
		dst[dstKey] = v
	}
}

// normalizeURLDomain extracts the host part of url and writes claude_code.domain.
func normalizeURLDomain(event map[string]interface{}) {
	u, ok := event["url"].(string)
	if !ok || u == "" {
		return
	}
	// Cheap domain extraction (no net/url dependency to keep binary small):
	//   scheme://host/path  →  host
	rest := u
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+3:]
	}
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		rest = rest[:i]
	}
	// strip user-info / port
	if i := strings.LastIndex(rest, "@"); i >= 0 {
		rest = rest[i+1:]
	}
	if i := strings.Index(rest, ":"); i >= 0 {
		rest = rest[:i]
	}
	if rest != "" {
		event["domain"] = strings.ToLower(rest)
	}
}

// normalizeFileEvent maps Claude Code FileChanged hook payloads onto
// claude_code.file_event ∈ {add, change, unlink, rename, ...}.
func normalizeFileEvent(hookInput, event map[string]interface{}) {
	// Direct passthrough first.
	if v, ok := hookInput["file_event"]; ok {
		event["file_event"] = v
		return
	}
	if hookInput["hook_event_name"] != "FileChanged" {
		return
	}
	if v, ok := hookInput["event"].(string); ok {
		event["file_event"] = v
	}
}

// normalizeMCP makes sure mcp_* fields are lowercase / canonical strings.
func normalizeMCP(event map[string]interface{}) {
	if v, ok := event["mcp_scope"].(string); ok {
		event["mcp_scope"] = strings.ToLower(v)
	}
}

// redactEventInPlace masks secrets in user-controllable string fields and
// applies the per-field 2 KB cap (HL-006 + HL-012).
//
// `redaction_status` is upgraded if anything changed.
func redactEventInPlace(event map[string]interface{}) {
	anyRedacted := false
	anyTruncated := false

	// Fields most likely to carry user / external secrets.
	stringFields := []string{
		"command", "evidence", "url", "file_path",
		"raw_excerpt", "transcript_path",
	}
	for _, k := range stringFields {
		v, ok := event[k].(string)
		if !ok {
			continue
		}
		capped, truncated := parser.CapField(v, maxFieldBytes)
		if truncated {
			anyTruncated = true
		}
		redacted, didRedact := parser.Redact(capped)
		if didRedact {
			anyRedacted = true
		}
		event[k] = redacted
	}

	// Set / upgrade redaction_status. We only upgrade — never silently
	// downgrade an upstream value.
	current, _ := event["redaction_status"].(string)
	if current == "" || current == "none" {
		event["redaction_status"] = parser.RedactionStatus(anyRedacted, anyTruncated, false)
	}
}

// appendJSONL atomically appends a single JSON object as a JSONL line.
// Creates the parent directory with mode 0700 and the file with mode 0600.
//
// HL-011 symlink race detection: after open we lstat the path and compare
// st_dev / st_ino with f.Stat(). If they differ, the path was a symlink at
// open time — bail out without writing. (We never follow symlinks for the
// JSONL output; the operator is expected to point us at a real file.)
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

	if err := verifyNoSymlinkRace(path, f); err != nil {
		return err
	}

	if _, err := f.Write(line); err != nil {
		return err
	}
	return nil
}

// verifyNoSymlinkRace reports an error if `path` (lstat) does not refer to
// the same inode as the open fd.
func verifyNoSymlinkRace(path string, f *os.File) error {
	pathStat, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("lstat path: %w", err)
	}
	// If lstat says the path itself is a symlink, refuse.
	if pathStat.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write through symlink: %s", path)
	}
	fdStat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("fstat fd: %w", err)
	}
	pSys, ok1 := pathStat.Sys().(*syscall.Stat_t)
	fSys, ok2 := fdStat.Sys().(*syscall.Stat_t)
	if !ok1 || !ok2 {
		// platform without stat_t: treat as best-effort.
		return nil
	}
	if pSys.Dev != fSys.Dev || pSys.Ino != fSys.Ino {
		return fmt.Errorf("symlink/inode race detected on %s", path)
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
		"session_id":       "malformed",
		"hook_event_name":  "MalformedInput",
		"event_name":       "MalformedInput",
		"raw_event_sha256": hex.EncodeToString(sum[:]),
		"event_size_bytes": len(rawInput),
		"risk_type":        "malformed_input",
		"risk_score":       0,
		"severity":         "notice",
		"redaction_status": "none",
		"evidence":         truncate(parseErr.Error(), maxFieldBytes),
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

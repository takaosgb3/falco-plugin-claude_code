package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validLine returns a complete schema-valid JSONL line with the given
// `extras` merged into the JSON body. It always includes the required
// fields per §10.3 (schema_version, received_at, session_id, hook_event_name).
func validLine(extras string) string {
	base := `"schema_version":"claude_code_security_event/v1","received_at":"2026-04-26T12:00:00Z","session_id":"s","hook_event_name":"PreToolUse"`
	if extras == "" {
		return "{" + base + "}"
	}
	return "{" + base + "," + extras + "}"
}

// --- Schema validation (§14.1 P-001 / P-002 / P-003) ---

func TestSchemaValidation_MissingSchemaVersion(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := `{"received_at":"2026-04-26T12:00:00Z","session_id":"s","hook_event_name":"PreToolUse"}`
	_, err := p.Parse(line)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema_version")
	c := p.CountersSnapshot()
	assert.Equal(t, uint64(1), c.Malformed)
}

func TestSchemaValidation_UnsupportedSchema(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := `{"schema_version":"some_other/v1","received_at":"2026-04-26T12:00:00Z","session_id":"s","hook_event_name":"PreToolUse"}`
	_, err := p.Parse(line)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestSchemaValidation_AcceptsMinorBump(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := `{"schema_version":"claude_code_security_event/v1.1","received_at":"2026-04-26T12:00:00Z","session_id":"s","hook_event_name":"PreToolUse"}`
	_, err := p.Parse(line)
	require.NoError(t, err, "minor schema bumps must be forward-compatible")
}

func TestSchemaValidation_MissingSessionID(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := `{"schema_version":"claude_code_security_event/v1","received_at":"2026-04-26T12:00:00Z","hook_event_name":"PreToolUse"}`
	_, err := p.Parse(line)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session_id")
}

func TestSchemaValidation_MissingEventName(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := `{"schema_version":"claude_code_security_event/v1","received_at":"2026-04-26T12:00:00Z","session_id":"s"}`
	_, err := p.Parse(line)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event_name")
}

func TestSchemaValidation_MalformedJSON(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	_, err := p.Parse(`{"schema_version":...broken...}`)
	require.Error(t, err)
	c := p.CountersSnapshot()
	assert.Equal(t, uint64(1), c.Malformed)
}

func TestSchemaValidation_UnknownFieldsIgnored(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"some_future_field":"hi","another":42`)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.NotNil(t, entry)
}

// --- Required-field type-conversion (§10.2) ---

// TestAllDomainFields_Types asserts that every claude_code.* field that the
// schema defines as bool / number gets the correct Go-typed value.
func TestAllDomainFields_Types(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`
"logger_version":"0.1.0","host":"h","user":"u",
"transcript_path":"/t","cwd":"/c","permission_mode":"default",
"source":"src","tool_name":"Bash","tool_use_id":"tu1",
"command":"echo hi","command_hash":"abc",
"file_path":"/p","file_event":"change","url":"http://x","domain":"x",
"mcp_server_name":"m","mcp_tool_name":"t","mcp_scope":"user",
"permission_destination":"userSettings","permission_behavior":"allow",
"risk_type":"upstream_set","risk_score":42,"severity":"info","evidence":"e",
"redaction_status":"none","raw_event_sha256":"deadbeef",
"event_size_bytes":100,"duration_ms":15,"latency_ms":3,
"tool_count":7,"failure_count":0,"dropped":true,"raw_excerpt":""`)

	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Equal(t, uint64(42), entry.RiskScore)
	assert.Equal(t, uint64(100), entry.EventSizeBytes)
	assert.Equal(t, uint64(15), entry.DurationMs)
	assert.Equal(t, uint64(3), entry.LatencyMs)
	assert.Equal(t, uint64(7), entry.ToolCount)
	assert.Equal(t, "true", entry.Dropped, "bool→string stringification per §10.2")
	assert.Equal(t, "Bash", entry.ToolName)
	assert.Equal(t, "abc", entry.CommandHash, "logger-supplied command_hash preserved")
}

func TestAllDomainFields_NegativeNumbersClipped(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"risk_score":-1,"latency_ms":-100,"duration_ms":-5,"tool_count":-7,"failure_count":-1,"event_size_bytes":-2`)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), entry.RiskScore)
	assert.Equal(t, uint64(0), entry.LatencyMs)
	assert.Equal(t, uint64(0), entry.DurationMs)
	assert.Equal(t, uint64(0), entry.ToolCount)
	assert.Equal(t, uint64(0), entry.FailureCount)
	// EventSizeBytes is auto-filled by parser when 0; check it's at least the
	// length of the input line (parser-side default).
	assert.GreaterOrEqual(t, entry.EventSizeBytes, uint64(len(line)/2))
}

// --- command_hash auto-computed when absent ---

func TestParser_AutoCommandHash(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"command":"echo hello"`)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.NotEmpty(t, entry.CommandHash, "command_hash must be auto-computed")
	assert.Len(t, entry.CommandHash, 16, "short hash is 16 hex chars (P-005)")
}

// TestParser_RawEventSHA256AutoComputed asserts P-006 / ES-007.
func TestParser_RawEventSHA256AutoComputed(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"command":"foo"`)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Len(t, entry.RawEventSHA256, 64, "full sha256 hex is 64 chars")
}

// --- Redaction defense-in-depth ---

func TestParser_RedactsCommand(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"command":"export AWS_KEY=AKIAIOSFODNN7EXAMPLE && echo done"`)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.NotContains(t, entry.Command, "AKIAIOSFODNN7EXAMPLE",
		"parser must redact AWS keys even if logger missed them")
	assert.Contains(t, []string{"redacted", "truncated"}, entry.RedactionStatus,
		"redaction_status must be upgraded after parser-side redaction")
	c := p.CountersSnapshot()
	assert.GreaterOrEqual(t, c.Redacted, uint64(1))
}

func TestParser_RedactsEvidence(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"evidence":"diff: +AKIAIOSFODNN7EXAMPLE"`)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.NotContains(t, entry.Evidence, "AKIAIOSFODNN7EXAMPLE")
}

// TestParser_PreservesUpstreamRedacted asserts we don't overwrite an upstream
// redaction_status="redacted" with "none".
func TestParser_PreservesUpstreamRedacted(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"redaction_status":"redacted","command":"echo ok"`)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Equal(t, "redacted", entry.RedactionStatus)
}

// TestParser_AppliesFieldCap verifies HL-012 (per-field cap) at parser side.
func TestParser_AppliesFieldCap(t *testing.T) {
	huge := strings.Repeat("a", 5000)
	p := New(Config{LogFormat: "json", MaxFieldLength: 2048})
	line := validLine(`"command":"` + huge + `"`)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(entry.Command), 2048,
		"oversized fields must be capped at MaxFieldLength")
	assert.Equal(t, "truncated", entry.RedactionStatus)
}

// TestParser_DetectorRunsOnlyWhenUpstreamMissing — D-006.
func TestParser_DetectorRunsWhenUpstreamMissing(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"hook_event_name":"PermissionRequest","permission_mode":"default"`)
	// validLine() sets hook_event_name to PreToolUse; override:
	line = strings.Replace(line,
		`"hook_event_name":"PreToolUse"`,
		`"hook_event_name":"PermissionRequest"`, 1)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Equal(t, "permission_update", entry.RiskType,
		"detector should classify when logger left risk_type empty")
}

func TestParser_DetectorSkipsWhenUpstreamSet(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := validLine(`"hook_event_name":"PermissionRequest","risk_type":"custom_upstream","risk_score":99`)
	line = strings.Replace(line,
		`"hook_event_name":"PreToolUse"`,
		`"hook_event_name":"PermissionRequest"`, 1)
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Equal(t, "custom_upstream", entry.RiskType,
		"upstream classification must win")
}

// TestParser_HeadersInitialized — P004 / P-005.
func TestParser_HeadersAlwaysInitialized(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	entry, err := p.Parse(validLine(""))
	require.NoError(t, err)
	require.NotNil(t, entry.Headers, "Headers must be non-nil even when JSON omits headers")
}

package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Common Field Tests ---
// These tests verify domain-independent behavior.
// claude_code domain-specific tests are added by plugin-parser SKILL in Phase 2.

func TestParseEmptyLine(t *testing.T) {
	p := New(Config{LogFormat: "json"})

	_, err := p.Parse("")
	assert.Error(t, err, "Empty line should return error")
}

func TestHeadersInitialized(t *testing.T) {
	p := New(Config{LogFormat: "json"})

	line := `{"received_at":"2026-04-26T12:34:56+09:00","schema_version":"claude_code_security_event/v1","session_id":"abc","hook_event_name":"PreToolUse","risk_type":"none","risk_score":0}`
	entry, err := p.Parse(line)

	require.NoError(t, err)
	assert.NotNil(t, entry.Headers, "Headers map must be initialized (P004)")
	assert.NotEmpty(t, entry.Raw, "Raw must be set")
	assert.False(t, entry.Timestamp.IsZero(), "Timestamp must be parsed")
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"RFC3339", "2026-02-27T10:00:00Z", true},
		{"RFC3339_offset", "2026-02-27T10:00:00+09:00", true},
		{"ISO8601", "2026-02-27T10:00:00", true},
		{"datetime", "2026-02-27 10:00:00", true},
		{"invalid", "not-a-date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := parseTimestamp(tt.input)
			if tt.valid {
				assert.False(t, ts.IsZero(), "Should parse valid timestamp: %s", tt.input)
			}
		})
	}
}

// --- Security Pattern Tests ---

func TestDetectSQLInjection(t *testing.T) {
	detector := NewSimpleSecurityDetector(10 * 1024)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"single quote OR", "' OR '1'='1", true},
		{"union select", "1 UNION SELECT * FROM users", true},
		{"sleep function", "1; SLEEP(5)", true},
		{"normal text", "hello world page=1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.DetectSQLInjection(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectXSS(t *testing.T) {
	detector := NewSimpleSecurityDetector(10 * 1024)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"script tag", "<script>alert(1)</script>", true},
		{"javascript URI", "javascript:alert(1)", true},
		{"event handler", "<img onerror=alert(1)>", true},
		{"normal text", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.DetectXSS(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectPathTraversal(t *testing.T) {
	detector := NewSimpleSecurityDetector(10 * 1024)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"dot-dot-slash", "../../../etc/passwd", true},
		{"encoded traversal", "..%2f..%2f..%2fetc%2fpasswd", true},
		{"etc passwd", "/etc/passwd", true},
		{"normal path", "/files/document.pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.DetectPathTraversal(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectCommandInjection(t *testing.T) {
	detector := NewSimpleSecurityDetector(10 * 1024)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"semicolon ls", ";ls", true},
		{"subshell", "$(cat /etc/passwd)", true},
		{"pipe", "|cat /etc/passwd", true},
		{"normal text", "search hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.DetectCommandInjection(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectSuspiciousAgent(t *testing.T) {
	detector := NewSimpleSecurityDetector(10 * 1024)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"sqlmap", "sqlmap/1.5.2#stable", true},
		{"nikto", "nikto/2.1.6", true},
		{"nmap", "Nmap Scripting Engine", true},
		{"normal browser", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.DetectSuspiciousAgent(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- URL Encoding Tests ---
//
// NOTE: The template's TestURLDecoding test calls DetectSecurityThreat() directly,
// but the detector intentionally does NOT URL-decode (R5-001: decoding happens
// once in the parser's detectSecurityPatterns()). To match the actual contract,
// this test exercises the detector with already-decoded input. This is a known
// dev-kit test-template smell to be reported via /dev-kit-feedback.

func TestURLDecoding(t *testing.T) {
	detector := NewSimpleSecurityDetector(10 * 1024)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"decoded SQLi quote OR", "' OR 1=1", true},
		{"decoded XSS script tag", "<script>alert(1)</script>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := detector.DetectSecurityThreat(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- JSON Parse Tests (claude_code schema basic) ---

func TestParseJSONClaudeCodeBasic(t *testing.T) {
	p := New(Config{LogFormat: "json"})

	line := `{
		"schema_version": "claude_code_security_event/v1",
		"received_at": "2026-04-26T12:34:56.789+09:00",
		"logger_version": "0.1.0",
		"host": "dev-macbook.local",
		"user": "alice",
		"session_id": "abc123",
		"hook_event_name": "PreToolUse",
		"tool_name": "Bash",
		"tool_use_id": "toolu_01ABC",
		"command": "echo hello",
		"risk_type": "none",
		"risk_score": 0,
		"severity": "info",
		"dropped": false
	}`
	entry, err := p.Parse(line)

	require.NoError(t, err)
	assert.Equal(t, "claude_code_security_event/v1", entry.SchemaVersion)
	assert.Equal(t, "alice", entry.User)
	assert.Equal(t, "abc123", entry.SessionID)
	assert.Equal(t, "PreToolUse", entry.EventName, "hook_event_name should map to EventName")
	assert.Equal(t, "Bash", entry.ToolName)
	assert.Equal(t, "false", entry.Dropped, "bool dropped should be stringified to \"false\"")
	assert.NotNil(t, entry.Headers, "Headers must be initialized (P004)")
}

func TestParseJSONDroppedTrueBool(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := `{"schema_version":"claude_code_security_event/v1","received_at":"2026-04-26T12:00:00Z","session_id":"s","hook_event_name":"PreToolUse","risk_type":"x","risk_score":1,"dropped":true}`
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Equal(t, "true", entry.Dropped)
}

func TestParseJSONRiskScoreUint(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := `{"schema_version":"claude_code_security_event/v1","received_at":"2026-04-26T12:00:00Z","session_id":"s","hook_event_name":"PreToolUse","risk_score":90}`
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Equal(t, uint64(90), entry.RiskScore)
}

func TestParseJSONNegativeNumberClippedToZero(t *testing.T) {
	p := New(Config{LogFormat: "json"})
	line := `{"schema_version":"claude_code_security_event/v1","received_at":"2026-04-26T12:00:00Z","session_id":"s","hook_event_name":"PreToolUse","latency_ms":-5}`
	entry, err := p.Parse(line)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), entry.LatencyMs, "negative numbers must be clipped to 0 (§10.2 type-conversion guidance)")
}

// --- Auto-Detect Parse Tests ---

func TestParseAutoJSON(t *testing.T) {
	p := New(Config{LogFormat: "auto"})

	line := `{"schema_version":"claude_code_security_event/v1","received_at":"2026-02-27T10:00:00Z","session_id":"s","hook_event_name":"PreToolUse"}`
	entry, err := p.Parse(line)

	require.NoError(t, err)
	assert.False(t, entry.Timestamp.IsZero(), "Auto-detect should parse JSON starting with '{'")
}

func TestParseAutoWhitespaceJSON(t *testing.T) {
	p := New(Config{LogFormat: "auto"})

	line := `  {"schema_version":"claude_code_security_event/v1","received_at":"2026-02-27T10:00:00Z","session_id":"s","hook_event_name":"PreToolUse"}`
	entry, err := p.Parse(line)

	require.NoError(t, err)
	assert.False(t, entry.Timestamp.IsZero(), "Auto-detect should handle whitespace-prefixed JSON")
}

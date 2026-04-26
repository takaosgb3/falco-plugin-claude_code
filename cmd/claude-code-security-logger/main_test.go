package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- HL-005 normalization ---

func TestNormalize_BashCommandPromoted(t *testing.T) {
	hookInput := map[string]interface{}{
		"hook_event_name": "PreToolUse",
		"session_id":      "s1",
		"tool_name":       "Bash",
		"tool_input": map[string]interface{}{
			"command": "git status",
		},
	}
	event := buildNormalizedEvent(hookInput, []byte(`{}`), time.Now())
	assert.Equal(t, "git status", event["command"], "Bash tool_input.command must be promoted to top level")
}

func TestNormalize_WebFetchURLPromoted(t *testing.T) {
	hookInput := map[string]interface{}{
		"hook_event_name": "PreToolUse",
		"session_id":      "s2",
		"tool_name":       "WebFetch",
		"tool_input": map[string]interface{}{
			"url":    "https://docs.example.com:8080/api?x=1",
			"prompt": "summarize",
		},
	}
	event := buildNormalizedEvent(hookInput, []byte(`{}`), time.Now())
	assert.Equal(t, "https://docs.example.com:8080/api?x=1", event["url"])
	assert.Equal(t, "docs.example.com", event["domain"], "domain extracted from url")
}

func TestNormalize_ReadFilePathPromoted(t *testing.T) {
	hookInput := map[string]interface{}{
		"hook_event_name": "PreToolUse",
		"session_id":      "s3",
		"tool_name":       "Read",
		"tool_input": map[string]interface{}{
			"file_path": "/etc/passwd",
		},
	}
	event := buildNormalizedEvent(hookInput, []byte(`{}`), time.Now())
	assert.Equal(t, "/etc/passwd", event["file_path"])
}

func TestNormalize_DurationMsTypedAsUint(t *testing.T) {
	hookInput := map[string]interface{}{
		"hook_event_name": "PostToolUse",
		"session_id":      "s4",
		"tool_name":       "Bash",
		"tool_response": map[string]interface{}{
			"duration_ms": float64(1234),
		},
	}
	event := buildNormalizedEvent(hookInput, []byte(`{}`), time.Now())
	v, ok := event["duration_ms"].(uint64)
	require.True(t, ok, "duration_ms must be uint64 not float64")
	assert.Equal(t, uint64(1234), v)
}

func TestNormalize_NegativeDurationClipped(t *testing.T) {
	hookInput := map[string]interface{}{
		"hook_event_name": "PostToolUse",
		"session_id":      "s5",
		"tool_name":       "Bash",
		"duration_ms":     float64(-50),
	}
	event := buildNormalizedEvent(hookInput, []byte(`{}`), time.Now())
	v, ok := event["duration_ms"].(uint64)
	require.True(t, ok)
	assert.Equal(t, uint64(0), v, "negative number must be clipped to 0 (§10.2)")
}

func TestNormalize_FileChangedEventPromoted(t *testing.T) {
	hookInput := map[string]interface{}{
		"hook_event_name": "FileChanged",
		"session_id":      "s6",
		"event":           "change",
		"file_path":       "/repo/.claude/settings.json",
	}
	event := buildNormalizedEvent(hookInput, []byte(`{}`), time.Now())
	assert.Equal(t, "change", event["file_event"])
	assert.Equal(t, "/repo/.claude/settings.json", event["file_path"])
}

func TestNormalize_CommandHashComputed(t *testing.T) {
	hookInput := map[string]interface{}{
		"hook_event_name": "PreToolUse",
		"session_id":      "s7",
		"tool_name":       "Bash",
		"tool_input": map[string]interface{}{
			"command": "echo hello world",
		},
	}
	event := buildNormalizedEvent(hookInput, []byte(`{}`), time.Now())
	hash, ok := event["command_hash"].(string)
	require.True(t, ok)
	assert.Len(t, hash, 16, "command_hash short form is 16 hex chars (P-005)")
}

// --- HL-006 redaction ---

func TestRedact_AWSKeyMasked(t *testing.T) {
	event := map[string]interface{}{
		"command":          "aws s3 cp x s3://b --profile dev #AKIAIOSFODNN7EXAMPLE",
		"redaction_status": "none",
	}
	redactEventInPlace(event)
	cmd := event["command"].(string)
	assert.NotContains(t, cmd, "AKIAIOSFODNN7EXAMPLE")
	assert.Equal(t, "redacted", event["redaction_status"])
}

func TestRedact_PrivateKeyMasked(t *testing.T) {
	pem := "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA\n-----END RSA PRIVATE KEY-----"
	event := map[string]interface{}{
		"evidence": pem,
	}
	redactEventInPlace(event)
	out := event["evidence"].(string)
	assert.NotContains(t, out, "MIIEowIBAAKCAQEA")
	assert.Equal(t, "redacted", event["redaction_status"])
}

func TestRedact_BenignUntouched(t *testing.T) {
	event := map[string]interface{}{
		"command":          "git status",
		"redaction_status": "none",
	}
	redactEventInPlace(event)
	assert.Equal(t, "git status", event["command"])
	assert.Equal(t, "none", event["redaction_status"])
}

func TestRedact_PreservesUpstream(t *testing.T) {
	event := map[string]interface{}{
		"command":          "echo ok",
		"redaction_status": "redacted",
	}
	redactEventInPlace(event)
	// Should not be downgraded to "none" even though no new redaction happened.
	assert.Equal(t, "redacted", event["redaction_status"])
}

// --- HL-012 per-field cap ---

func TestRedact_FieldCapTriggersTruncated(t *testing.T) {
	huge := strings.Repeat("a", 5*1024)
	event := map[string]interface{}{
		"evidence":         huge,
		"redaction_status": "none",
	}
	redactEventInPlace(event)
	got := event["evidence"].(string)
	assert.LessOrEqual(t, len(got), maxFieldBytes,
		"evidence must be capped to %d bytes (HL-012)", maxFieldBytes)
	assert.Equal(t, "truncated", event["redaction_status"])
}

// --- HL-001 / HL-002 ---

func TestResolvePath_RejectsTraversal(t *testing.T) {
	_, err := resolvePath("/tmp/../etc/passwd")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal")
}

func TestResolvePath_TildeExpansion(t *testing.T) {
	got, err := resolvePath("~/.claude/security/events.jsonl")
	require.NoError(t, err)
	assert.False(t, strings.HasPrefix(got, "~/"), "~/ must be expanded: %s", got)
}

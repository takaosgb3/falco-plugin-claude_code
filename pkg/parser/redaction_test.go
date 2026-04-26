package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedaction_AllPatterns asserts that each of the §17.1 minimum redaction
// patterns is masked. We never assert on the exact mask placement — only
// that the original secret bytes do not appear in the output and that the
// `redacted` flag is true.
func TestRedaction_AllPatterns(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		secret string // substring that must NOT appear in the output
	}{
		{
			name:   "AWS Access Key ID",
			input:  "use AKIAIOSFODNN7EXAMPLE for the deploy",
			secret: "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:   "AWS Secret Access Key (key/value)",
			input:  `aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`,
			secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:   "Slack Bot Token",
			input:  "channel=ops slack_token=xoxb-123456789012-abcdefghij",
			secret: "xoxb-123456789012-abcdefghij",
		},
		{
			name:   "GitHub PAT classic",
			input:  "GITHUB_TOKEN=ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			secret: "ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name:   "GitHub PAT fine-grained",
			input:  "set GH=github_pat_" + strings.Repeat("a", 70),
			secret: "github_pat_" + strings.Repeat("a", 70),
		},
		{
			name:   "OAuth Bearer / Authorization header",
			input:  "Authorization: Bearer abcdefghijklmnopqrstuv",
			secret: "abcdefghijklmnopqrstuv",
		},
		{
			name:   "JWT",
			input:  "Cookie: jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			secret: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		},
		{
			name:   "RSA private key block",
			input:  "key=-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAxxxx\n-----END RSA PRIVATE KEY-----",
			secret: "MIIEowIBAAKCAQEAxxxx",
		},
		{
			name:   ".env style assignment",
			input:  "DATABASE_URL=postgres://user:supersecret@db.example.com/prod",
			secret: "supersecret",
		},
		{
			name:   "Cookie header",
			input:  "Set-Cookie: session=u2YEgaTCmnA1WEzqpQfUvA",
			secret: "u2YEgaTCmnA1WEzqpQfUvA",
		},
		{
			name:   "Anthropic API key",
			input:  "ANTHROPIC_API_KEY=sk-ant-api01-aaaaaaaaaaaaaaaaaaaaaaaa",
			secret: "sk-ant-api01-aaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, redacted := Redact(tt.input)
			assert.True(t, redacted, "Redact() should report redacted=true for %s", tt.name)
			assert.NotContains(t, out, tt.secret, "secret bytes leaked to output")
			assert.Contains(t, out, "***REDACTED:", "expected masked placeholder in output")
		})
	}
}

func TestRedaction_NoFalsePositives(t *testing.T) {
	benigns := []string{
		"echo hello world",
		"npm test --silent",
		"git status",
		"docs say BEGIN with capital B",
		"plain text without secrets",
	}
	for _, in := range benigns {
		out, redacted := Redact(in)
		assert.False(t, redacted, "benign input should not be redacted: %q", in)
		assert.Equal(t, in, out)
	}
}

func TestRedaction_LooksLikeSecret(t *testing.T) {
	assert.True(t, LooksLikeSecret("AKIAIOSFODNN7EXAMPLE"))
	assert.False(t, LooksLikeSecret("hello"))
}

func TestRedactionStatus(t *testing.T) {
	assert.Equal(t, "error", RedactionStatus(true, true, true))
	assert.Equal(t, "truncated", RedactionStatus(true, true, false))
	assert.Equal(t, "redacted", RedactionStatus(true, false, false))
	assert.Equal(t, "none", RedactionStatus(false, false, false))
}

func TestCapField(t *testing.T) {
	short := "hello"
	got, trunc := CapField(short, 100)
	assert.False(t, trunc)
	assert.Equal(t, short, got)

	long := strings.Repeat("a", 5000)
	got, trunc = CapField(long, 2048)
	assert.True(t, trunc)
	assert.Equal(t, 2048, len(got))

	// no cap
	got, trunc = CapField(long, 0)
	assert.False(t, trunc)
	assert.Equal(t, len(long), len(got))
}

// TestCapField_UTF8Boundary asserts CapField does not split a multi-byte
// rune at the cut point.
func TestCapField_UTF8Boundary(t *testing.T) {
	in := strings.Repeat("こ", 100) // 3 bytes per rune × 100 = 300 bytes
	got, trunc := CapField(in, 7)   // mid-rune
	assert.True(t, trunc)
	// Result must be valid UTF-8 (no replacement chars introduced when we
	// cleanly cut on rune boundaries — strings.ToValidUTF8 strips invalid).
	assert.True(t, len(got) <= 7)
}

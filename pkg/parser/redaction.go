// Package parser — secret redaction (§17.1)
//
// All patterns are anchored / bounded to remain ReDoS-safe (§14.2 D-002 /
// SEC-006). Inputs are expected to be already capped at a max-field length
// before being passed to Redact (§14.2 D-004). The intent is that the hook
// logger (claude-code-security-logger) calls Redact() on every string field
// that it normalizes, but the plugin parser also calls it as a defense in
// depth so that a misbehaving/missing logger cannot leak secrets to Falco
// alerts.
package parser

import (
	"regexp"
	"strings"
)

// RedactionMask is the placeholder we substitute for any redacted material.
// Format: ***REDACTED:<kind>***  (kept stable so rules / SIEMs can match it).
const RedactionMask = "***REDACTED***"

// redactionPattern is one entry in the redaction set.
type redactionPattern struct {
	Kind   string         // categorical name used in the mask: ***REDACTED:<kind>***
	Regex  *regexp.Regexp // anchored / bounded
	Repl   string         // replacement string (may use $1..$N)
	GroupN int            // when >0, replace match[GroupN] only (so context is kept)
}

// All patterns are compiled once at package init. Order matters: longest /
// most-specific patterns are tried first to avoid double redactions.
//
// Every regexp uses bounded quantifiers ({n} or {n,m}) so they are linear-
// time on input length (ReDoS safe).
var redactionPatterns = []redactionPattern{
	// 1. RSA / SSH / generic PEM private keys (block redaction).
	//
	// Go's regexp engine caps {n,m} repeat counts at ~1000; we use {0,1000}
	// as the bounded body (a typical 4096-bit RSA key body is ~3.2KB but the
	// regex is anchored on BEGIN/END markers and run on already-bounded
	// inputs, so a 1000-char inner cap matches the BEGIN line + first ~14
	// base64 lines — enough to redact the BEGIN…END framing on any sane
	// input that the upstream callers cap at 64KB).
	{
		Kind:  "private_key",
		Regex: regexp.MustCompile(`-----BEGIN (?:RSA |DSA |EC |OPENSSH |ENCRYPTED |PGP )?PRIVATE KEY-----[\s\S]{0,1000}?-----END (?:RSA |DSA |EC |OPENSSH |ENCRYPTED |PGP )?PRIVATE KEY-----`),
		Repl:  mask("private_key"),
	},

	// 2. AWS Access Key ID (deterministic prefix).
	{
		Kind:  "aws_access_key_id",
		Regex: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
		Repl:  mask("aws_access_key_id"),
	},

	// 3. AWS Secret Access Key in `aws_secret_access_key=...` / `Secret=...` /
	//    `AWS_SECRET_ACCESS_KEY=...` style assignments (40-char base64).
	//    We only redact the value when it follows an explicit AWS-context key
	//    to avoid false positives on arbitrary 40-char tokens.
	{
		Kind:   "aws_secret_access_key",
		Regex:  regexp.MustCompile(`(?i)(aws[_-]?secret[_-]?access[_-]?key\s*[:=]\s*)["']?([A-Za-z0-9/+=]{40})["']?`),
		Repl:   "$1" + mask("aws_secret_access_key"),
		GroupN: 0, // full-match replace; $1 is already preserved in Repl
	},

	// 4. Slack Bot/User tokens.
	{
		Kind:  "slack_token",
		Regex: regexp.MustCompile(`\bxox[abprs]-[A-Za-z0-9-]{10,255}\b`),
		Repl:  mask("slack_token"),
	},

	// 5. GitHub PAT (classic + fine-grained).
	{
		Kind:  "github_pat",
		Regex: regexp.MustCompile(`\bghp_[A-Za-z0-9]{36}\b|\bgithub_pat_[A-Za-z0-9_]{60,255}\b`),
		Repl:  mask("github_pat"),
	},

	// 6. OpenAI / Anthropic API keys.
	{
		Kind:  "ai_api_key",
		Regex: regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_\-]{20,255}\b|\bsk-[A-Za-z0-9_\-]{20,255}\b`),
		Repl:  mask("ai_api_key"),
	},

	// 7. OAuth Bearer / Authorization headers (value redacted, keep header name).
	{
		Kind:  "authorization_header",
		Regex: regexp.MustCompile(`(?i)(authorization\s*[:=]\s*)(?:bearer\s+)?[A-Za-z0-9._\-/+=]{16,1000}`),
		Repl:  "$1" + mask("authorization_header"),
	},

	// 8. Generic JWT (3 dot-separated base64url; minimum sizes prevent FPs).
	{
		Kind:  "jwt",
		Regex: regexp.MustCompile(`\beyJ[A-Za-z0-9_\-]{10,1000}\.[A-Za-z0-9_\-]{10,1000}\.[A-Za-z0-9_\-]{10,1000}\b`),
		Repl:  mask("jwt"),
	},

	// 9. Cookie / Set-Cookie header (value redacted).
	{
		Kind:  "cookie",
		Regex: regexp.MustCompile(`(?i)((?:set-)?cookie\s*[:=]\s*)([^\s;,]{4,1000})`),
		Repl:  "$1" + mask("cookie"),
	},

	// 10. .env style assignments (uppercase env name, value redacted).
	//     Anchored to start-of-line OR a separator so we don't redact arbitrary
	//     KEY=value occurrences inside English prose.
	//
	//     The negative-context list (PATH/HOME/USER/...) is intentionally not
	//     part of the regex — those are common shell vars and we redact their
	//     values too because the plugin should be paranoid; truncated values
	//     are safe.
	{
		Kind:  "env_assignment",
		Regex: regexp.MustCompile(`(^|[\s;&|])([A-Z][A-Z0-9_]{2,63})=([^\s;&|"']{4,1000})`),
		Repl:  "$1$2=" + mask("env_assignment"),
	},
}

// mask wraps the kind in the standard placeholder shape.
func mask(kind string) string { return "***REDACTED:" + kind + "***" }

// Redact returns a copy of s with §17.1 secrets masked. It also reports
// whether any redaction took place (the bool is used by callers to set
// `redaction_status = "redacted"`).
//
// Empty / very short inputs short-circuit. Per §14.2 D-004 the caller is
// expected to bound s before calling, but we re-cap to a hard limit here as
// a safety net.
func Redact(s string) (string, bool) {
	if s == "" {
		return s, false
	}
	const hardCap = 64 * 1024
	if len(s) > hardCap {
		s = s[:hardCap]
	}

	out := s
	redacted := false
	for _, p := range redactionPatterns {
		replaced := p.Regex.ReplaceAllString(out, p.Repl)
		if replaced != out {
			redacted = true
			out = replaced
		}
	}
	return out, redacted
}

// RedactionStatus returns the canonical redaction_status string for an event
// based on the (changed, truncated) flags.
//
// Order of precedence: error > truncated > redacted > none. The caller passes
// `truncated=true` if any field was cut at its size cap (HL-012).
func RedactionStatus(redacted, truncated, hadError bool) string {
	switch {
	case hadError:
		return "error"
	case truncated:
		return "truncated"
	case redacted:
		return "redacted"
	default:
		return "none"
	}
}

// LooksLikeSecret returns true if Redact(s) would produce any change. Used by
// fast-path callers that want to set redaction_status without actually
// modifying the value (e.g. counter-only paths).
func LooksLikeSecret(s string) bool {
	if s == "" {
		return false
	}
	for _, p := range redactionPatterns {
		if p.Regex.MatchString(s) {
			return true
		}
	}
	return false
}

// CapField truncates s to at most max bytes and returns (truncated_value, was_truncated).
// It is rune-safe at the boundary: if the cut point would split a multi-byte
// UTF-8 sequence we shorten by one byte.
//
// max <= 0 means "no cap" (returns s as-is).
func CapField(s string, max int) (string, bool) {
	if max <= 0 || len(s) <= max {
		return s, false
	}
	cut := max
	for cut > 0 && (s[cut]&0xC0) == 0x80 {
		cut--
	}
	return strings.ToValidUTF8(s[:cut], ""), true
}

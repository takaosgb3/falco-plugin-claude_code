// Package parser provides log parsing functionality for the claude-code Falco plugin.
//
// Input format: JSONL produced by claude-code-security-logger.
// Schema: claude_code_security_event/v1 (see requirements v3 §10.1).
package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// SecurityThreatType represents the type of security threat detected.
type SecurityThreatType int

const (
	NoThreat SecurityThreatType = iota
	SQLInjection
	XSSAttempt
	PathTraversal
	CommandInjection
	SuspiciousUserAgent
)

func (t SecurityThreatType) String() string {
	switch t {
	case SQLInjection:
		return "sqli"
	case XSSAttempt:
		return "xss"
	case PathTraversal:
		return "path_traversal"
	case CommandInjection:
		return "cmd_injection"
	case SuspiciousUserAgent:
		return "suspicious_agent"
	default:
		return "none"
	}
}

// supportedSchemaPrefix is the schema family the parser understands. Minor
// versions (v1, v1.1, ...) are forward-compatible; major version bumps emit
// a malformed counter increment but are not fatal (P-002).
const supportedSchemaPrefix = "claude_code_security_event/v"

// LogEntry represents a parsed claude_code security event line.
// Structure: Common fields (fixed) + claude_code domain fields (§10.2 of requirements v3).
//
// Note: domain detection categories (T-001..T-018) are exposed via RiskType / RiskScore
// fields (set by the hook logger / detector), not via SecurityThreatType. The
// SecurityThreatType field is retained from the dev-kit common skeleton for
// completeness but is not used for claude_code rules in v0.1.
type LogEntry struct {
	// --- Common fields (fixed) ---
	Timestamp      time.Time
	SecurityThreat SecurityThreatType
	Headers        map[string]string
	Raw            string

	// --- Domain-specific fields (claude_code.*; §10.2) ---
	// ${DOMAIN_FIELDS_STRUCT}
	SchemaVersion         string
	ReceivedAt            string
	LoggerVersion         string
	Host                  string
	User                  string
	SessionID             string
	TranscriptPath        string
	Cwd                   string
	PermissionMode        string
	EventName             string
	Source                string
	ToolName              string
	ToolUseID             string
	Command               string
	CommandHash           string
	FilePath              string
	FileEvent             string
	URL                   string
	Domain                string
	MCPServerName         string
	MCPToolName           string
	MCPScope              string
	PermissionDestination string
	PermissionBehavior    string
	RiskType              string
	RiskScore             uint64
	Severity              string
	Evidence              string
	RedactionStatus       string
	RawEventSHA256        string
	EventSizeBytes        uint64
	DurationMs            uint64
	LatencyMs             uint64
	ToolCount             uint64
	FailureCount          uint64
	Dropped               string
	RawExcerpt            string
}

// Parser is the main log parser.
type Parser struct {
	config           Config
	parseFunc        func(string) (*LogEntry, error)
	textParseFunc    func(string) (*LogEntry, error) // Fallback for auto mode (non-JSON)
	timeLayout       string
	securityDetector *SimpleSecurityDetector
	domainDetector   *ClaudeCodeDetector
	counters         Counters
}

// New creates a new Parser with the given configuration.
func New(cfg Config) *Parser {
	p := &Parser{
		config:     cfg,
		timeLayout: time.RFC3339, // §10.1 timezone policy: RFC3339
	}

	maxFieldLen := cfg.MaxFieldLength
	if maxFieldLen <= 0 {
		maxFieldLen = 10 * 1024 // Default 10KB (§14.2 D-004)
	}
	if cfg.SecurityPatterns {
		p.securityDetector = NewSimpleSecurityDetector(maxFieldLen)
	}
	// Domain detector is always wired up: §14.2 D-001 says the detector must
	// emit risk_type/score/severity/evidence. It is a *fallback* — it only
	// runs if the upstream logger left risk_type empty/none.
	p.domainDetector = NewClaudeCodeDetector(maxFieldLen)

	switch cfg.LogFormat {
	case "auto":
		p.parseFunc = p.parseAuto
		p.textParseFunc = p.parseCombined
	case "common":
		p.parseFunc = p.parseCommon
	case "json":
		p.parseFunc = p.parseJSON
	case "custom":
		p.parseFunc = p.parseCustom
	default: // "combined"
		p.parseFunc = p.parseCombined
	}

	return p
}

// CountersSnapshot returns the current parser counters (atomic snapshot).
func (p *Parser) CountersSnapshot() Counters { return p.counters.Snapshot() }

// Parse parses a single log line and returns a LogEntry.
//
// On a malformed line (invalid JSON or missing required fields per §10.3),
// Parse increments the malformed counter and returns an error; callers are
// expected to drop the line and continue (§14.1 P-003 / ES-006).
func (p *Parser) Parse(line string) (*LogEntry, error) {
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	entry, err := p.parseFunc(line)
	if err != nil {
		p.counters.IncMalformed()
		return nil, err
	}

	entry.Raw = line

	// P004 / P-005: prevent nil map panic during GOB encode.
	if entry.Headers == nil {
		entry.Headers = make(map[string]string)
	}

	// Defense-in-depth redaction (§17.1): the hook logger should already have
	// redacted, but we re-apply on the string fields most likely to leak.
	p.redactInPlace(entry)

	// Compute command_hash if logger did not (§14.1 P-004 / detector contract).
	if entry.Command != "" && entry.CommandHash == "" {
		entry.CommandHash = shortHash(entry.Command)
	}

	// Compute event_size_bytes from raw line if logger omitted it.
	if entry.EventSizeBytes == 0 {
		entry.EventSizeBytes = uint64(len(line))
	}

	// Compute raw_event_sha256 if logger omitted it.
	if entry.RawEventSHA256 == "" {
		sum := sha256.Sum256([]byte(line))
		entry.RawEventSHA256 = hex.EncodeToString(sum[:])
	}

	// Domain detector (§14.2 D-001): only fires if logger left risk_type empty
	// or "none". Updates counters when a risk is assigned.
	if p.domainDetector != nil && p.domainDetector.Classify(entry) {
		p.counters.IncDetected()
	}

	// Generic security pattern detector (sqli/xss/cmd-injection/path-traversal),
	// stored on entry.SecurityThreat. This is a secondary informational signal.
	if p.securityDetector != nil {
		p.detectSecurityPatterns(entry)
	}

	return entry, nil
}

// parseCombined parses a Combined log format line.
// Not used by claude_code (JSONL only). Kept as a stub to preserve the
// dev-kit common parser surface for shared tests.
func (p *Parser) parseCombined(line string) (*LogEntry, error) {
	return nil, fmt.Errorf("combined format not supported by claude_code parser")
}

// parseCommon parses a Common log format line.
// Not used by claude_code (JSONL only).
func (p *Parser) parseCommon(line string) (*LogEntry, error) {
	return nil, fmt.Errorf("common format not supported by claude_code parser")
}

// parseJSON parses a JSON format log line (claude_code_security_event/v1 schema).
//
// Implements §14.1:
//   - P-001: parse JSONL → LogEntry
//   - P-002: unknown fields are ignored (json.Unmarshal into map silently)
//   - P-003: required fields (schema_version, received_at, session_id,
//     hook_event_name, risk_type, risk_score) — missing required fields
//     mark the line malformed.
//   - P-004: timestamp parse fallback to received_at, then time.Now()
//   - P-005: Headers map initialized
func (p *Parser) parseJSON(line string) (*LogEntry, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	entry := &LogEntry{}

	// --- Timestamp ---
	// claude_code uses `received_at` (§10.1). Fallbacks: `timestamp`, `time`, time.Now().
	if v, ok := raw["received_at"].(string); ok && v != "" {
		entry.Timestamp = parseTimestamp(v)
		entry.ReceivedAt = v
	} else if v, ok := raw["timestamp"].(string); ok {
		entry.Timestamp = parseTimestamp(v)
	} else if v, ok := raw["time"].(string); ok {
		entry.Timestamp = parseTimestamp(v)
	} else {
		entry.Timestamp = time.Now()
	}

	// --- ${DOMAIN_FIELDS_PARSE_JSON} (§10.2 claude_code.*) ---
	stringField(raw, "schema_version", &entry.SchemaVersion)
	stringField(raw, "logger_version", &entry.LoggerVersion)
	stringField(raw, "host", &entry.Host)
	stringField(raw, "user", &entry.User)
	stringField(raw, "session_id", &entry.SessionID)
	stringField(raw, "transcript_path", &entry.TranscriptPath)
	stringField(raw, "cwd", &entry.Cwd)
	stringField(raw, "permission_mode", &entry.PermissionMode)
	// hook_event_name (§6.1.1 HL-004) is normalized to event_name in plugin field naming.
	if v, ok := raw["event_name"].(string); ok && v != "" {
		entry.EventName = v
	} else {
		stringField(raw, "hook_event_name", &entry.EventName)
	}
	stringField(raw, "source", &entry.Source)
	stringField(raw, "tool_name", &entry.ToolName)
	stringField(raw, "tool_use_id", &entry.ToolUseID)
	stringField(raw, "command", &entry.Command)
	stringField(raw, "command_hash", &entry.CommandHash)
	stringField(raw, "file_path", &entry.FilePath)
	stringField(raw, "file_event", &entry.FileEvent)
	stringField(raw, "url", &entry.URL)
	stringField(raw, "domain", &entry.Domain)
	stringField(raw, "mcp_server_name", &entry.MCPServerName)
	stringField(raw, "mcp_tool_name", &entry.MCPToolName)
	stringField(raw, "mcp_scope", &entry.MCPScope)
	stringField(raw, "permission_destination", &entry.PermissionDestination)
	stringField(raw, "permission_behavior", &entry.PermissionBehavior)
	stringField(raw, "risk_type", &entry.RiskType)
	uintField(raw, "risk_score", &entry.RiskScore)
	stringField(raw, "severity", &entry.Severity)
	stringField(raw, "evidence", &entry.Evidence)
	stringField(raw, "redaction_status", &entry.RedactionStatus)
	stringField(raw, "raw_event_sha256", &entry.RawEventSHA256)
	uintField(raw, "event_size_bytes", &entry.EventSizeBytes)
	uintField(raw, "duration_ms", &entry.DurationMs)
	uintField(raw, "latency_ms", &entry.LatencyMs)
	uintField(raw, "tool_count", &entry.ToolCount)
	uintField(raw, "failure_count", &entry.FailureCount)
	// dropped is bool in JSON (§10.1) but exposed as string ("true"/"false") to Falco (§10.2).
	if v, ok := raw["dropped"].(bool); ok {
		if v {
			entry.Dropped = "true"
		} else {
			entry.Dropped = "false"
		}
	} else {
		stringField(raw, "dropped", &entry.Dropped)
	}
	stringField(raw, "raw_excerpt", &entry.RawExcerpt)

	// Headers map from optional `headers` object
	if hdrs, ok := raw["headers"].(map[string]interface{}); ok {
		entry.Headers = make(map[string]string)
		for k, v := range hdrs {
			if s, ok := v.(string); ok {
				entry.Headers[strings.ToLower(k)] = s
			}
		}
	}

	// §14.1 P-003: required-field validation.
	//
	// We are intentionally lenient: if any required field is missing we still
	// surface the partial event (with as much detail as the line had) but
	// flag the malformed counter via the returned error. Callers that want
	// strict-drop behaviour can check Parse()'s error; callers that want
	// fail-open behaviour can ignore the error and inspect entry.SchemaVersion.
	if err := validateRequired(raw, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// validateRequired enforces §10.3 required-fields. The check is loose: we
// require schema_version/family match and presence of session_id, hook_event_name,
// risk_type, risk_score (any of received_at/timestamp/time satisfies the
// timestamp requirement).
func validateRequired(raw map[string]interface{}, entry *LogEntry) error {
	if entry.SchemaVersion == "" {
		return fmt.Errorf("missing schema_version")
	}
	if !strings.HasPrefix(entry.SchemaVersion, supportedSchemaPrefix) {
		return fmt.Errorf("unsupported schema_version: %q", entry.SchemaVersion)
	}
	if entry.SessionID == "" {
		return fmt.Errorf("missing session_id")
	}
	if entry.EventName == "" {
		return fmt.Errorf("missing event_name / hook_event_name")
	}
	if entry.ReceivedAt == "" {
		// We accept the alternative timestamp fields.
		if _, ok := raw["timestamp"].(string); !ok {
			if _, ok := raw["time"].(string); !ok {
				return fmt.Errorf("missing received_at / timestamp")
			}
		}
	}
	// risk_type/risk_score are nominally required but the hook logger emits
	// defaults ("none"/0); we do not fail on them. We also do not fail on
	// schema-major mismatches when the family prefix matched.
	return nil
}

// redactInPlace applies §17.1 redaction to the user-controllable string
// fields and updates the redaction_status.
func (p *Parser) redactInPlace(entry *LogEntry) {
	maxField := p.config.MaxFieldLength
	if maxField <= 0 {
		maxField = 10 * 1024
	}

	anyRedacted := false
	anyTruncated := false

	apply := func(in string) string {
		if in == "" {
			return in
		}
		capped, truncated := CapField(in, maxField)
		if truncated {
			anyTruncated = true
		}
		out, redacted := Redact(capped)
		if redacted {
			anyRedacted = true
			p.counters.IncRedacted()
		}
		return out
	}

	entry.Command = apply(entry.Command)
	entry.Evidence = apply(entry.Evidence)
	entry.URL = apply(entry.URL)
	entry.FilePath = apply(entry.FilePath)
	entry.RawExcerpt = apply(entry.RawExcerpt)

	// Only upgrade redaction_status; never downgrade. If logger already said
	// "redacted" or "truncated" we keep that; if logger said "none" or empty
	// and the parser saw something, we upgrade.
	current := entry.RedactionStatus
	if current == "" || current == "none" {
		entry.RedactionStatus = RedactionStatus(anyRedacted, anyTruncated, false)
	}
}

// stringField copies a string field from raw into dst if present and non-empty.
func stringField(raw map[string]interface{}, key string, dst *string) {
	if v, ok := raw[key].(string); ok {
		*dst = v
	}
}

// uintField copies a numeric (JSON number → float64) field into a uint64 dst.
// Negative values are clipped to 0 (§10.2 type-conversion guidance).
func uintField(raw map[string]interface{}, key string, dst *uint64) {
	if v, ok := raw[key].(float64); ok {
		if v < 0 {
			*dst = 0
			return
		}
		*dst = uint64(v)
	}
}

// shortHash returns the first 16 hex chars of SHA-256(input).
// §14.1 P-005 short hash for command identity.
func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:16]
}

// parseTimestamp attempts to parse a timestamp string with multiple formats.
// Falls back to time.Now() if all formats fail (§14.1 P-004).
func parseTimestamp(s string) time.Time {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Now()
}

// parseCustom parses a custom format log line.
func (p *Parser) parseCustom(line string) (*LogEntry, error) {
	return nil, fmt.Errorf("custom format not yet implemented")
}

// parseAuto auto-detects the log format and parses accordingly.
func (p *Parser) parseAuto(line string) (*LogEntry, error) {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "{") {
		return p.parseJSON(line)
	}
	if p.textParseFunc != nil {
		return p.textParseFunc(line)
	}
	return nil, fmt.Errorf("auto-detect: non-JSON line but no text parser configured")
}

// detectSecurityPatterns checks for security threats in the log entry.
//
// For claude_code, the primary detection pipeline is:
//
//	hook logger (redaction + risk_type/risk_score) → JSONL → plugin → Falco rule.
//
// The plugin-side SimpleSecurityDetector is therefore a *secondary* signal that
// looks at `evidence` + `command` for opportunistic SQLi/XSS/path-traversal/CMDi
// patterns. This is informational only; rules should still rely on
// claude_code.risk_type / claude_code.risk_score from the detector (§14.2 D-001).
func (p *Parser) detectSecurityPatterns(entry *LogEntry) {
	if p.securityDetector == nil {
		return
	}

	input := entry.Evidence
	if input == "" {
		input = entry.Command
	}
	if input == "" {
		return
	}

	// URL decode up to 3 levels (§14.2 D-003)
	decoded := input
	for i := 0; i < 3; i++ {
		newDecoded, err := url.QueryUnescape(decoded)
		if err != nil || newDecoded == decoded {
			break
		}
		decoded = newDecoded
	}

	threatType, found := p.securityDetector.DetectSecurityThreat(decoded)
	if found {
		switch threatType {
		case "cmd_injection":
			entry.SecurityThreat = CommandInjection
		case "sqli":
			entry.SecurityThreat = SQLInjection
		case "xss":
			entry.SecurityThreat = XSSAttempt
		case "path_traversal":
			entry.SecurityThreat = PathTraversal
		}
	}
}

package parser

import (
	"strings"
)

// ClaudeCodeDetector implements the T-001..T-018 detector categories described
// in requirements v3 §12.1 / §14.2. Detection is deterministic, ReDoS safe
// (string-only operations) and bounded.
//
// Per requirements §15.1 / §6.1 HL-005, the hook logger is the primary place
// for risk classification (it has the full hook input and immediate context).
// The plugin-side detector here acts as a *fallback*: if the upstream logger
// did not assign a risk_type, the plugin tries to classify the event based on
// the normalized fields it received.
//
// The detector NEVER overwrites a non-empty risk_type set by the logger
// (D-006: rule and detector overlap is allowed; complex judgements live in the
// logger / detector layer rather than rules).
type ClaudeCodeDetector struct {
	maxFieldLen int
}

// NewClaudeCodeDetector returns a detector with the given per-field input cap.
// maxFieldLen <= 0 falls back to 10 KiB (§14.2 D-004).
func NewClaudeCodeDetector(maxFieldLen int) *ClaudeCodeDetector {
	if maxFieldLen <= 0 {
		maxFieldLen = 10 * 1024
	}
	return &ClaudeCodeDetector{maxFieldLen: maxFieldLen}
}

// Classify inspects the normalized fields of `entry` and assigns
// (risk_type, risk_score, severity) if no upstream classification exists.
//
// Returns true if a classification was applied (even if upstream value was
// preserved, the function inspects the entry but does not flag a change).
func (d *ClaudeCodeDetector) Classify(entry *LogEntry) bool {
	// Respect upstream risk_type ("none" is treated as not-yet-classified per
	// §10.3 default-emission rules in the logger; the logger emits "none"
	// when it does not detect a risk).
	if entry.RiskType != "" && entry.RiskType != "none" {
		return false
	}

	// T-001 dangerous_bash is left to the logger / Falco rule layer for v0.1
	// because rule-side string match (claude_code.command icontains ...) is
	// already deterministic and the patterns are owned by the rule file.
	// The plugin-side SimpleSecurityDetector still flags `cmd_injection` etc.
	// in entry.SecurityThreat; rules can use both.

	// Order: most specific / highest confidence first.
	switch {
	case d.detectPermissionBypass(entry):
		d.set(entry, "permission_bypass", 90, "critical", "permission_mode is bypass")
	case d.detectHookDisabled(entry):
		d.set(entry, "hook_disabled", 90, "critical", "hooks were disabled / removed")
	case d.detectMCPConfigChanged(entry):
		d.set(entry, "mcp_config_changed", 70, "warning", "MCP configuration changed")
	case d.detectSettingsModified(entry):
		d.set(entry, "settings_modified", 60, "warning", "claude settings file modified")
	case d.detectPermissionUpdate(entry):
		d.set(entry, "permission_update", 60, "warning", "permission update requested")
	case d.detectGitDestructive(entry):
		d.set(entry, "git_destructive", 70, "warning", "destructive git operation")
	case d.detectSecretExfiltration(entry):
		d.set(entry, "secret_exfiltration", 90, "critical", "secret exfiltration pattern")
	case d.detectSensitiveFileRead(entry):
		d.set(entry, "sensitive_file_read", 70, "warning", "sensitive file read attempt")
	case d.detectWorkspaceEscape(entry):
		d.set(entry, "workspace_escape", 60, "warning", "workspace escape pattern")
	default:
		return false
	}
	return true
}

func (d *ClaudeCodeDetector) set(entry *LogEntry, riskType string, score uint64, severity, evidence string) {
	entry.RiskType = riskType
	entry.RiskScore = score
	entry.Severity = severity
	if entry.Evidence == "" {
		// Evidence already truncated upstream; just note the detector reason.
		entry.Evidence = "detector:" + evidence
	}
}

// --- Per-category logic ---

// T-003 permission_bypass: permission_mode is set to "bypassPermissions" or
// command contains the dangerous --dangerously-skip-permissions flag.
func (d *ClaudeCodeDetector) detectPermissionBypass(entry *LogEntry) bool {
	if entry.PermissionMode == "bypassPermissions" {
		return true
	}
	cmd := d.bound(entry.Command)
	lower := strings.ToLower(cmd)
	return strings.Contains(lower, "--dangerously-skip-permissions") ||
		strings.Contains(lower, "skipdangerousmodepermissionprompt")
}

// T-006 hook_disabled: ConfigChange referencing disableAllHooks=true OR a hook
// removal in evidence (logger emits the diff context as evidence).
func (d *ClaudeCodeDetector) detectHookDisabled(entry *LogEntry) bool {
	if entry.EventName != "ConfigChange" {
		return false
	}
	combined := strings.ToLower(d.bound(entry.Evidence) + " " + d.bound(entry.Source))
	if strings.Contains(combined, "disableallhooks") {
		return true
	}
	// Hook deletion patterns in normalized evidence.
	if strings.Contains(combined, "hook_removed") || strings.Contains(combined, "\"hooks\":{}") {
		return true
	}
	return false
}

// T-007 mcp_config_changed: ConfigChange or FileChanged referencing
// .mcp.json / ~/.claude.json / managed-mcp.json or mcp_* prefixed event_name.
func (d *ClaudeCodeDetector) detectMCPConfigChanged(entry *LogEntry) bool {
	fp := strings.ToLower(d.bound(entry.FilePath))
	src := strings.ToLower(d.bound(entry.Source))
	if entry.EventName != "ConfigChange" && entry.EventName != "FileChanged" {
		return false
	}
	if strings.Contains(fp, ".mcp.json") || strings.Contains(fp, "managed-mcp.json") || strings.HasSuffix(fp, "/.claude.json") {
		return true
	}
	if strings.Contains(src, "mcp") {
		return true
	}
	return false
}

// T-005 settings_modified: ConfigChange/FileChanged referencing
// ~/.claude/settings.json, .claude/settings.json, .claude/settings.local.json.
func (d *ClaudeCodeDetector) detectSettingsModified(entry *LogEntry) bool {
	if entry.EventName != "ConfigChange" && entry.EventName != "FileChanged" {
		return false
	}
	fp := strings.ToLower(d.bound(entry.FilePath))
	return strings.HasSuffix(fp, "/.claude/settings.json") ||
		strings.HasSuffix(fp, "/settings.json") && strings.Contains(fp, ".claude") ||
		strings.HasSuffix(fp, "/settings.local.json")
}

// T-004 permission_update: PermissionRequest event or evidence containing
// "updatedPermissions"/"add allow"/destination=userSettings/projectSettings.
func (d *ClaudeCodeDetector) detectPermissionUpdate(entry *LogEntry) bool {
	if entry.EventName == "PermissionRequest" {
		return true
	}
	dest := strings.ToLower(entry.PermissionDestination)
	if dest == "usersettings" || dest == "projectsettings" {
		return true
	}
	ev := strings.ToLower(d.bound(entry.Evidence))
	return strings.Contains(ev, "updatedpermissions") || strings.Contains(ev, "add allow")
}

// T-011 git_destructive: command matches git destructive patterns.
func (d *ClaudeCodeDetector) detectGitDestructive(entry *LogEntry) bool {
	cmd := strings.ToLower(d.bound(entry.Command))
	if cmd == "" {
		return false
	}
	patterns := []string{
		"git push -f", "git push --force",
		"git reset --hard",
		"git clean -fdx", "git clean -fd",
		"rm -rf .git",
		"git branch -d",  // forced branch delete (uppercase D too)
		"git branch -D ", // explicit forced delete with arg
		"git update-ref -d",
	}
	for _, p := range patterns {
		if strings.Contains(cmd, p) {
			return true
		}
	}
	return false
}

// T-002 secret_exfiltration: command/url evidence matches a "send credential
// material to network egress" shape. Examples:
//   - curl/wget/scp/nc with .env / id_rsa / .aws/credentials / kubeconfig
//   - pbcopy of /etc/passwd or any of the sensitive paths above
func (d *ClaudeCodeDetector) detectSecretExfiltration(entry *LogEntry) bool {
	cmd := strings.ToLower(d.bound(entry.Command))
	if cmd == "" {
		return false
	}

	hasNetTool := false
	for _, tool := range []string{"curl ", "wget ", "scp ", "nc ", "ncat ", "pbcopy", "rsync ", "ssh "} {
		if strings.Contains(cmd, tool) {
			hasNetTool = true
			break
		}
	}
	if !hasNetTool {
		return false
	}

	for _, secret := range []string{
		".env", "id_rsa", "id_ed25519", "id_dsa", "id_ecdsa",
		".aws/credentials", ".kube/config", "kubeconfig",
		".gcp/", "gcloud/credentials",
		"/etc/shadow",
	} {
		if strings.Contains(cmd, secret) {
			return true
		}
	}
	return false
}

// T-009 sensitive_file_read: tool_name=Read/Grep/Glob with file_path matching
// known sensitive patterns. Falls back to inspecting command for cat/less/head/tail.
func (d *ClaudeCodeDetector) detectSensitiveFileRead(entry *LogEntry) bool {
	target := strings.ToLower(d.bound(entry.FilePath))
	if target == "" {
		// Look in command for `cat .env` etc. when tool=Bash.
		cmd := strings.ToLower(d.bound(entry.Command))
		for _, prefix := range []string{"cat ", "less ", "head ", "tail ", "more ", "bat "} {
			if idx := strings.Index(cmd, prefix); idx >= 0 {
				target = cmd[idx+len(prefix):]
				break
			}
		}
	}
	if target == "" {
		return false
	}
	for _, pat := range []string{
		".env", "id_rsa", "id_ed25519", "id_dsa", "id_ecdsa",
		".aws/credentials", ".kube/config", "kubeconfig",
		".git/config", ".npmrc", ".pypirc",
		"/etc/shadow", "/etc/passwd",
	} {
		if strings.Contains(target, pat) {
			return true
		}
	}
	return false
}

// T-010 workspace_escape: file_path or command contains absolute path outside
// repo (cwd), explicit ".." segments, or known sensitive directories.
func (d *ClaudeCodeDetector) detectWorkspaceEscape(entry *LogEntry) bool {
	fp := d.bound(entry.FilePath)
	cmd := d.bound(entry.Command)
	cwd := d.bound(entry.Cwd)

	// Explicit ".." in either field.
	if strings.Contains(fp, "..") || strings.Contains(cmd, "..") {
		return true
	}

	candidates := []string{fp, cmd}
	for _, c := range candidates {
		lower := strings.ToLower(c)
		for _, sensitive := range []string{
			"/etc/", "/root/", "/var/log/",
			"/private/etc/", // macOS canonical
			"$home/.ssh", "/.ssh/",
		} {
			if strings.Contains(lower, sensitive) {
				return true
			}
		}
	}

	// Absolute file_path that does not start with cwd (and cwd is non-empty)
	// counts as escape even without explicit ".." (e.g. /Users/other/...).
	if cwd != "" && fp != "" && strings.HasPrefix(fp, "/") && !strings.HasPrefix(fp, cwd) {
		return true
	}
	return false
}

// bound truncates s to at most d.maxFieldLen runes (we use bytes here for
// performance; D-004 talks about bytes).
func (d *ClaudeCodeDetector) bound(s string) string {
	if len(s) <= d.maxFieldLen {
		return s
	}
	return s[:d.maxFieldLen]
}

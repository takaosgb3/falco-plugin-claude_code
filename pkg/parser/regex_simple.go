package parser

import (
	"strings"
)

// SimpleSecurityDetector provides string-matching based security detection.
// No regex is used (ReDoS safe; §12.2). Uses strings.Contains / strings.ToLower.
type SimpleSecurityDetector struct {
	maxInputLength int
}

// NewSimpleSecurityDetector creates a new SimpleSecurityDetector with the given max input length.
func NewSimpleSecurityDetector(maxInputLength int) *SimpleSecurityDetector {
	return &SimpleSecurityDetector{
		maxInputLength: maxInputLength,
	}
}

// DetectSecurityThreat checks for security threats (4 categories).
// C-002: This aggregation function handles SQLi, XSS, PathTraversal, CmdInjection.
// Suspicious Agent detection uses DetectSuspiciousAgent() separately.
func (d *SimpleSecurityDetector) DetectSecurityThreat(input string) (string, bool) {
	if len(input) > d.maxInputLength {
		input = input[:d.maxInputLength] // truncate して続行（先頭 10KB 内の脅威を検出, T2-1）
	}

	lower := strings.ToLower(input)

	if d.DetectCommandInjection(lower) {
		return "cmd_injection", true
	}
	if d.DetectSQLInjection(lower) {
		return "sqli", true
	}
	if d.DetectXSS(lower) {
		return "xss", true
	}
	if d.DetectPathTraversal(lower) {
		return "path_traversal", true
	}

	return "", false
}

// DetectSQLInjection checks for SQL injection patterns.
func (d *SimpleSecurityDetector) DetectSQLInjection(input string) bool {
	lower := strings.ToLower(input)
	patterns := []string{
		"' or '", "' and '", "' or 1=1", "' and 1=1",
		"union select", "union all select",
		"select from", "select * from",
		"drop table", "drop database",
		"insert into", "delete from", "update set",
		"'; --", "';--",
		"sleep(", "benchmark(", "waitfor delay",
		"1=1", "1'='1",
		"/*", "*/",
		"@@version", "information_schema",
		"load_file(", "into outfile", "into dumpfile",
	}

	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	if idx := strings.Index(lower, "0x"); idx >= 0 {
		rest := lower[idx+2:]
		if len(rest) >= 2 && isHexString(rest[:2]) {
			return true
		}
	}

	return false
}

// DetectXSS checks for Cross-Site Scripting patterns.
func (d *SimpleSecurityDetector) DetectXSS(input string) bool {
	lower := strings.ToLower(input)
	patterns := []string{
		"<script", "</script", "<iframe", "</iframe",
		"javascript:", "vbscript:",
		"onerror=", "onload=", "onclick=", "onmouseover=",
		"onfocus=", "onblur=", "onsubmit=",
		"<img src=", "<svg onload=", "<body onload=",
		"alert(", "confirm(", "prompt(",
		"document.cookie", "document.write",
		"eval(", "expression(",
	}

	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// DetectPathTraversal checks for path traversal patterns.
func (d *SimpleSecurityDetector) DetectPathTraversal(input string) bool {
	lower := strings.ToLower(input)
	patterns := []string{
		"../", "..\\",
		"%2e%2e%2f", "%2e%2e/", "..%2f",
		"%2e%2e%5c", "%2e%2e\\", "..%5c",
		"%c0%ae%c0%ae",    // Overlong UTF-8 encoding
		"%252e%252e%252f", // Double URL encoding
		"/etc/passwd", "/etc/shadow", "/etc/hosts",
		"/proc/self", "/proc/version",
		"boot.ini", "win.ini",
		"web.config", "wp-config.php",
	}

	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// DetectCommandInjection checks for command injection patterns.
func (d *SimpleSecurityDetector) DetectCommandInjection(input string) bool {
	lower := strings.ToLower(input)

	simplePatterns := []string{
		"$(", "`",
		"%0a", "%0d",
	}
	for _, pattern := range simplePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	separators := []string{";", "|", "&", "&&", "||"}
	commands := []string{
		"ls", "cat", "echo", "rm", "bash", "sh", "curl", "wget",
		"python", "perl", "ruby", "php", "nc", "ncat",
		"chmod", "chown", "mkdir", "cp", "mv",
		"id", "whoami", "uname", "ifconfig", "ip",
		"ping", "traceroute", "nslookup", "dig",
	}

	for _, sep := range separators {
		if idx := strings.Index(lower, sep); idx >= 0 {
			after := strings.TrimSpace(lower[idx+len(sep):])
			for _, cmd := range commands {
				if strings.HasPrefix(after, cmd+" ") || strings.HasPrefix(after, cmd+"/") || after == cmd {
					return true
				}
			}
		}
	}

	return false
}

// DetectSuspiciousAgent checks for suspicious User-Agent strings.
// C-002: This is a separate method called specifically for User-Agent field.
func (d *SimpleSecurityDetector) DetectSuspiciousAgent(input string) bool {
	lower := strings.ToLower(input)

	prefixes := []string{
		"sqlmap", "nikto/", "nmap", "acunetix",
		"dirbuster", "gobuster", "wpscan",
		"masscan", "zgrab", "nuclei",
		"burpsuite", "owasp",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) || strings.Contains(lower, prefix) {
			return true
		}
	}

	if strings.HasSuffix(lower, "scanner") {
		return true
	}

	suspiciousSubstrings := []string{
		"bot scanner", "vulnerability scanner",
		"security scanner", "web scanner",
	}
	for _, substr := range suspiciousSubstrings {
		if strings.Contains(lower, substr) {
			return true
		}
	}

	return false
}

// isHexString checks if all characters are hexadecimal digits.
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return len(s) > 0
}

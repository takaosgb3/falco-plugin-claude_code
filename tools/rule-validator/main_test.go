package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	return path
}

// TestValidateFile_ProjectFiles exercises the validator against the actual
// rule files this Phase 3 commit ships. The test fails (and the validator
// reports specifics) if any of the production rules regress.
func TestValidateFile_ProjectFiles(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		minRules int
	}{
		{"main rules", "../../rules/claude-code_rules.yaml", 18},
		{"health rules", "../../rules/claude_code_health.yaml", 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issues, stats, err := validateFile(tc.path)
			if err != nil {
				t.Fatalf("validateFile(%s): %v", tc.path, err)
			}
			if len(issues) != 0 {
				for _, iss := range issues {
					t.Errorf("issue: %s", iss)
				}
			}
			if stats.rules < tc.minRules {
				t.Errorf("expected at least %d rules in %s, got %d",
					tc.minRules, tc.path, stats.rules)
			}
		})
	}
}

// TestC4_RPVMustBeFirst asserts that required_plugin_versions outside index 0
// is flagged (C4).
func TestC4_RPVMustBeFirst(t *testing.T) {
	body := `- macro: foo
  condition: claude_code.event_name = "x"

- required_plugin_versions:
    - name: claude-code
      version: 0.1.0
`
	p := writeTemp(t, "bad.yaml", body)
	issues, _, err := validateFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(issues, "C4") {
		t.Fatalf("expected C4 violation, got: %v", issues)
	}
}

// TestC6_SourceMissingFlagged asserts that a rule without source is reported.
func TestC6_SourceMissingFlagged(t *testing.T) {
	body := `- required_plugin_versions:
    - name: claude-code
      version: 0.1.0

- rule: missing-source
  desc: x
  condition: claude_code.event_name = "x"
  output: x
  priority: NOTICE
  tags: [t]
`
	p := writeTemp(t, "bad.yaml", body)
	issues, _, err := validateFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(issues, "C5") && !hasCode(issues, "C6") {
		t.Fatalf("expected C5/C6 violation for missing source, got: %v", issues)
	}
}

// TestC6_WrongSourceFlagged asserts that a non-claude_code source is reported.
func TestC6_WrongSourceFlagged(t *testing.T) {
	body := `- required_plugin_versions:
    - name: claude-code
      version: 0.1.0

- rule: wrong-source
  desc: x
  condition: claude_code.event_name = "x"
  output: x
  priority: NOTICE
  source: syscall
  tags: [t]
`
	p := writeTemp(t, "bad.yaml", body)
	issues, _, err := validateFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(issues, "C6") {
		t.Fatalf("expected C6 violation for source=syscall, got: %v", issues)
	}
}

// TestC7_EvtTypeFlagged asserts that evt.type usage is reported.
func TestC7_EvtTypeFlagged(t *testing.T) {
	body := `- required_plugin_versions:
    - name: claude-code
      version: 0.1.0

- rule: bad
  desc: x
  condition: evt.type = "open" and claude_code.event_name = "x"
  output: x
  priority: NOTICE
  source: claude_code
  tags: [t]
`
	p := writeTemp(t, "bad.yaml", body)
	issues, _, err := validateFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(issues, "C7") {
		t.Fatalf("expected C7 violation for evt.type, got: %v", issues)
	}
}

// TestC8_UndefinedMacroFlagged asserts that referencing an undefined macro
// from a rule condition is flagged.
func TestC8_UndefinedMacroFlagged(t *testing.T) {
	body := `- required_plugin_versions:
    - name: claude-code
      version: 0.1.0

- rule: bad
  desc: x
  condition: not_a_real_macro and claude_code.event_name = "x"
  output: x
  priority: NOTICE
  source: claude_code
  tags: [t]
`
	p := writeTemp(t, "bad.yaml", body)
	issues, _, err := validateFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(issues, "C8") {
		t.Fatalf("expected C8 violation for undefined macro, got: %v", issues)
	}
}

// TestReferencedNames_StripsStringsAndFields ensures we don't false-flag
// identifiers that appear inside double-quoted literals or after claude_code.
func TestReferencedNames_StripsStringsAndFields(t *testing.T) {
	got := referencedNames(`claude_code.command icontains "rm -rf /" and my_macro`)
	if !contains(got, "my_macro") {
		t.Errorf("expected my_macro in result, got %v", got)
	}
	for _, n := range got {
		if n == "rm" || n == "rf" {
			t.Errorf("string literal token leaked into referenced names: %q", n)
		}
		if n == "command" {
			t.Errorf("claude_code.* trailing identifier leaked: %q", n)
		}
	}
}

// TestC9_BadPriorityFlagged asserts that an invalid priority is reported.
func TestC9_BadPriorityFlagged(t *testing.T) {
	body := `- required_plugin_versions:
    - name: claude-code
      version: 0.1.0

- rule: bad
  desc: x
  condition: claude_code.event_name = "x"
  output: x
  priority: HORRIBLE
  source: claude_code
  tags: [t]
`
	p := writeTemp(t, "bad.yaml", body)
	issues, _, err := validateFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(issues, "C9") {
		t.Fatalf("expected C9 violation for unknown priority, got: %v", issues)
	}
}

// --- helpers ---

func hasCode(issues []Issue, code string) bool {
	for _, i := range issues {
		if i.Code == code {
			return true
		}
	}
	return false
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// Defensive: TestMain prints a hint if the YAML lib is not vendored when run
// in unusual environments. Strip when go.sum is healthy.
func TestMain(m *testing.M) {
	// Sanity: detect the go.mod root so error messages are useful when the
	// validator is run outside the repo. (Diagnostic only.)
	if _, err := os.Stat("../../go.mod"); err != nil {
		_ = strings.TrimSpace
	}
	os.Exit(m.Run())
}

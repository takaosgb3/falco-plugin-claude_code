// Command rule-validator performs static validation of Falco rule YAML files
// for the claude-code plugin in environments where neither `falco -V` nor
// `yamllint` is available (e.g. macOS dev box without Falco installed).
//
// Phase 3 quality gate (see CLAUDE.md and .claude/agents/plugin-dev-workflow.md):
// when Falco itself cannot be invoked, this validator is the substitute. It
// covers YAML syntax + the structural rules that PROBLEM_PATTERNS.md / R-001..R-011
// would otherwise leave to runtime.
//
// Checks performed:
//
//	C1  YAML parses cleanly.
//	C2  Top-level is a sequence (Falco rule files always are).
//	C3  Each entry is exactly one of: required_plugin_versions, list, macro, rule.
//	C4  required_plugin_versions appears exactly once and as the first entry
//	    (R-003 / P020).
//	C5  Every rule entry has the required keys: rule, desc, condition, output,
//	    priority, source, tags.
//	C6  Every rule has source == "claude_code" (P003 / R-001).
//	C7  No rule references evt.type anywhere in its condition or output
//	    (P005 / R-002).
//	C8  Every macro / list referenced in any condition is defined in the same
//	    file. Built-in Falco operators (and / or / not / in / startswith /
//	    endswith / icontains / contains / pmatch) and field accessors
//	    (claude_code.*) are not flagged.
//	C9  priority is one of EMERGENCY/ALERT/CRITICAL/ERROR/WARNING/NOTICE/
//	    INFORMATIONAL/DEBUG (Falco accepted set).
//
// Exit code 0 = all files passed. Exit code 1 = at least one violation.
//
// Usage:
//
//	rule-validator rules/claude-code_rules.yaml rules/claude_code_health.yaml
package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// expectedSource is hard-coded for this plugin. The validator is intentionally
// not generic — making the constraint visible at the top of the file is more
// valuable than parametrisation.
const expectedSource = "claude_code"

// validPriorities follows Falco's accepted set (case-insensitive). See
// https://falco.org/docs/rules/basic-elements/#priority.
var validPriorities = map[string]struct{}{
	"EMERGENCY":     {},
	"ALERT":         {},
	"CRITICAL":      {},
	"ERROR":         {},
	"WARNING":       {},
	"NOTICE":        {},
	"INFORMATIONAL": {},
	"INFO":          {},
	"DEBUG":         {},
}

// falcoOperators are the language-level tokens (NOT macro / list names) that
// can legitimately appear in a condition. Anything else that looks like an
// identifier MUST resolve to a macro or a list, otherwise C8 fails.
//
// The "field" namespace claude_code.* is excluded by a regex ahead of the
// identifier check (see referencedNames).
var falcoOperators = map[string]struct{}{
	"and":         {},
	"or":          {},
	"not":         {},
	"in":          {},
	"intersects":  {},
	"pmatch":      {},
	"contains":    {},
	"icontains":   {},
	"startswith":  {},
	"endswith":    {},
	"glob":        {},
	"exists":      {},
	"true":        {},
	"false":       {},
	"null":        {},
	"none":        {},
	// Quantifiers / sentinel tokens that occasionally show up in Falco
	// conditions. Including them avoids false positives without weakening the
	// macro reference check (those are uppercase / dotted words).
}

// Issue is a single validator finding.
type Issue struct {
	File string
	Path string // YAML path, e.g. "rules[3].condition" or "macros[1]"
	Code string
	Msg  string
}

func (i Issue) String() string {
	return fmt.Sprintf("[%s] %s: %s -- %s", i.Code, i.File, i.Path, i.Msg)
}

// ruleEntry is the shape of a Falco rule. We use map[string]any during walk
// and project into this struct for the per-rule check (C5/C6/C7/C9).
type ruleEntry struct {
	Name      string
	Condition string
	Output    string
	Priority  string
	Source    string
	Tags      []string
	Desc      string
	Path      string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: rule-validator <yaml-file> [<yaml-file>...]")
		os.Exit(2)
	}

	totalIssues := 0
	totalRules := 0
	totalMacros := 0
	totalLists := 0

	for _, path := range os.Args[1:] {
		issues, stats, err := validateFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", path, err)
			totalIssues++
			continue
		}
		for _, iss := range issues {
			fmt.Println(iss.String())
		}
		totalIssues += len(issues)
		totalRules += stats.rules
		totalMacros += stats.macros
		totalLists += stats.lists

		if len(issues) == 0 {
			fmt.Printf("OK   %s (%d rules, %d macros, %d lists)\n",
				path, stats.rules, stats.macros, stats.lists)
		} else {
			fmt.Printf("FAIL %s (%d issues, %d rules)\n", path, len(issues), stats.rules)
		}
	}

	fmt.Printf("\nSummary: %d rules, %d macros, %d lists across %d files. %d issues.\n",
		totalRules, totalMacros, totalLists, len(os.Args)-1, totalIssues)

	if totalIssues > 0 {
		os.Exit(1)
	}
}

type fileStats struct {
	rules, macros, lists int
}

func validateFile(path string) ([]Issue, fileStats, error) {
	stats := fileStats{}
	data, err := os.ReadFile(path) // #nosec G304 — operator-supplied path
	if err != nil {
		return nil, stats, fmt.Errorf("read: %w", err)
	}

	// C1: YAML parses.
	var doc []map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, stats, fmt.Errorf("yaml unmarshal: %w", err)
	}

	// C2 implicit: yaml.v3 returned a slice; if the file was a scalar / mapping
	// at the top, Unmarshal into []map would have failed.

	var issues []Issue

	rpvCount := 0
	rpvIndex := -1
	rules := []ruleEntry{}
	definedMacros := map[string]bool{}
	definedLists := map[string]bool{}

	for i, entry := range doc {
		if entry == nil {
			issues = append(issues, Issue{
				File: path, Path: fmt.Sprintf("[%d]", i), Code: "C3",
				Msg: "empty document entry",
			})
			continue
		}

		// C3: each entry must be exactly one of the four kinds.
		kindCount := 0
		if _, ok := entry["required_plugin_versions"]; ok {
			kindCount++
			rpvCount++
			rpvIndex = i
		}
		if name, ok := entry["list"].(string); ok {
			kindCount++
			definedLists[name] = true
			stats.lists++
		}
		if name, ok := entry["macro"].(string); ok {
			kindCount++
			definedMacros[name] = true
			stats.macros++
		}
		if _, ok := entry["rule"]; ok {
			kindCount++
			r, perRule := projectRule(path, i, entry)
			issues = append(issues, perRule...)
			if r != nil {
				rules = append(rules, *r)
				stats.rules++
			}
		}

		if kindCount == 0 {
			keys := mapKeys(entry)
			issues = append(issues, Issue{
				File: path, Path: fmt.Sprintf("[%d]", i), Code: "C3",
				Msg: fmt.Sprintf("entry is none of required_plugin_versions/list/macro/rule (keys: %v)", keys),
			})
		} else if kindCount > 1 {
			issues = append(issues, Issue{
				File: path, Path: fmt.Sprintf("[%d]", i), Code: "C3",
				Msg: "entry mixes multiple kinds (e.g. macro + rule); split into separate documents",
			})
		}
	}

	// C4: required_plugin_versions exactly once at index 0.
	if rpvCount == 0 {
		issues = append(issues, Issue{
			File: path, Path: "[0]", Code: "C4",
			Msg: "required_plugin_versions block missing (R-003 / P020)",
		})
	} else if rpvCount > 1 {
		issues = append(issues, Issue{
			File: path, Path: "*", Code: "C4",
			Msg: fmt.Sprintf("required_plugin_versions appears %d times; must appear exactly once", rpvCount),
		})
	} else if rpvIndex != 0 {
		issues = append(issues, Issue{
			File: path, Path: fmt.Sprintf("[%d]", rpvIndex), Code: "C4",
			Msg: "required_plugin_versions must be the first entry of the file",
		})
	}

	// Per-rule structural checks (C5/C6/C7/C9).
	for _, r := range rules {
		if r.Source == "" {
			issues = append(issues, Issue{
				File: path, Path: r.Path, Code: "C6",
				Msg: fmt.Sprintf("rule %q is missing source (P003 / R-001)", r.Name),
			})
		} else if r.Source != expectedSource {
			issues = append(issues, Issue{
				File: path, Path: r.Path, Code: "C6",
				Msg: fmt.Sprintf("rule %q has source=%q, expected %q (P003 / R-001)",
					r.Name, r.Source, expectedSource),
			})
		}

		// C7: no evt.type anywhere in condition / output.
		if strings.Contains(r.Condition, "evt.type") {
			issues = append(issues, Issue{
				File: path, Path: r.Path + ".condition", Code: "C7",
				Msg: fmt.Sprintf("rule %q references evt.type (P005 / R-002)", r.Name),
			})
		}
		if strings.Contains(r.Output, "evt.type") {
			issues = append(issues, Issue{
				File: path, Path: r.Path + ".output", Code: "C7",
				Msg: fmt.Sprintf("rule %q references evt.type in output (P005 / R-002)", r.Name),
			})
		}

		// C9: priority value is in Falco's accepted set.
		if r.Priority != "" {
			if _, ok := validPriorities[strings.ToUpper(r.Priority)]; !ok {
				issues = append(issues, Issue{
					File: path, Path: r.Path + ".priority", Code: "C9",
					Msg: fmt.Sprintf("rule %q has unknown priority %q",
						r.Name, r.Priority),
				})
			}
		}
	}

	// C8: every identifier referenced in a condition that is neither a Falco
	// operator nor a claude_code.* field must be a defined macro or list.
	for _, r := range rules {
		for _, name := range referencedNames(r.Condition) {
			if _, ok := falcoOperators[name]; ok {
				continue
			}
			// Boolean / numeric literals: skip.
			if isLiteral(name) {
				continue
			}
			// In Falco condition, list references appear bare (e.g. "in (foo)")
			// or after `in`. We accept either macros or lists by name.
			if definedMacros[name] || definedLists[name] {
				continue
			}
			issues = append(issues, Issue{
				File: path, Path: r.Path + ".condition", Code: "C8",
				Msg: fmt.Sprintf("rule %q references undefined identifier %q (not a defined macro/list, not a falco operator, not claude_code.*)",
					r.Name, name),
			})
		}
	}

	// Macro definitions also reference each other; check those too so the
	// graph is closed.
	for i, entry := range doc {
		macroName, ok := entry["macro"].(string)
		if !ok {
			continue
		}
		cond, _ := entry["condition"].(string)
		for _, name := range referencedNames(cond) {
			if _, ok := falcoOperators[name]; ok {
				continue
			}
			if isLiteral(name) {
				continue
			}
			if definedMacros[name] || definedLists[name] {
				continue
			}
			issues = append(issues, Issue{
				File: path, Path: fmt.Sprintf("[%d].condition (macro %q)", i, macroName), Code: "C8",
				Msg: fmt.Sprintf("macro %q references undefined identifier %q",
					macroName, name),
			})
		}
	}

	// Stable order for snapshot-friendly diffs.
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Code != issues[j].Code {
			return issues[i].Code < issues[j].Code
		}
		return issues[i].Path < issues[j].Path
	})

	return issues, stats, nil
}

// projectRule extracts a typed view of a rule entry and emits per-rule
// issues for missing required keys (C5).
func projectRule(file string, idx int, entry map[string]any) (*ruleEntry, []Issue) {
	rPath := fmt.Sprintf("rules[%d]", idx)
	name, _ := entry["rule"].(string)

	var issues []Issue

	requiredStringKeys := []string{"rule", "desc", "condition", "output", "priority", "source"}
	for _, key := range requiredStringKeys {
		v, ok := entry[key]
		if !ok {
			issues = append(issues, Issue{
				File: file, Path: rPath, Code: "C5",
				Msg: fmt.Sprintf("rule %q is missing required key %q", name, key),
			})
			continue
		}
		s, ok := v.(string)
		if !ok {
			issues = append(issues, Issue{
				File: file, Path: rPath + "." + key, Code: "C5",
				Msg: fmt.Sprintf("rule %q key %q is not a string (got %T)",
					name, key, v),
			})
			continue
		}
		if strings.TrimSpace(s) == "" {
			issues = append(issues, Issue{
				File: file, Path: rPath + "." + key, Code: "C5",
				Msg: fmt.Sprintf("rule %q key %q is empty", name, key),
			})
		}
	}

	tags, hasTags := entry["tags"]
	if !hasTags {
		issues = append(issues, Issue{
			File: file, Path: rPath, Code: "C5",
			Msg: fmt.Sprintf("rule %q is missing required key %q", name, "tags"),
		})
	}

	// Project into ruleEntry.
	r := &ruleEntry{
		Name:      name,
		Condition: getString(entry, "condition"),
		Output:    getString(entry, "output"),
		Priority:  getString(entry, "priority"),
		Source:    getString(entry, "source"),
		Desc:      getString(entry, "desc"),
		Path:      rPath,
	}
	if tagSlice, ok := tags.([]any); ok {
		for _, t := range tagSlice {
			if s, ok := t.(string); ok {
				r.Tags = append(r.Tags, s)
			}
		}
	}

	return r, issues
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func mapKeys(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// identifierRe captures bare identifiers in a Falco condition (a-z, A-Z, 0-9,
// '_'). Field accessors are explicitly stripped first via fieldRe, and string
// literals are stripped via stringRe to avoid matching content inside quotes.
var (
	stringRe     = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	fieldRe      = regexp.MustCompile(`claude_code\.[a-zA-Z_][a-zA-Z0-9_]*`)
	identifierRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
	numericRe    = regexp.MustCompile(`^[0-9]+$`)
)

// referencedNames returns the deduplicated set of bare identifiers in
// `condition` that are NOT Falco built-ins and NOT claude_code.* fields.
// Returned names are candidates for macro/list resolution (C8).
func referencedNames(condition string) []string {
	if condition == "" {
		return nil
	}
	// 1) strip string literals (the contents of "rm -rf /" must not be
	// interpreted as a macro reference even though the bytes look like one).
	cleaned := stringRe.ReplaceAllString(condition, " ")
	// 2) strip claude_code.* field accessors so we don't see the trailing
	// segment as an undefined macro.
	cleaned = fieldRe.ReplaceAllString(cleaned, " ")

	matches := identifierRe.FindAllString(cleaned, -1)

	seen := map[string]bool{}
	var out []string
	for _, m := range matches {
		if seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

func isLiteral(s string) bool {
	if numericRe.MatchString(s) {
		return true
	}
	switch strings.ToLower(s) {
	case "true", "false", "null":
		return true
	}
	return false
}

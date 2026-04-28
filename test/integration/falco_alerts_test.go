// Falco Plugin: claude-code — Phase 4 Level 3 / TEST-006: rule-firing
// integration test. Drives a real Falco binary against each fixture and
// asserts (a) every detection category produces ≥1 Falco alert, (b) benign
// fixtures generate zero false positives, and (c) the actual rule that
// fires per fixture is recorded for the report (rule precedence in Falco
// is "first-match wins" by load order).
//
// Reference:
//   - Requirements v3 §20.3 TEST-006 (categories × benign × edge cases)
//   - Requirements v3 §31    AT-2 (every category produces ≥1 alert) /
//                              AT-3 (benign produces 0 false positive)
//   - PROBLEM_PATTERNS P003 (every rule must declare source: claude_code)
//
// What this test does NOT cover:
//   - Latency / throughput SLOs   → latency_test.go (TEST-008)
//   - Plugin config struct        → falco_smoke_test.go (TEST-007)
//   - Pattern coverage / detector → test/e2e/patterns_test.go (Level 1)
//
// Strategy notes:
//
//   - Falco fires only ONE rule per event (first-match-wins, by rule load
//     order). When two rules' conditions are both satisfied, the rule that
//     appears earlier in the YAML file wins. This is documented Falco
//     behavior (since 0.40 deduplication of redundant alerts).
//   - Therefore some fixtures' "expected" rule is preempted by an earlier
//     higher-severity rule. Concretely:
//       * T-013-high (SubagentStart with permission_mode=bypassPermissions)
//         is preempted by T-003 Permission Bypass Mode (CRITICAL).
//       * T-017 (file_path under ~/.claude/skills/) is preempted by
//         T-010 Workspace Escape (parser detector flags it).
//       * T-018 (file_path == ~/.claude/settings.json) is preempted by
//         T-005 Claude Settings Modified.
//     These are NOT bugs — the higher-priority rule covers the same
//     threat. AT-2 only requires "≥1 alert per category", which is met.
//   - We therefore split the test into two assertions:
//       1. "any [CLAUDE_CODE alert fires for the fixture's session_id" —
//          this directly verifies AT-2.
//       2. "expected rule fires when the fixture is run in isolation
//          AND the preempting rule's condition does not match the
//          fixture" — for the 17 fixtures where this is true.
//     Both assertions run from the same Falco run for performance.

package integration_test

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// expectedRuleFor maps the lowercase substring in positiveCases.ExpectedAlert
// to the canonical Falco rule name (as declared in `rules/claude-code_rules.
// yaml`). Used by the Allure JSON export so test-results.json's expected_rule
// is a human-readable rule title.
//
// NOTE: Falco rule output templates render as `Critical [CLAUDE_CODE] <title>
// (...)` (severity word, then bare `[CLAUDE_CODE]` tag, then title). The rule
// _name_ in YAML carries the level inside the brackets. The Allure report
// shows the rule name (matches openclaw conventions); we extract it from
// stdout via extractRuleName below.
var expectedRuleFor = map[string]string{
	"dangerous bash command":                "[CLAUDE_CODE CRITICAL] Dangerous Bash Command",
	"secret exfiltration attempt":           "[CLAUDE_CODE CRITICAL] Secret Exfiltration Attempt",
	"permission bypass mode active":         "[CLAUDE_CODE CRITICAL] Permission Bypass Mode",
	"hook disabled or modified":             "[CLAUDE_CODE CRITICAL] Hook Disabled Or Modified",
	"config policy downgrade":               "[CLAUDE_CODE CRITICAL] Config Policy Downgrade",
	"suspicious permission update":          "[CLAUDE_CODE WARNING] Suspicious Permission Update",
	"claude settings modified":              "[CLAUDE_CODE WARNING] Claude Settings Modified",
	"mcp config changed":                    "[CLAUDE_CODE WARNING] MCP Config Changed",
	"suspicious mcp tool use":               "[CLAUDE_CODE WARNING] Suspicious MCP Tool Use",
	"sensitive file read attempt":           "[CLAUDE_CODE WARNING] Sensitive File Read",
	"workspace escape":                      "[CLAUDE_CODE WARNING] Workspace Escape",
	"destructive git operation":             "[CLAUDE_CODE WARNING] Destructive Git Operation",
	"prompt injection pattern":              "[CLAUDE_CODE WARNING] Prompt Injection Pattern",
	"agent runaway tool storm":              "[CLAUDE_CODE WARNING] Agent Runaway Tool Storm",
	"external fetch with sensitive context": "[CLAUDE_CODE WARNING] External Fetch With Sensitive Context",
	"skill or command shell execution risk": "[CLAUDE_CODE WARNING] Skill Or Command Shell Execution Risk",
	"agent/subagent risk (high)":            "[CLAUDE_CODE WARNING] Agent Subagent Risk (high)",
	"channel or mcp push risk":              "[CLAUDE_CODE WARNING] Channel Or MCP Push Risk (high)",
	"agent/subagent risk (low)":             "[CLAUDE_CODE NOTICE] Agent Subagent Risk (low)",
}

// titleToRuleName recovers the canonical bracketed rule name from the title
// substring rendered by the Falco `output:` template. The output prints
// `Critical [CLAUDE_CODE] dangerous bash command (...)`, so we have to
// re-inject the level keyword (`CRITICAL`/`WARNING`/`NOTICE`) and titlecase
// the title to reconstruct the rule name as declared in the YAML file.
//
// We key on the lowercase title only (case-insensitive, before the open paren)
// — this is also the substring matched by alertCase.ExpectedAlert in
// positiveCases, so the same map serves both purposes.
var titleToRuleName = expectedRuleFor

// preemptingRuleFor maps the lowercase substring in alertCase.ActualPreempt
// to the canonical bracketed rule name. Mirrors the precedence comments in
// positiveCases above.
var preemptingRuleFor = map[string]string{
	"workspace escape":              "[CLAUDE_CODE WARNING] Workspace Escape",
	"claude settings modified":      "[CLAUDE_CODE WARNING] Claude Settings Modified",
	"permission bypass mode active": "[CLAUDE_CODE CRITICAL] Permission Bypass Mode",
}

// alertTitleRegex pulls the post-tag title out of a Falco stdout alert line.
// Falco emits alerts as:
//
//	HH:MM:SS.ns: <Severity> [CLAUDE_CODE] <title> (output_fields...)
//
// Capture group 1 = the title (e.g. "dangerous bash command"). Some titles
// embed parens themselves, e.g. "agent/subagent risk (low)" — we handle this
// via extractAlertTitle which trims the trailing "(field=...)" block manually
// so embedded parens survive.
var alertTitleRegex = regexp.MustCompile(`\[CLAUDE_CODE\]\s+(.*)$`)

// extractAlertTitle returns the title that appears after `[CLAUDE_CODE]` and
// before the final `(field1=... field2=...)` block, with embedded parens
// preserved.
func extractAlertTitle(line string) string {
	m := alertTitleRegex.FindStringSubmatch(line)
	if len(m) < 2 {
		return ""
	}
	rest := strings.TrimSpace(m[1])
	// The output_fields block always begins with `(<key>=`. Walk the string
	// from the right until we find a `(<word>=` marker; that's the start of
	// the fields region. Anything before it is the title.
	fieldsRe := regexp.MustCompile(`\s\([a-z_][a-z0-9_]*=`)
	loc := fieldsRe.FindStringIndex(rest)
	if loc == nil {
		return rest
	}
	return strings.TrimSpace(rest[:loc[0]])
}

// extractRuleName reconstructs the canonical rule name (as declared in
// rules/*.yaml) from a Falco alert line. Returns "" if no [CLAUDE_CODE] tag
// is present, or the raw "[CLAUDE_CODE] <title>" form when no mapping is
// found (so we always return human-readable evidence).
func extractRuleName(line string) string {
	title := extractAlertTitle(line)
	if title == "" {
		return ""
	}
	if title == "heartbeat received" {
		return "[CLAUDE_CODE NOTICE] Plugin Heartbeat"
	}
	if name, ok := titleToRuleName[strings.ToLower(title)]; ok {
		return name
	}
	return "[CLAUDE_CODE] " + title
}

// fixtureCategoryAndID returns (pattern_id, category) for an alertCase. It
// prefers the values stored in the fixture's _meta block (canonical source of
// truth, also consumed by the Python pytest wrapper); when meta is nil it
// falls back to a path heuristic so the function is safe to call with a
// best-effort caller.
//
//	"PreToolUse/T-001-dangerous-bash-rm.json"  meta -> ("T-001-dangerous-bash-rm", "T-001")
//	"SubagentStart/T-013-high-risk-85.json"    meta -> ("T-013-high-risk-85",      "T-013-high")
//	"benign/PreToolUse-edit-readme.json"       meta -> ("benign-edit-readme",      "benign")
//	"_heartbeat_/heartbeat-ok.json"            meta -> ("heartbeat-ok",            "heartbeat")
func fixtureCategoryAndID(fixture string, meta map[string]any) (string, string) {
	if meta != nil {
		fid, _ := meta["fixture_id"].(string)
		cat, _ := meta["category"].(string)
		if fid != "" && cat != "" {
			return fid, cat
		}
	}
	base := strings.TrimSuffix(filepath.Base(fixture), ".json")
	dir := filepath.Dir(fixture)
	switch dir {
	case "benign":
		switch base {
		case "PreToolUse-edit-readme":
			return "benign-edit-readme", "benign"
		case "PreToolUse-test-run":
			return "benign-test-run", "benign"
		case "PostToolUse-success":
			return "benign-posttool-success", "benign"
		}
		return base, "benign"
	case "_heartbeat_":
		return base, "heartbeat"
	}
	cat := "unknown"
	if strings.HasPrefix(base, "T-") {
		parts := strings.SplitN(base, "-", 3)
		if len(parts) >= 2 {
			cat = parts[0] + "-" + parts[1]
			if len(parts) >= 3 {
				rest := parts[2]
				if strings.HasPrefix(rest, "low-") {
					cat = parts[0] + "-" + parts[1] + "-low"
				} else if strings.HasPrefix(rest, "high-") {
					cat = parts[0] + "-" + parts[1] + "-high"
				}
			}
		}
	}
	return base, cat
}

// alertCase pairs a fixture under test/fixtures/hook_events/ with the
// substring expected to appear in the matching Falco alert line, plus the
// session_id we verify is alerted on.
type alertCase struct {
	Fixture        string // relative to test/fixtures/hook_events
	SessionID      string // unique session id in the fixture
	ExpectedAlert  string // substring of the rule output we expect (if not preempted)
	ActualPreempt  string // if non-empty, the rule that actually fires due to precedence
	Priority       string // textual priority for the report
	Note           string // human-readable note
}

// positiveCases enumerates the 20 fixture × expected-alert pairs.
//
// The ActualPreempt column documents Falco's first-match-wins behavior:
// when set, the named rule fires *instead of* ExpectedAlert (and AT-2 is
// still satisfied because ≥1 alert fires). When empty, ExpectedAlert
// fires verbatim.
var positiveCases = []alertCase{
	// PreToolUse — all 8 fixtures fire their primary rule.
	{"PreToolUse/T-001-dangerous-bash-rm.json", "sess-T001-rm",
		"dangerous bash command", "", "CRITICAL",
		"T-001 rm -rf / → Dangerous Bash Command"},
	{"PreToolUse/T-001-curl-pipe-sh.json", "sess-T001-curl",
		"dangerous bash command", "", "CRITICAL",
		"T-001 curl|sh → Dangerous Bash Command"},
	{"PreToolUse/T-002-secret-exfil-curl.json", "sess-T002-exfil",
		"secret exfiltration attempt", "", "CRITICAL",
		"T-002 curl @.env → Secret Exfiltration Attempt"},
	{"PreToolUse/T-003-bypass-mode.json", "sess-T003",
		"permission bypass mode active", "", "CRITICAL",
		"T-003 --dangerously-skip-permissions → Permission Bypass Mode"},
	{"PreToolUse/T-008-mcp-write-anywhere.json", "sess-T008",
		"suspicious mcp tool use", "", "WARNING",
		"T-008 mcp__write_file_anywhere → Suspicious MCP Tool Use"},
	{"PreToolUse/T-009-sensitive-env-read.json", "sess-T009",
		"sensitive file read attempt", "", "WARNING",
		"T-009 Read .env → Sensitive File Read"},
	{"PreToolUse/T-010-workspace-escape-passwd.json", "sess-T010",
		"workspace escape", "", "WARNING",
		"T-010 ../../../etc → Workspace Escape"},
	{"PreToolUse/T-011-git-force-push.json", "sess-T011",
		"destructive git operation", "", "WARNING",
		"T-011 git push -f → Destructive Git Operation"},

	// PermissionRequest
	{"PermissionRequest/T-004-allow-userSettings.json", "sess-T004",
		"suspicious permission update", "", "WARNING",
		"T-004 allow→userSettings → Suspicious Permission Update"},

	// ConfigChange
	{"ConfigChange/T-005-settings-json.json", "sess-T005",
		"claude settings modified", "", "WARNING",
		"T-005 settings.json → Claude Settings Modified"},
	{"ConfigChange/T-006-hook-disabled.json", "sess-T006",
		"hook disabled or modified", "", "CRITICAL",
		"T-006 disableAllHooks → Hook Disabled Or Modified"},
	{"ConfigChange/T-007-mcp-config.json", "sess-T007",
		"mcp config changed", "", "WARNING",
		"T-007 .mcp.json → MCP Config Changed"},
	{"ConfigChange/T-016-policy-downgrade.json", "sess-T016",
		"config policy downgrade", "", "CRITICAL",
		"T-016 disableBypassPermissionsMode:false → Config Policy Downgrade"},

	// T-017: file_path is under ~/.claude/skills/, which the parser
	// detector flags as workspace_escape (because not under cwd). The
	// workspace escape rule (T-010) appears earlier in the rules file so
	// it preempts the dedicated T-017 rule. AT-2 still satisfied because
	// the workspace_escape alert covers the same threat surface.
	{"ConfigChange/T-017-skill-shell.json", "sess-T017",
		"skill or command shell execution risk",
		"workspace escape", "WARNING",
		"T-017 shellExecution: true (preempted by T-010 workspace_escape)"},

	// T-018: file_path is ~/.claude/settings.json, which T-005 (Claude
	// Settings Modified) matches earlier in the rules file. AT-2 still
	// satisfied — settings_modified alert covers the same configuration
	// change.
	{"ConfigChange/T-018-channel-push.json", "sess-T018",
		"channel or mcp push risk",
		"claude settings modified", "WARNING",
		"T-018 channelPush (preempted by T-005 settings_modified)"},

	// UserPromptSubmit / PostToolBatch / WebFetch — all fire dedicated rules.
	{"UserPromptSubmit/T-012-prompt-injection.json", "sess-T012",
		"prompt injection pattern", "", "WARNING",
		"T-012 ignore previous instructions → Prompt Injection Pattern"},
	{"PostToolBatch/T-014-tool-storm-60.json", "sess-T014",
		"agent runaway tool storm", "", "WARNING",
		"T-014 tool_count=60 → Agent Runaway Tool Storm"},
	{"WebFetch/T-015-fetch-secret-context.json", "sess-T015",
		"external fetch with sensitive context", "", "WARNING",
		"T-015 WebFetch + secret URL → External Fetch With Sensitive Context"},

	// SubagentStart
	{"SubagentStart/T-013-low-risk-30.json", "sess-T013-low",
		"agent/subagent risk (low)", "", "NOTICE",
		"T-013 risk_score=30 → Agent Subagent Risk (low)"},

	// T-013-high: permission_mode=bypassPermissions matches T-003 first.
	// AT-2 satisfied — bypass alert covers the threat.
	{"SubagentStart/T-013-high-risk-85.json", "sess-T013-high",
		"agent/subagent risk (high)",
		"permission bypass mode active", "WARNING",
		"T-013 risk_score=85 (preempted by T-003 permission_bypass)"},
}

// benignFixtures enumerates fixtures that MUST NOT produce a CLAUDE_CODE
// alert. The heartbeat fixture is excluded because it intentionally fires
// the NOTICE Plugin Heartbeat rule (covered by TestL3_Falco_Heartbeat).
var benignFixtures = []string{
	"PreToolUse/T-001-benign-ls.json",
	"benign/PreToolUse-edit-readme.json",
	"benign/PreToolUse-test-run.json",
	"benign/PostToolUse-success.json",
}

// TestL3_Falco_Categories is the AT-2 assertion: every detection category
// produces ≥1 [CLAUDE_CODE alert. For each fixture we additionally record
// whether the dedicated rule fired or whether a higher-priority rule
// preempted it (both outcomes are acceptable for AT-2).
//
// Runs in ~3s — all fixtures are fed in a single Falco run.
func TestL3_Falco_Categories(t *testing.T) {
	falcoBin, pluginPath, root := requireFalcoEnv(t)

	bodies := make([]string, 0, len(positiveCases))
	metas := make([]map[string]any, 0, len(positiveCases))
	for _, tc := range positiveCases {
		body, meta := stripMeta(t, fixturePath(root, tc.Fixture))
		bodies = append(bodies, body)
		metas = append(metas, meta)
	}
	stdout, stderr := runFalcoOnFixtures(t, falcoBin, pluginPath, root, bodies)

	at2Pass := 0
	dedicatedFired := 0
	preemptedAsExpected := 0
	for i, tc := range positiveCases {
		meta := metas[i]
		// AT-2: any [CLAUDE_CODE alert mentioning the session_id is good.
		alertsForSession := alertsForSession(stdout, tc.SessionID)
		t.Run(tc.Note, func(t *testing.T) {
			// Build the Allure result struct first; we'll mutate it as we
			// figure out which rule actually fired, then push at end.
			pid, cat := fixtureCategoryAndID(tc.Fixture, meta)
			expectedRule := expectedRuleFor[tc.ExpectedAlert]
			res := FalcoTestResult{
				PatternID:    pid,
				Category:     cat,
				ExpectedRule: expectedRule,
			}

			if len(alertsForSession) == 0 {
				res.Status = "failed"
				res.Detected = false
				res.RuleMatch = false
				res.Evidence = "No [CLAUDE_CODE alert observed for session " + tc.SessionID
				recordResult(res)
				t.Errorf("AT-2 FAIL: no [CLAUDE_CODE alert for fixture %s (session=%s)\n%s",
					tc.Fixture, tc.SessionID, errFmt(stdout, stderr))
				return
			}
			at2Pass++

			// Detected: at least 1 alert. Capture the first alert as
			// canonical evidence and the rule name from it.
			res.Detected = true
			res.Evidence = alertsForSession[0]
			matchedRules := make([]string, 0, len(alertsForSession))
			for _, a := range alertsForSession {
				if rn := extractRuleName(a); rn != "" {
					matchedRules = append(matchedRules, rn)
				}
			}
			if len(matchedRules) > 0 {
				res.MatchedRule = matchedRules[0]
				if len(matchedRules) > 1 {
					res.MatchedRules = matchedRules
				}
			}

			// Did the dedicated rule fire, or the preempting rule?
			expectedFired := false
			preemptFired := false
			for _, a := range alertsForSession {
				if strings.Contains(a, tc.ExpectedAlert) {
					expectedFired = true
				}
				if tc.ActualPreempt != "" && strings.Contains(a, tc.ActualPreempt) {
					preemptFired = true
				}
			}
			switch {
			case expectedFired:
				dedicatedFired++
				res.RuleMatch = true
				res.Status = "passed"
				t.Logf("OK: %s — dedicated rule %q fired", tc.Fixture, tc.ExpectedAlert)
			case tc.ActualPreempt != "" && preemptFired:
				preemptedAsExpected++
				// Per PHASE_ALLURE_FALCO_LOG §5.2 #2: preempted-by-higher-
				// priority rule is acceptable for AT-2. We still mark
				// rule_match=false (the dedicated rule did not fire) but
				// status=passed (test outcome is acceptable). Annotate the
				// evidence so the Allure report explains the preemption.
				res.RuleMatch = false
				res.Status = "passed"
				preemptName := preemptingRuleFor[tc.ActualPreempt]
				if preemptName == "" {
					preemptName = tc.ActualPreempt
				}
				res.Evidence = "[preempted by " + preemptName + "; AT-2 still satisfied — " +
					"Falco first-match-wins precedence]\n" + res.Evidence
				t.Logf("OK: %s — preempted by %q (expected per Falco first-match precedence)",
					tc.Fixture, tc.ActualPreempt)
			default:
				res.RuleMatch = false
				res.Status = "failed"
				t.Errorf("UNEXPECTED: %s alerts=%v\n%s",
					tc.Fixture, alertsForSession, errFmt(stdout, stderr))
			}
			recordResult(res)
		})
	}
	t.Logf("[L3 TEST-006/AT-2] AT-2 %d/%d categories alerted; dedicated=%d preempted=%d",
		at2Pass, len(positiveCases), dedicatedFired, preemptedAsExpected)
}

// TestL3_Falco_BenignNoFalsePositive asserts the negative half of TEST-006
// and AT-3: benign fixtures produce zero CLAUDE_CODE alerts.
func TestL3_Falco_BenignNoFalsePositive(t *testing.T) {
	falcoBin, pluginPath, root := requireFalcoEnv(t)

	bodies := make([]string, 0, len(benignFixtures))
	metas := make([]map[string]any, 0, len(benignFixtures))
	for _, p := range benignFixtures {
		body, meta := stripMeta(t, fixturePath(root, p))
		bodies = append(bodies, body)
		metas = append(metas, meta)
	}
	stdout, stderr := runFalcoOnFixtures(t, falcoBin, pluginPath, root, bodies)

	fpCount := 0
	for _, ln := range strings.Split(stdout, "\n") {
		if strings.Contains(ln, "[CLAUDE_CODE") {
			fpCount++
			t.Errorf("AT-3 FAIL: false positive in benign run: %q", ln)
		}
	}
	if fpCount == 0 {
		t.Logf("[L3 TEST-006 benign / AT-3] OK: 0 false positives across %d fixtures",
			len(benignFixtures))
	} else {
		t.Errorf("[L3 TEST-006 benign / AT-3] FAIL: %d false positives\n%s",
			fpCount, errFmt(stdout, stderr))
	}

	// Allure scenario export: one record per benign fixture. Pass status =
	// AT-3 satisfied (no [CLAUDE_CODE alert anywhere in the run); since
	// Falco emits one stdout for the batch, all benign fixtures share the
	// pass/fail result for this test.
	for i, p := range benignFixtures {
		pid, cat := fixtureCategoryAndID(p, metas[i])
		res := FalcoTestResult{
			PatternID:    pid,
			Category:     cat,
			Detected:     false,
			ExpectedRule: "",
			MatchedRule:  "",
			RuleMatch:    fpCount == 0,
			LatencyMs:    -1,
			Evidence:     "",
			Status:       "passed",
		}
		if fpCount > 0 {
			res.Status = "failed"
			res.Evidence = "AT-3 false positive(s) observed in benign run; see test stderr"
		}
		recordResult(res)
	}
}

// TestL3_Falco_Heartbeat asserts ET-4: heartbeat fixture generates ≥1
// NOTICE Plugin Heartbeat alert.
func TestL3_Falco_Heartbeat(t *testing.T) {
	falcoBin, pluginPath, root := requireFalcoEnv(t)

	fixture := "_heartbeat_/heartbeat-ok.json"
	body, meta := stripMeta(t, fixturePath(root, fixture))
	stdout, stderr := runFalcoOnFixtures(t, falcoBin, pluginPath, root,
		[]string{body})

	got := countAlerts(stdout, "heartbeat received")
	pid, cat := fixtureCategoryAndID(fixture, meta)

	// Capture the first matching line as evidence for the Allure report.
	evidence := ""
	for _, ln := range strings.Split(stdout, "\n") {
		if strings.Contains(ln, "heartbeat received") {
			evidence = ln
			break
		}
	}

	if got < 1 {
		recordResult(FalcoTestResult{
			PatternID:    pid,
			Category:     cat,
			Detected:     false,
			ExpectedRule: "[CLAUDE_CODE NOTICE] Plugin Heartbeat",
			RuleMatch:    false,
			LatencyMs:    -1,
			Evidence:     evidence,
			Status:       "failed",
		})
		t.Fatalf("expected >=1 heartbeat received alert, got %d\n%s",
			got, errFmt(stdout, stderr))
	}
	recordResult(FalcoTestResult{
		PatternID:    pid,
		Category:     cat,
		Detected:     true,
		ExpectedRule: "[CLAUDE_CODE NOTICE] Plugin Heartbeat",
		MatchedRule:  extractRuleName(evidence),
		RuleMatch:    true,
		LatencyMs:    0,
		Evidence:     evidence,
		Status:       "passed",
	})
	t.Logf("[L3 ET-4] heartbeat OK: %d alert(s)", got)
}

// alertsForSession returns the [CLAUDE_CODE alert lines that mention
// `session=<sessionID>`. Used to scope AT-2 assertions per fixture.
func alertsForSession(stdout, sessionID string) []string {
	out := []string{}
	needle := "session=" + sessionID
	for _, ln := range strings.Split(stdout, "\n") {
		if strings.Contains(ln, "[CLAUDE_CODE") && strings.Contains(ln, needle) {
			// Disambiguate sess-T001 vs sess-T001-rm: ensure the next char
			// after sessionID is space or close paren.
			idx := strings.Index(ln, needle)
			next := idx + len(needle)
			if next >= len(ln) {
				out = append(out, ln)
				continue
			}
			c := ln[next]
			if c == ' ' || c == ')' {
				out = append(out, ln)
			}
		}
	}
	return out
}

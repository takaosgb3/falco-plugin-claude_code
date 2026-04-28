// Falco Plugin: claude-code — Level 1 (Pattern Coverage) Tests
//
// These tests load every JSON fixture under `test/fixtures/hook_events/`,
// feed each through `pkg/parser.Parser.Parse()`, and assert on the
// expectation block embedded in the fixture's `_meta` object.
//
// Reference:
//   - Requirements v3 §20.1 (3-layer E2E)
//   - Requirements v3 §20.2 (fixture layout)
//   - Requirements v3 §20.3 TEST-002 / TEST-003 (pattern coverage / benign)
//   - Requirements v3 §10.1 / §10.2 (event schema / Falco fields)
//   - PROBLEM_PATTERNS P004 (nil map), P010 (Fields/Extract), P020 (truncate)
//
// What we cover here:
//   - 9 parser-classified categories (T-002, T-003, T-004, T-005, T-006,
//     T-007, T-009, T-010, T-011) → expected_risk_type assertion.
//   - 9 rule-only categories (T-001, T-008, T-012, T-013, T-014, T-015,
//     T-016, T-017, T-018) → no parser risk_type expected (or, for fixtures
//     where the detector secondarily labels the event, the expected
//     workspace_escape side-effect is documented).
//   - benign fixtures: parser risk_type stays "none" / empty.
//   - redaction smoke: every parser-detected fixture must round-trip
//     `redaction_status` ∈ {"none","redacted","truncated"} (per §17.1).
//
// What we do NOT cover here (deferred to Level 3 / Falco integration):
//   - Falco rule firing — that's Level 3 and requires a real Falco binary.
//   - Latency SLO measurement — that's TEST-008 / §20.3.1.
//   - Rotation scenario — that's Level 2 / TestPipeline_Rotation_*.
package e2e_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/takaosgb3/falco-plugin-claude_code/pkg/parser"
	"github.com/takaosgb3/falco-plugin-claude_code/pkg/testutil"
)

// repoRoot returns the path to the repository root from this package.
// `test/e2e/patterns_test.go` is two levels deep, so go up twice.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return root
}

// newTestParser returns a parser configured the same way the production
// plugin configures itself (Init() in cmd/plugin-sdk/plugin.go). Keeping
// this in lockstep with production wiring is critical: a Level 1 test that
// configures the parser differently would not be representative.
func newTestParser() *parser.Parser {
	return parser.New(parser.Config{
		LogFormat:        "json",
		SecurityPatterns: true,
	})
}

// TestPattern_AllFixturesParse asserts that every fixture is structurally
// valid JSONL that the parser accepts (P-001 / §14.1). This is the cheapest
// gate against typos and missing required fields. ParserPanics on a fixture
// would show up here as a failed Parse() with non-nil error.
func TestPattern_AllFixturesParse(t *testing.T) {
	root := testutil.FixturesRoot(repoRoot(t))
	cases := testutil.LoadAllFixtures(t, root)
	require.NotEmpty(t, cases, "no fixtures found under %s", root)

	require.GreaterOrEqual(t, len(cases), 20,
		"requirement: at least 20 fixtures (got %d)", len(cases))

	p := newTestParser()
	for _, c := range cases {
		c := c
		t.Run(c.ShortName(), func(t *testing.T) {
			entry, err := p.Parse(c.JSONLine)
			require.NoError(t, err, "fixture %s failed to parse: %v", c.Path, err)
			require.NotNil(t, entry)

			// P004: Headers map must be non-nil for GOB safety. The parser
			// guarantees this; we double-check on every fixture.
			require.NotNil(t, entry.Headers, "Headers map must be initialized (P004)")

			// Expectation cross-check (only when the fixture pinned event_name).
			if c.Meta.ExpectedEventName != "" {
				assert.Equal(t, c.Meta.ExpectedEventName, entry.EventName,
					"event_name mismatch for %s", c.Path)
			}
		})
	}
}

// TestPattern_ParserCategories drives the 9 parser-classified categories
// (T-002, T-003, T-004, T-005, T-006, T-007, T-009, T-010, T-011 plus the
// secondary workspace_escape side-effects documented in T-008/T-016/T-017
// fixtures). For each fixture whose `_meta.expected_risk_type` is non-empty,
// we assert risk_type / risk_score / severity match what the detector
// produced.
func TestPattern_ParserCategories(t *testing.T) {
	root := testutil.FixturesRoot(repoRoot(t))
	cases := testutil.LoadAllFixtures(t, root)
	p := newTestParser()

	covered := make(map[string]bool)

	for _, c := range cases {
		c := c
		if c.Meta.ExpectedRiskType == "" {
			continue
		}
		t.Run(c.ShortName(), func(t *testing.T) {
			entry, err := p.Parse(c.JSONLine)
			require.NoError(t, err)

			assert.Equal(t, c.Meta.ExpectedRiskType, entry.RiskType,
				"risk_type mismatch for %s; notes=%q", c.Path, c.Meta.Notes)

			if c.Meta.ExpectedRiskScoreMin > 0 {
				assert.GreaterOrEqual(t, entry.RiskScore, c.Meta.ExpectedRiskScoreMin,
					"risk_score below floor for %s", c.Path)
			}
			if c.Meta.ExpectedSeverity != "" {
				assert.Equal(t, c.Meta.ExpectedSeverity, entry.Severity,
					"severity mismatch for %s", c.Path)
			}
			covered[c.Meta.ExpectedRiskType] = true
		})
	}

	// Sanity: at least 5 distinct parser-classified risk_types must be
	// represented. Hard requirement is the 9 detector categories; we set 5
	// as the lower bound so this test fails loudly even if half the fixtures
	// are accidentally deleted.
	require.GreaterOrEqual(t, len(covered), 5,
		"too few parser-classified risk_types covered: %v", covered)

	// Each of the canonical 9 detector categories *should* appear at least
	// once across the fixture set. We assert the 9 here (failing loudly if
	// any goes missing during fixture refactors).
	expectedCategories := []string{
		"secret_exfiltration", // T-002
		"permission_bypass",   // T-003
		"permission_update",   // T-004
		"settings_modified",   // T-005
		"hook_disabled",       // T-006
		"mcp_config_changed",  // T-007
		"sensitive_file_read", // T-009
		"workspace_escape",    // T-010 (also T-008/T-016/T-017 as side effect)
		"git_destructive",     // T-011
	}
	for _, cat := range expectedCategories {
		assert.True(t, covered[cat],
			"detector category %q has no covering fixture", cat)
	}
}

// TestPattern_BenignNoFalsePositive asserts that every fixture flagged as
// benign produces no parser-side classification. The parser's internal
// detector must not assign risk_type to benign events (TEST-003).
func TestPattern_BenignNoFalsePositive(t *testing.T) {
	root := testutil.FixturesRoot(repoRoot(t))
	cases := testutil.LoadFixturesForCategory(t, root, "benign")
	require.GreaterOrEqual(t, len(cases), 3, "expect at least 3 benign fixtures")

	p := newTestParser()
	for _, c := range cases {
		c := c
		t.Run(c.ShortName(), func(t *testing.T) {
			entry, err := p.Parse(c.JSONLine)
			require.NoError(t, err)

			// Detector must NOT have assigned a risk_type. The upstream
			// logger emits "none" by default; the detector promotes that
			// only when it actually sees a pattern.
			assert.Contains(t, []string{"", "none"}, entry.RiskType,
				"benign fixture %s got unexpected risk_type=%q",
				c.Path, entry.RiskType)
			assert.Less(t, entry.RiskScore, uint64(50),
				"benign fixture %s got high risk_score=%d (TEST-003)",
				c.Path, entry.RiskScore)
		})
	}
}

// TestPattern_RuleOnlyCategoriesPreserved confirms that for the rule-side
// categories the parser passes through the upstream-supplied risk fields
// without dropping or corrupting them. This is the contract the Falco rule
// engine relies on: rule-side categories cannot be detected by the parser
// alone, so the rule needs to evaluate the original `claude_code.*` fields.
func TestPattern_RuleOnlyCategoriesPreserved(t *testing.T) {
	root := testutil.FixturesRoot(repoRoot(t))
	cases := testutil.LoadAllFixtures(t, root)
	p := newTestParser()

	// fixtureID → the field the corresponding Falco rule keys off of.
	// We only assert preservation of those fields.
	ruleSideExpectations := map[string]struct {
		mustHaveField string
		mustHaveValue string // substring check (icontains semantics)
	}{
		"T-001-dangerous-bash-rm":      {"command", "rm -rf /"},
		"T-001-curl-pipe-sh":           {"command", "| sh"},
		"T-008-mcp-write-anywhere":     {"tool_name", "mcp__write_file_anywhere"},
		"T-012-prompt-injection":       {"evidence", "ignore previous instructions"},
		"T-014-tool-storm-60":          {"tool_count", ""}, // numeric — special-cased
		"T-015-fetch-secret-context":   {"url", "secret"},
		"T-016-policy-downgrade":       {"evidence", "disableBypassPermissionsMode"},
		"T-017-skill-shell":            {"evidence", "shellExecution: true"},
		"T-018-channel-push":           {"evidence", "channelPush"},
		"T-013-low-risk-30":            {"risk_score", ""},  // numeric
		"T-013-high-risk-85":           {"permission_mode", "bypassPermissions"},
	}

	for _, c := range cases {
		c := c
		expect, ok := ruleSideExpectations[c.Meta.FixtureID]
		if !ok {
			continue
		}
		t.Run(c.ShortName(), func(t *testing.T) {
			entry, err := p.Parse(c.JSONLine)
			require.NoError(t, err)

			switch expect.mustHaveField {
			case "command":
				assert.Contains(t, entry.Command, expect.mustHaveValue,
					"rule key field 'command' must preserve %q", expect.mustHaveValue)
			case "tool_name":
				assert.Equal(t, expect.mustHaveValue, entry.ToolName,
					"rule key field 'tool_name' must preserve")
			case "evidence":
				assert.Contains(t, entry.Evidence, expect.mustHaveValue,
					"rule key field 'evidence' must preserve %q", expect.mustHaveValue)
			case "url":
				assert.Contains(t, entry.URL, expect.mustHaveValue,
					"rule key field 'url' must preserve %q", expect.mustHaveValue)
			case "tool_count":
				assert.GreaterOrEqual(t, entry.ToolCount, uint64(50),
					"tool_count must clear T-014 storm threshold (>=50)")
			case "risk_score":
				assert.Greater(t, entry.RiskScore, uint64(0),
					"upstream-supplied risk_score must round-trip")
			case "permission_mode":
				assert.Equal(t, expect.mustHaveValue, entry.PermissionMode,
					"permission_mode must round-trip")
			default:
				t.Fatalf("unhandled rule-side expectation %q", expect.mustHaveField)
			}
		})
	}
}

// TestPattern_RedactionSmoke ensures the redaction layer (§17.1) does not
// blow up on any fixture. We don't assert that any single fixture is
// redacted — that's the unit test's job — but we do require that the
// redaction_status remains a valid enum value end to end.
func TestPattern_RedactionSmoke(t *testing.T) {
	root := testutil.FixturesRoot(repoRoot(t))
	cases := testutil.LoadAllFixtures(t, root)
	p := newTestParser()

	validStatus := map[string]bool{
		"":          true, // logger may legitimately leave empty if it wrote nothing
		"none":      true,
		"redacted":  true,
		"truncated": true,
		"error":     true,
	}

	for _, c := range cases {
		c := c
		t.Run(c.ShortName(), func(t *testing.T) {
			entry, err := p.Parse(c.JSONLine)
			require.NoError(t, err)

			assert.True(t, validStatus[entry.RedactionStatus],
				"unknown redaction_status %q for %s", entry.RedactionStatus, c.Path)

			// TEST-004: verify no raw secret-shaped tokens slip through.
			// We grep the parser-output strings for the example secret
			// markers we use in fixtures (`AKIAEXAMPLE`, `xoxb-EXAMPLE`,
			// `ghp_EXAMPLE`). None of our fixtures contain those today
			// (we use synthetic placeholders), so this is a proactive
			// guard against future fixtures leaking real-shaped tokens.
			for _, leakable := range []string{
				"AKIA0123456789012345",      // 20-char real-shape AWS key
				"xoxb-real-token-value-123", // real-shape Slack
				"ghp_realgithubpattokenvaluexxxxxxxxxxxxxxxx", // real-shape GitHub PAT
			} {
				assert.NotContains(t, entry.Command+entry.Evidence+entry.URL+entry.RawExcerpt,
					leakable,
					"fixture %s leaked real-shaped secret %q", c.Path, leakable)
			}
		})
	}
}

// TestPattern_FixtureSchemaSanity verifies every fixture conforms to the
// minimum §10.1 / §10.3 schema requirements: schema_version family prefix,
// presence of session_id, hook_event_name, and a timestamp field.
//
// The parser already enforces this via validateRequired(); here we verify
// the *fixtures themselves* are healthy so a Level 1 test failure points to
// a schema bug rather than a fixture bug.
func TestPattern_FixtureSchemaSanity(t *testing.T) {
	root := testutil.FixturesRoot(repoRoot(t))
	cases := testutil.LoadAllFixtures(t, root)

	for _, c := range cases {
		c := c
		t.Run(c.ShortName(), func(t *testing.T) {
			line := c.JSONLine
			require.True(t,
				strings.Contains(line, `"schema_version":"claude_code_security_event/v1"`),
				"fixture %s missing schema_version=claude_code_security_event/v1", c.Path)
			require.True(t,
				strings.Contains(line, `"session_id":`) && !strings.Contains(line, `"session_id":""`),
				"fixture %s missing or empty session_id", c.Path)
			require.True(t,
				strings.Contains(line, `"hook_event_name":`),
				"fixture %s missing hook_event_name", c.Path)
			require.True(t,
				strings.Contains(line, `"received_at":`),
				"fixture %s missing received_at", c.Path)
		})
	}
}

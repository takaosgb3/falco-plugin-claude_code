// Falco Plugin: claude-code — Phase ALLURE-FALCO
//
// allure_result_test.go: shared types for openclaw-compatible Allure
// scenario JSON. Available unconditionally so the on/off recorder pair can
// reference the same struct; this file has NO `_test.go`-package side effects
// other than the type declaration.
//
// Schema is intentionally identical to:
//
//	/Users/takaos/lab/falco-plugin-openclaw/e2e/results/test-results.json
//
// 9 keys: pattern_id / category / detected / expected_rule / matched_rule /
// rule_match / latency_ms / evidence / status. Optional matched_rules slice
// (omitempty) is supported for future multi-match scenarios.

package integration_test

// FalcoTestResult mirrors the openclaw scenario record. One per fixture run.
//
// JSON tags are required: the Python pytest wrapper consumes
// pattern_id/category/detected/etc. from this exact key set.
type FalcoTestResult struct {
	PatternID    string   `json:"pattern_id"`
	Category     string   `json:"category"`
	Detected     bool     `json:"detected"`
	ExpectedRule string   `json:"expected_rule"`
	MatchedRule  string   `json:"matched_rule"`
	MatchedRules []string `json:"matched_rules,omitempty"`
	RuleMatch    bool     `json:"rule_match"`
	LatencyMs    int64    `json:"latency_ms"`
	Evidence     string   `json:"evidence"`
	Status       string   `json:"status"` // "passed" | "failed"
}

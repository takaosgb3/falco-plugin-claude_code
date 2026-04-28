// Package testutil provides shared helpers for loading hook event fixtures and
// building expectation tables used by the Phase 4 Level 1 (pattern) and
// Level 2 (pipeline) tests.
//
// Fixtures are JSON files under `test/fixtures/hook_events/<event_name>/`.
// Each fixture is a single normalized claude_code event (per requirements
// v3 §10.1) plus an embedded `_meta` object that records the test
// expectation. The `_meta` object is stripped before the fixture is fed to
// the parser, so production code never sees it.
package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// ExpectedMeta is the contract a fixture's `_meta` block must follow.
// Fields are optional; an empty string / zero value means "no expectation".
type ExpectedMeta struct {
	FixtureID           string `json:"fixture_id"`
	Category            string `json:"category"`            // e.g. "T-001", "benign", "heartbeat"
	ExpectedDetection   string `json:"expected_detection"`  // "parser", "rule", "none"
	ExpectedRiskType    string `json:"expected_risk_type"`  // empty = no expectation
	ExpectedRiskScoreMin uint64 `json:"expected_risk_score_min"`
	ExpectedSeverity    string `json:"expected_severity"`
	ExpectedEventName   string `json:"expected_event_name"`
	Notes               string `json:"notes"`
}

// FixtureCase pairs a single fixture's expectation with the JSONL line that
// the parser should consume.
type FixtureCase struct {
	Path     string       // absolute path on disk
	Meta     ExpectedMeta // expectation block
	JSONLine string       // _meta-stripped, JSONL-ready line (no trailing newline)
}

// FixturesRoot returns the absolute path to the hook_events fixture root.
// Callers pass the test workdir's depth via `relRoot` so the helper works
// from any test package. For tests under cmd/plugin-sdk and test/e2e the
// caller passes the relative prefix that walks up to the repo root.
func FixturesRoot(repoRoot string) string {
	return filepath.Join(repoRoot, "test", "fixtures", "hook_events")
}

// LoadFixture loads a single fixture file and strips its `_meta` block.
// Returns (case, error). On any failure (read / parse / shape) the error is
// returned; callers typically fail the test with t.Fatal.
func LoadFixture(path string) (FixtureCase, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return FixtureCase{}, fmt.Errorf("read %s: %w", path, err)
	}

	var generic map[string]json.RawMessage
	if err := json.Unmarshal(raw, &generic); err != nil {
		return FixtureCase{}, fmt.Errorf("parse %s: %w", path, err)
	}

	var meta ExpectedMeta
	if metaRaw, ok := generic["_meta"]; ok {
		if err := json.Unmarshal(metaRaw, &meta); err != nil {
			return FixtureCase{}, fmt.Errorf("parse _meta in %s: %w", path, err)
		}
		delete(generic, "_meta")
	}

	cleaned, err := json.Marshal(generic)
	if err != nil {
		return FixtureCase{}, fmt.Errorf("re-marshal %s: %w", path, err)
	}

	return FixtureCase{
		Path:     path,
		Meta:     meta,
		JSONLine: string(cleaned),
	}, nil
}

// MustLoadFixture is the t.Helper variant of LoadFixture.
func MustLoadFixture(t *testing.T, path string) FixtureCase {
	t.Helper()
	c, err := LoadFixture(path)
	if err != nil {
		t.Fatalf("LoadFixture: %v", err)
	}
	return c
}

// LoadAllFixtures walks `root` and returns every fixture found, sorted by
// path. Non-`*.json` files are skipped silently (test/fixtures may also
// contain `.txt` malformed-line fixtures in the future).
func LoadAllFixtures(t *testing.T, root string) []FixtureCase {
	t.Helper()
	var cases []FixtureCase
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}
		c, err := LoadFixture(path)
		if err != nil {
			return err
		}
		cases = append(cases, c)
		return nil
	})
	if err != nil {
		t.Fatalf("LoadAllFixtures: %v", err)
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].Path < cases[j].Path })
	return cases
}

// LoadFixturesForCategory returns fixtures whose `_meta.category` matches
// `category` exactly. Useful for category-level subtests.
func LoadFixturesForCategory(t *testing.T, root, category string) []FixtureCase {
	t.Helper()
	all := LoadAllFixtures(t, root)
	out := make([]FixtureCase, 0, len(all))
	for _, c := range all {
		if c.Meta.Category == category {
			out = append(out, c)
		}
	}
	return out
}

// ShortName returns a stable t.Run-friendly short name for a fixture.
// Format: "<category>:<basename-without-ext>".
func (c FixtureCase) ShortName() string {
	base := filepath.Base(c.Path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if c.Meta.Category != "" {
		return c.Meta.Category + ":" + base
	}
	return base
}

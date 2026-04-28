// Falco Plugin: claude-code — Phase ALLURE-FALCO
//
// allure_export_on_test.go: build-tagged recorder. Activated by:
//
//	go test -tags=allure -count=1 ./test/integration/... -run TestL3_Falco_Categories
//
// Behavior:
//   - integration tests call recordResult() to push a FalcoTestResult into an
//     in-memory slice;
//   - TestMain (this file) flushes the slice to
//     test/integration/results/test-results.json on exit, in a schema
//     compatible with /Users/takaos/lab/falco-plugin-openclaw/e2e/results/
//     test-results.json (9 keys per record).
//
// This file is excluded from the default build (no `allure` tag) so the
// integration suite has zero filesystem side-effects under `go test ./...`.

//go:build allure

package integration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
)

// resultsMu guards results against parallel sub-tests in TestL3_Falco_*.
var (
	resultsMu sync.Mutex
	results   = make([]FalcoTestResult, 0, 64)
)

// recordResult appends a single test outcome. Safe to call from t.Run
// goroutines.
func recordResult(r FalcoTestResult) {
	resultsMu.Lock()
	defer resultsMu.Unlock()
	results = append(results, r)
}

// flushResults writes the accumulated results to
// test/integration/results/test-results.json with deterministic ordering
// (alphabetical by pattern_id) so reruns produce stable diffs.
func flushResults() error {
	resultsMu.Lock()
	defer resultsMu.Unlock()

	if len(results) == 0 {
		return nil
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].PatternID < results[j].PatternID
	})

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	outDir := filepath.Join(root, "test", "integration", "results")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outDir, err)
	}
	outPath := filepath.Join(outDir, "test-results.json")

	body, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}
	if err := os.WriteFile(outPath, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	fmt.Fprintf(os.Stdout, "[allure-export] wrote %d results to %s\n",
		len(results), outPath)
	return nil
}

// TestMain runs after all tests; flushes JSON before reporting the exit code
// so a single `go test -tags=allure ./test/integration/...` produces both the
// PASS/FAIL summary AND test-results.json.
func TestMain(m *testing.M) {
	code := m.Run()
	if err := flushResults(); err != nil {
		fmt.Fprintf(os.Stderr, "[allure-export] flush error: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

// Falco Plugin: claude-code — Phase ALLURE-FALCO
//
// allure_export_off_test.go: no-op recorder when the `allure` build tag is
// NOT set. The integration tests call recordResult() unconditionally; in the
// default build this resolves to a single map-store-and-discard without any
// disk I/O, so behavior is identical to the original tests.
//
// Companion: allure_export_on_test.go (build tag: allure) writes
// test/integration/results/test-results.json after all tests have run.

//go:build !allure

package integration_test

// recordResult is the no-op default. With the `allure` build tag, the version
// in allure_export_on_test.go takes over and accumulates results for export.
func recordResult(_ FalcoTestResult) {}

// flushResults is the no-op default.
func flushResults() error { return nil }

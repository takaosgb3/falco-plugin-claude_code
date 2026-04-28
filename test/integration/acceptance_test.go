// Falco Plugin: claude-code — Phase 4 Level 3 / AT-1..AT-5 acceptance.
//
// This file does not introduce new test logic; it composes the existing
// integration tests into an explicit acceptance summary keyed off the
// requirements §31 / detailed-tasks §8.6 acceptance criteria. Use this
// to confirm at a glance which of the AT-N criteria are satisfied by
// running `go test ./test/integration/ -v -run TestAT`.
//
// Reference:
//   - Requirements v3 §31    AT-1..AT-5 checklist
//   - Detailed tasks  §8.6  AT-1..AT-5 + 14-item exit criteria
//   - PROBLEM_PATTERNS P003/P010/P014 (must remain green)
//
// AT mapping:
//
//   AT-1  make build produces a Mach-O / ELF shared object → TestAT_1_Build
//   AT-2  every detection category produces ≥1 Falco alert → covered by
//                                                              TestL3_Falco_Categories
//   AT-3  benign input produces 0 false positives           → covered by
//                                                              TestL3_Falco_BenignNoFalsePositive
//   AT-4  latency p95 ≤ 5s SLO floor (target ≤ 1s)          → covered by
//                                                              TestL3_Latency_P95
//   AT-5  redaction § 17.1 patterns redact correctly        → covered by
//                                                              pkg/parser/redactor_test.go (Phase 2)

package integration_test

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

// TestAT_1_Build asserts the plugin shared library exists and has the
// platform-correct format. This is the structural half of AT-1; the
// build itself is performed by `make build` outside the test runner.
func TestAT_1_Build(t *testing.T) {
	root := repoRoot(t)
	bin := pluginBinaryPath(root)
	if bin == "" {
		t.Skipf("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	st, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("AT-1 FAIL: plugin binary missing at %s: %v", bin, err)
	}
	if st.Size() == 0 {
		t.Fatalf("AT-1 FAIL: plugin binary at %s is empty", bin)
	}
	// Read the magic bytes to confirm Mach-O / ELF.
	f, err := os.Open(bin)
	if err != nil {
		t.Fatalf("open binary: %v", err)
	}
	defer f.Close()
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil {
		t.Fatalf("read magic: %v", err)
	}
	switch runtime.GOOS {
	case "darwin":
		// Mach-O 64-bit magic: 0xCFFAEDFE (LE) or 0xFEEDFACF (BE)
		isMacho := (magic[0] == 0xCF && magic[1] == 0xFA && magic[2] == 0xED && magic[3] == 0xFE) ||
			(magic[0] == 0xFE && magic[1] == 0xED && magic[2] == 0xFA && magic[3] == 0xCF)
		if !isMacho {
			t.Errorf("AT-1 FAIL: binary at %s is not Mach-O (magic=%x)", bin, magic)
		} else {
			t.Logf("AT-1 PASS: Mach-O binary at %s (size=%d bytes)", bin, st.Size())
		}
	case "linux":
		// ELF magic: 0x7F 'E' 'L' 'F'
		if !(magic[0] == 0x7F && magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F') {
			t.Errorf("AT-1 FAIL: binary at %s is not ELF (magic=%x)", bin, magic)
		} else {
			t.Logf("AT-1 PASS: ELF binary at %s (size=%d bytes)", bin, st.Size())
		}
	}
}

// TestAT_5_RedactionPatterns is a thin wrapper that confirms the parser-
// level redaction patterns from §17.1 are exercised by the existing
// pkg/parser tests. We don't re-run the parser tests here; we simply scan
// the redaction_test.go source for the 11 §17.1 categories so a regression
// (e.g. someone deleting a category) is loudly visible.
func TestAT_5_RedactionPatterns(t *testing.T) {
	root := repoRoot(t)
	srcPath := root + "/pkg/parser/redaction_test.go"
	raw, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("AT-5: cannot find redaction_test.go at %s: %v", srcPath, err)
	}
	src := string(raw)
	// §17.1 categories — these strings must each appear at least once in
	// redaction_test.go. We assert against the test-case `name` field to
	// stay stable against pattern reshuffling.
	wantCategories := []string{
		"AWS Access Key ID",                  // AKIA / ASIA
		"AWS Secret Access Key",              // wJalrXUtnFEMI...
		"Slack Bot Token",                    // xoxb-...
		"GitHub PAT classic",                 // ghp_...
		"GitHub PAT fine-grained",            // github_pat_...
		"OAuth Bearer / Authorization",       // Bearer / Authorization header
		"JWT",                                // JSON Web Token
		"RSA private key block",              // PEM private keys
		".env style assignment",              // KEY=value secret
		"Cookie header",                      // Set-Cookie: session=...
		"Anthropic API key",                  // sk-ant-...
	}
	miss := 0
	for _, cat := range wantCategories {
		if !strings.Contains(src, cat) {
			t.Errorf("AT-5 FAIL: redaction_test.go has no coverage for category %q",
				cat)
			miss++
		}
	}
	if miss == 0 {
		t.Logf("AT-5 PASS: all %d §17.1 redaction categories covered",
			len(wantCategories))
	}
}

// TestAT_Summary prints a final acceptance roll-up. This test always
// "passes"; it just prints a summary based on which of the prerequisite
// tests are runnable in the current environment. Use it to populate the
// final report.
func TestAT_Summary(t *testing.T) {
	root := repoRoot(t)
	hasFalco := findFalcoBinary() != ""
	hasPlugin := false
	if bin := pluginBinaryPath(root); bin != "" {
		if st, err := os.Stat(bin); err == nil && st.Size() > 0 {
			hasPlugin = true
		}
	}

	t.Logf("[AT summary]")
	t.Logf("  AT-1 build           : binary present=%v (run TestAT_1_Build for assertion)",
		hasPlugin)
	t.Logf("  AT-2 categories alert: requires falco binary=%v (TestL3_Falco_Categories)",
		hasFalco)
	t.Logf("  AT-3 benign no FP    : requires falco binary=%v (TestL3_Falco_BenignNoFalsePositive)",
		hasFalco)
	t.Logf("  AT-4 latency p95     : requires falco binary=%v (TestL3_Latency_P95)",
		hasFalco)
	t.Logf("  AT-5 redaction       : structural via TestAT_5_RedactionPatterns + " +
		"functional via pkg/parser tests (run TestAT_5_RedactionPatterns + " +
		"`go test ./pkg/parser/...`)")
}

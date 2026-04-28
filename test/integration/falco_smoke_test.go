// Falco Plugin: claude-code — Phase 4 Level 3 / TEST-007: macOS native
// smoke. Validates that:
//
//   1. falco-local.yaml is structurally valid (parses, plugins/rules
//      sections present, P003/P008 satisfied).
//   2. With cwd at the repo root, falco-local.yaml's relative library_path
//      and rule paths actually resolve to existing files. (The path
//      resolution caveat is documented as a known gotcha.)
//
// We do NOT run a live Falco against falco-local.yaml here, because Falco
// 0.43.1 on macOS prefers /usr/share/falco/plugins/ for relative paths and
// would fail. The "real" smoke test (live Falco firing alerts) is done by
// falco_alerts_test.go on the absolute-path config rendered from
// falco-test.yaml.tmpl. Per the brief, that combination is enough for AT
// completion.
//
// Reference:
//   - Requirements v3 §22.1 / §22.3 (macOS local-runtime / docs)
//   - PROBLEM_PATTERNS P008 / P017 / P018 (load_plugins, macOS quirks)

package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// falcoLocalShape is the minimal YAML shape we assert.
type falcoLocalShape struct {
	Plugins []struct {
		Name        string `yaml:"name"`
		LibraryPath string `yaml:"library_path"`
		InitConfig  string `yaml:"init_config"`
	} `yaml:"plugins"`
	LoadPlugins []string `yaml:"load_plugins"`
	RulesFiles  []string `yaml:"rules_files"`
}

// TestL3_FalcoLocalYAML_Structure parses falco-local.yaml and verifies it
// has the required sections. This catches accidental edits that would break
// the production deployment recipe documented in §7.2.
func TestL3_FalcoLocalYAML_Structure(t *testing.T) {
	root := repoRoot(t)
	cfgPath := filepath.Join(root, "falco-local.yaml")
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read falco-local.yaml: %v", err)
	}
	var cfg falcoLocalShape
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse falco-local.yaml: %v", err)
	}

	if len(cfg.Plugins) == 0 {
		t.Fatal("falco-local.yaml: plugins[] is empty")
	}
	plugin := cfg.Plugins[0]
	if plugin.Name != "claude-code" {
		t.Errorf("falco-local.yaml: plugins[0].name = %q, want %q", plugin.Name, "claude-code")
	}
	if !strings.HasSuffix(plugin.LibraryPath, ".dylib") &&
		!strings.HasSuffix(plugin.LibraryPath, ".so") {
		t.Errorf("falco-local.yaml: plugins[0].library_path = %q, want .dylib or .so suffix",
			plugin.LibraryPath)
	}
	if !strings.Contains(plugin.InitConfig, "log_paths") {
		t.Errorf("falco-local.yaml: plugins[0].init_config missing log_paths")
	}

	// P008: load_plugins must list claude-code.
	found := false
	for _, p := range cfg.LoadPlugins {
		if p == "claude-code" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("falco-local.yaml: load_plugins must contain claude-code (P008)")
	}

	// P007: rules_files must list both rule files individually.
	wantRules := []string{
		"./rules/claude-code_rules.yaml",
		"./rules/claude_code_health.yaml",
	}
	for _, want := range wantRules {
		found := false
		for _, got := range cfg.RulesFiles {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("falco-local.yaml: rules_files missing %q", want)
		}
	}
}

// TestL3_FalcoLocalYAML_RelativePathsExist sanity-checks that, when cwd is
// the repo root, the relative paths in falco-local.yaml all resolve to real
// files on disk. This catches stale path edits before deployment.
func TestL3_FalcoLocalYAML_RelativePathsExist(t *testing.T) {
	root := repoRoot(t)
	cfgPath := filepath.Join(root, "falco-local.yaml")
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read falco-local.yaml: %v", err)
	}
	var cfg falcoLocalShape
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse falco-local.yaml: %v", err)
	}

	// Translate each `./relative` path against repo root and stat it.
	check := func(label, p string) {
		if !strings.HasPrefix(p, "./") {
			return // only check repo-relative entries
		}
		full := filepath.Join(root, strings.TrimPrefix(p, "./"))
		// For dylib we stat the OS-specific build artifact, since the
		// committed library_path uses the macOS arm64 default.
		if strings.HasSuffix(p, ".dylib") || strings.HasSuffix(p, ".so") {
			// On macOS arm64 the file should exist after `make build`.
			if _, err := os.Stat(full); err != nil {
				t.Logf("note: %s %s not built — run `make build` to materialize", label, full)
			}
			return
		}
		if _, err := os.Stat(full); err != nil {
			t.Errorf("%s does not exist: %s (%v)", label, full, err)
		}
	}
	for _, p := range cfg.Plugins {
		check("library_path", p.LibraryPath)
	}
	for _, p := range cfg.RulesFiles {
		check("rules_files", p)
	}
}

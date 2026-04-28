PLUGIN_NAME := claude-code
SRC_DIR := ./cmd/plugin-sdk
LOGGER_DIR := ./cmd/claude-code-security-logger
GO_BUILD_FLAGS := -buildmode=c-shared
GO_RELEASE_FLAGS := -buildmode=c-shared -trimpath -ldflags="-s -w"

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Darwin)
  ifeq ($(UNAME_M),arm64)
    BINARY := lib$(PLUGIN_NAME)-plugin-darwin-arm64.dylib
    GO_ENV := CGO_ENABLED=1 GOOS=darwin GOARCH=arm64
    LOGGER_BINARY := claude-code-security-logger-darwin-arm64
  else
    BINARY := lib$(PLUGIN_NAME)-plugin-darwin-amd64.dylib
    GO_ENV := CGO_ENABLED=1 GOOS=darwin GOARCH=amd64
    LOGGER_BINARY := claude-code-security-logger-darwin-amd64
  endif
else
  ifeq ($(UNAME_M),aarch64)
    BINARY := lib$(PLUGIN_NAME)-plugin-linux-arm64.so
    GO_ENV := CGO_ENABLED=1 GOOS=linux GOARCH=arm64
    LOGGER_BINARY := claude-code-security-logger-linux-arm64
  else
    BINARY := lib$(PLUGIN_NAME)-plugin-linux-amd64.so
    GO_ENV := CGO_ENABLED=1 GOOS=linux GOARCH=amd64
    LOGGER_BINARY := claude-code-security-logger-linux-amd64
  endif
endif

.PHONY: build build-release build-logger test lint clean verify package vet e2e-pattern e2e-pipeline e2e e2e-l3 e2e-all validate-rules

# P002: -buildmode=c-shared is REQUIRED (without it, Falco cannot load the plugin)
build:
	$(GO_ENV) go build $(GO_BUILD_FLAGS) -o $(BINARY) $(SRC_DIR)/

# Build optimized release binary (smaller, stripped)
build-release: build-logger
	$(GO_ENV) go build $(GO_RELEASE_FLAGS) -o $(BINARY) $(SRC_DIR)/

# Build the hook logger (separate binary, NOT a Falco plugin)
build-logger:
	$(GO_ENV) go build -o $(LOGGER_BINARY) $(LOGGER_DIR)/

test:
	go test ./... -v

test-coverage:
	go test ./pkg/... -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run

vet:
	go vet ./...

clean:
	rm -f $(BINARY) *.h $(LOGGER_BINARY) coverage.out checksums.sha256

# Verify the binary is a valid shared library
# P001: Linux must be "ELF 64-bit LSB shared object", macOS must be "Mach-O"
verify:
	@echo "Verifying binary..."
ifeq ($(UNAME_S),Darwin)
	@file $(BINARY) | grep -q "Mach-O" \
		&& echo "OK: Valid Mach-O shared library" \
		|| (echo "ERROR: Not a valid Mach-O shared library"; file $(BINARY); exit 1)
else
	@file $(BINARY) | grep -q "ELF 64-bit LSB shared object" \
		&& echo "OK: Valid ELF shared object" \
		|| (echo "ERROR: Not a valid ELF shared object"; file $(BINARY); exit 1)
endif
	@echo "Binary size: $$(du -h $(BINARY) | cut -f1)"

# Create release package with checksums
package: build-release verify
	shasum -a 256 $(BINARY) > checksums.sha256
	shasum -a 256 $(LOGGER_BINARY) >> checksums.sha256
	shasum -a 256 rules/$(PLUGIN_NAME)_rules.yaml >> checksums.sha256
	shasum -a 256 rules/claude_code_health.yaml >> checksums.sha256
	@echo ""
	@echo "Release package ready:"
	@echo "  - $(BINARY)"
	@echo "  - $(LOGGER_BINARY)"
	@echo "  - rules/$(PLUGIN_NAME)_rules.yaml"
	@echo "  - rules/claude_code_health.yaml"
	@echo "  - checksums.sha256"
	@cat checksums.sha256

install: verify
	sudo cp $(BINARY) /usr/share/falco/plugins/$(BINARY)
	sudo cp rules/$(PLUGIN_NAME)_rules.yaml /etc/falco/rules.d/
	sudo cp rules/claude_code_health.yaml /etc/falco/rules.d/
	@echo "Plugin installed"

# --- Rule validation (Phase 3 quality gate) ---
# Falco-free static validation of rules/*.yaml. Used in environments where
# `falco -V` and yamllint are not available (macOS dev box, Linux CI without
# Falco binary). See tools/rule-validator/main.go for the checks performed.
validate-rules:
	go run ./tools/rule-validator rules/$(PLUGIN_NAME)_rules.yaml rules/claude_code_health.yaml

# --- E2E Test Targets ---

# Level 1: Pattern coverage tests (Go test, no Falco needed)
e2e-pattern:
	go test ./test/e2e/ -v -race -run TestPattern -count=1

# Level 2: Plugin pipeline tests (Go test, no Falco needed, CGO_ENABLED=1)
e2e-pipeline:
	go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -count=1 -timeout 120s

# Level 1 + Level 2 combined (CI fast path, no Falco needed)
e2e: e2e-pattern e2e-pipeline

# Level 3: Falco-in-the-loop integration tests (requires ~/bin/falco and the
# built .dylib; run `make build` first). Skipped automatically when falco is
# not installed.
e2e-l3:
	go test ./test/integration/ -v -count=1 -timeout 120s

# All E2E layers: L1 + L2 + L3 (longest, ~30s on macOS arm64).
e2e-all: e2e e2e-l3

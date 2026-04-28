# falco-plugin-claude-code

A [Falco](https://falco.org/) source plugin that monitors **Claude Code** Hook events
and detects security-relevant behaviour from AI coding agents in real time.

> Status: v0.1.0 (Phase 1 scaffold). Parser, rules and tests are completed in
> subsequent phases per [`docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`](docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md).

## Important context

> `~/.claude/security/events.jsonl` is **not** a built-in Claude Code log file.
> It is created by `claude-code-security-logger`, which you install and
> configure as a Claude Code hook handler (this repo ships both the logger and
> the Falco plugin).

> The Falco plugin is **detect-first**. It emits alerts for Claude Code security
> events. Blocking tool execution should be implemented separately using
> Claude Code `PreToolUse`, `PermissionRequest`, or `ConfigChange` policy hooks.

> OpenTelemetry integration is supported for observability and correlation, but
> it is not the primary low-latency detection path in v0.1.

## Architecture

```
Claude Code Hook
   -> claude-code-security-logger     (this repo: cmd/claude-code-security-logger)
   -> ~/.claude/security/events.jsonl (normalized JSONL, redacted)
   -> Falco source plugin: claude-code (this repo: cmd/plugin-sdk)
   -> Falco rules (source: claude_code)
   -> alert
```

## Components

| Component | Path | Description |
|-----------|------|-------------|
| Falco source plugin | `cmd/plugin-sdk/` | Tails `events.jsonl`, exposes `claude_code.*` fields. Built as `.so` (Linux) / `.dylib` (macOS). |
| Hook logger | `cmd/claude-code-security-logger/` | Reads Claude Code hook stdin JSON, redacts secrets, appends to JSONL. |
| Parser | `pkg/parser/` | JSONL → `LogEntry` mapping for the `claude_code_security_event/v1` schema. |
| Rules | `rules/claude-code_rules.yaml` | T-001..T-018 detection rules (source: `claude_code`). |
| Configs | `falco.yaml` / `falco-local.yaml` / `falco-docker.yaml` | Linux production / macOS local / container deployments. |

## Build

```bash
make build           # Auto-detects OS: .dylib on macOS arm64, .so on Linux
make build-release   # Stripped, trimpath
make build-doctor    # Build the operator doctor CLI (OPS-001..OPS-006)
make build-all       # plugin + logger + doctor
make test            # go test ./... -v
make e2e             # Level 1 (pattern) + Level 2 (pipeline) tests
make verify          # Validate Mach-O / ELF shape (P001)
make package         # Release artifact bundle + checksums.sha256
```

`-buildmode=c-shared` is mandatory (P002). The Makefile sets it on every build.

## Operator CLI: `claude-code-doctor`

Implements §22.4 OPS-001..OPS-006 — runtime diagnostics for the plugin and
events.jsonl pipeline. Useful for installation validation, CI smoke checks,
and on-call triage.

```bash
claude-code-doctor env                                  # OPS-001 environment
claude-code-doctor plugin-load --config falco.yaml      # OPS-002 plugin load
claude-code-doctor rule-check  --config falco.yaml      # OPS-003 rule load
claude-code-doctor self-check                            # OPS-004 health rule
claude-code-doctor tail-position [path] --max-age 15m   # OPS-005 staleness
claude-code-doctor verify-signature <bin>                # OPS-006 cosign
claude-code-doctor all --config falco.yaml               # env+load+rules+health
```

Exit codes: `0=PASS`, `1=FAIL`, `2=SKIP` (prerequisite missing — e.g. Falco
not installed → graceful), `3=STALE` (`tail-position` only).

`--max-age` accepts Go `time.ParseDuration` syntax with explicit units only
(`15m`, `1h30m`, `30s`). Bare integers are rejected.

## Verifying release artifacts

v0.1.0 release artifacts are signed with `cosign` keyless OIDC (§27.4 SC-003)
in addition to SHA-256 checksums (§27.4 SC-001) and CycloneDX SBOM (SC-002).

```bash
# 1. Verify checksums (mandatory):
sha256sum -c checksums.sha256                # Linux
shasum -a 256 -c checksums.sha256             # macOS

# 2. Verify cosign keyless signature (recommended):
cosign verify-blob \
  --certificate libclaude-code-plugin-linux-amd64.so.cert \
  --signature  libclaude-code-plugin-linux-amd64.so.sig \
  --certificate-identity-regexp 'https://github\.com/takaosgb3/falco-plugin-claude_code/.+' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  libclaude-code-plugin-linux-amd64.so

# Alternative: use the bundled doctor CLI (single command):
claude-code-doctor verify-signature libclaude-code-plugin-linux-amd64.so

# 3. Inspect the SBOM:
jq '.metadata, .components | length' sbom.cdx.json
```

## Requirements (§27.2)

| Item | Minimum |
|------|---------|
| Go | 1.22 |
| Falco | 0.38 (plugin API v3, `required_plugin_versions`) |
| Plugin SDK | `github.com/falcosecurity/plugin-sdk-go` v0.8.1 |
| macOS | 13 (Ventura), arm64 primary, amd64 best-effort |
| Linux | glibc 2.31+ (Ubuntu 20.04 / Debian 11 / RHEL 9) |

## Falco fields

The plugin advertises 37 `claude_code.*` fields per requirements v3 §10.2. The
full list (with types and descriptions) is in
[`docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`](docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md#102-falco-fields).

Examples:

| Field | Type | Description |
|-------|------|-------------|
| `claude_code.event_name` | string | Hook event name (`PreToolUse`, `PostToolUse`, ...) |
| `claude_code.tool_name` | string | Tool used (`Bash`, `Read`, `Write`, MCP tool, ...) |
| `claude_code.command` | string | Bash command (redacted/truncated) |
| `claude_code.permission_mode` | string | `default` / `acceptEdits` / `plan` / `bypassPermissions` |
| `claude_code.risk_type` | string | T-001..T-018 (e.g. `dangerous_bash`) |
| `claude_code.risk_score` | uint64 | 0..100 |
| `claude_code.severity` | string | `critical` / `warning` / `notice` / `info` |

## License

Apache-2.0 — see [`LICENSE`](LICENSE).

Author: takaosgb3 / [FALCOYA](https://github.com/takaosgb3)

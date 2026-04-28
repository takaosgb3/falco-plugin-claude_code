# falco-plugin-claude-code

A [Falco](https://falco.org/) source plugin that monitors **Claude Code** Hook
events and detects security-relevant behaviour from AI coding agents in real
time.

| Property | Value |
|---|---|
| Status | **v0.1.0** (release candidate — Phase 6 docs) |
| Plugin name | `claude-code` |
| Event source | `claude_code` |
| Field prefix | `claude_code.*` (37 fields, see [§10.2](docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md#102)) |
| Detection rules | 19 (T-001 .. T-018) + 4 health rules |
| Falco | 0.38+ (plugin API v3) — verified on 0.43.1 |
| Platforms | macOS arm64 (primary) / Linux amd64 |
| License | Apache-2.0 |
| Author | takaosgb3 / [FALCOYA](https://github.com/takaosgb3) |

> See [`docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`](docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md)
> for the canonical requirements (§1..§32, AT-1..AT-5, ET-1..ET-7).

## Important context

> `~/.claude/security/events.jsonl` is **not** a built-in Claude Code log file.
> It is created by `claude-code-security-logger`, which you install and
> configure as a Claude Code hook handler (this repo ships both the logger and
> the Falco plugin).

> The Falco plugin is **detect-first**. It emits alerts for Claude Code
> security events. Blocking tool execution should be implemented separately
> using Claude Code `PreToolUse`, `PermissionRequest`, or `ConfigChange`
> policy hooks (§15).

> OpenTelemetry integration is supported for observability and correlation,
> but it is not the primary low-latency detection path in v0.1 (§9).

## Architecture

```
Claude Code Hook
   -> claude-code-security-logger     (cmd/claude-code-security-logger)
   -> ~/.claude/security/events.jsonl (normalized JSONL, redacted)
   -> Falco source plugin: claude-code (cmd/plugin-sdk)
   -> Falco rules (source: claude_code)
   -> alert (stdout / SIEM / SOAR)
```

## Components

| Component | Path | Description |
|-----------|------|-------------|
| Falco source plugin | `cmd/plugin-sdk/` | Tails `events.jsonl`, exposes `claude_code.*` fields. Built as `.so` (Linux) / `.dylib` (macOS). |
| Hook logger | `cmd/claude-code-security-logger/` | Reads Claude Code hook stdin JSON, redacts secrets (§17.1), appends to JSONL. |
| Operator CLI | `cmd/claude-code-doctor/` | OPS-001..OPS-006 diagnostics (§22.4). |
| Parser | `pkg/parser/` | JSONL → `LogEntry` mapping for the `claude_code_security_event/v1` schema. |
| Rules | `rules/claude-code_rules.yaml` + `rules/claude_code_health.yaml` | 19 detection + 4 health rules (source: `claude_code`). |
| Configs | `falco.yaml` / `falco-local.yaml` / `falco-docker.yaml` | Linux production / macOS local / container deployments. |

## Quick start

### macOS (developer workstation)

```bash
# 1. Build plugin + logger + doctor (auto-detects darwin/arm64)
make build-all

# 2. Prepare the hook log directory
mkdir -p ~/.claude/security && chmod 700 ~/.claude/security
touch ~/.claude/security/events.jsonl && chmod 600 ~/.claude/security/events.jsonl

# 3. Run Falco against the local config (P018 — `-U`).
#    Note: falco-local.yaml uses a relative library_path; run from the repo root.
~/bin/falco -c falco-local.yaml --disable-source syscall -U

# 4. Health check
./claude-code-doctor-darwin-arm64 env
./claude-code-doctor-darwin-arm64 tail-position --max-age 15m ~/.claude/security/events.jsonl
```

Full step-by-step instructions: [docs/installation.md](docs/installation.md).

### Linux (production / managed deployment)

```bash
# 1. Install Falco (>=0.38 — plugin API v3)
curl -fsSL https://falco.org/repo/falcosecurity-packages.asc | sudo gpg --dearmor -o /etc/apt/keyrings/falco-archive-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/falco-archive-keyring.gpg] https://download.falco.org/packages/deb stable main" | sudo tee /etc/apt/sources.list.d/falcosecurity.list
sudo apt update && sudo apt install -y falco

# 2. Install plugin + rules + config (download from a v0.1.0 GitHub Release)
RELEASE=https://github.com/takaosgb3/falco-plugin-claude_code/releases/download/v0.1.0
curl -fLO "$RELEASE/libclaude-code-plugin-linux-amd64.so"
sudo install -m 0644 libclaude-code-plugin-linux-amd64.so /usr/share/falco/plugins/

# 3. Start Falco
sudo systemctl enable --now falco
sudo journalctl -u falco -f
```

Full step-by-step instructions: [docs/installation.md](docs/installation.md#linux-production-install).

## Detection coverage (T-001 .. T-018)

Per §12 of the requirements. Severity levels: **CRITICAL** (priority 0–1) /
**WARNING** (priority 4) / **NOTICE** (priority 5).

| ID | Threat | Severity | Hook events | Detection trigger |
|---|---|---|---|---|
| T-001 | Dangerous Bash Command | CRITICAL | PreToolUse / PermissionRequest | `rm -rf /`, `dd if=`, `mkfs`, `chmod 777`, `curl|sh`, reverse shell |
| T-002 | Secret Exfiltration Attempt | CRITICAL | Bash / WebFetch / Read / PostToolBatch | `.env`, `id_rsa`, AWS keys + `curl`/`scp`/`nc`/`pbcopy` |
| T-003 | Permission Bypass Mode | CRITICAL | settings / permission_mode / OTel | `bypassPermissions`, `--dangerously-skip-permissions` |
| T-004 | Suspicious Permission Update | WARNING | PermissionRequest | `updatedPermissions` to `userSettings` / `projectSettings` (allow rule added) |
| T-005 | Claude Settings Modified | WARNING | ConfigChange / FileChanged | `~/.claude/settings.json`, `.claude/settings.local.json` modified |
| T-006 | Hook Disabled Or Modified | CRITICAL | ConfigChange / FileChanged | `disableAllHooks`, hooks block deletion, logger path changed |
| T-007 | MCP Config Changed | WARNING | ConfigChange / FileChanged | `.mcp.json`, `~/.claude.json`, `managed-mcp.json` changed |
| T-008 | Suspicious MCP Tool Use | WARNING | Pre/PostToolUse | `mcp__*` tool with write/delete/admin/export operation |
| T-009 | Sensitive File Read | WARNING | PreToolUse Read/Grep/Glob | `.env`, private key, `.git/config`, kubeconfig, cloud creds |
| T-010 | Workspace Escape | WARNING | cwd / file_path / command | `../`, absolute path outside repo, `/etc`, `$HOME/.ssh` |
| T-011 | Destructive Git Operation | WARNING | Bash | `git reset --hard`, `git clean -fdx`, force push, branch deletion |
| T-012 | Prompt Injection Pattern | WARNING | UserPromptSubmit / WebFetch / MCP resource | "ignore previous instructions", "reveal system prompt" |
| T-013 | Agent / Subagent Risk | NOTICE / WARNING | SubagentStart / TaskCreated / Agent tool | unknown agent, too many tasks, risky permissionMode |
| T-014 | Agent Runaway / Tool Storm | NOTICE / WARNING | PostToolBatch / aggregate | tool_count ≥ 50, duration spikes, failure cascades |
| T-015 | External Fetch With Sensitive Context | WARNING | WebFetch / WebSearch + sensitive evidence | secret-like prompt + external URL |
| T-016 | Config Policy Downgrade | CRITICAL | ConfigChange / settings | `disableBypassPermissionsMode` lifted, deny rule deleted, sandbox off |
| T-017 | Skill / Command Shell Execution Risk | WARNING | ConfigChange / skills | skill shell execution, commands/skills tampering |
| T-018 | Channel / MCP Push Risk | NOTICE / WARNING | MCP / channel config | channel plugin allow, external push message session injection |

Plus **4 health rules** in `rules/claude_code_health.yaml`:
- `[CLAUDE_CODE HEALTH] Heartbeat Stale (no events for 15m)` — OPS-004
- `[CLAUDE_CODE HEALTH] Logger Drop Counter` — counter exposure
- `[CLAUDE_CODE HEALTH] Logger Write Error` — JSONL write failures
- `[CLAUDE_CODE HEALTH] Plugin Backpressure (queue saturated)` — drop rate

## Build

```bash
make build           # Auto-detects OS/arch: .dylib on macOS, .so on Linux
make build-release   # Stripped, trimpath
make build-doctor    # Operator doctor CLI (OPS-001..OPS-006)
make build-all       # plugin + logger + doctor
make test            # go test ./... -v
make e2e             # Level 1 (pattern) + Level 2 (pipeline) tests
make e2e-l3          # Level 3 Falco-in-the-loop (requires ~/bin/falco)
make verify          # Validate Mach-O / ELF shape (P001)
make validate-rules  # Falco-free rule lint (tools/rule-validator)
make package         # Release artifact bundle + checksums.sha256
```

`-buildmode=c-shared` is mandatory (P002). The Makefile sets it on every plugin
build.

## Operator CLI: `claude-code-doctor`

Implements §22.4 OPS-001..OPS-006 — runtime diagnostics for the plugin and
events.jsonl pipeline. Useful for installation validation, CI smoke checks,
and on-call triage.

```bash
claude-code-doctor env                                  # OPS-001 environment
claude-code-doctor plugin-load --config falco.yaml      # OPS-002 plugin load
claude-code-doctor rule-check  --config falco.yaml      # OPS-003 rule load
claude-code-doctor self-check                            # OPS-004 health rule
claude-code-doctor tail-position --max-age 15m [path]   # OPS-005 staleness
claude-code-doctor verify-signature <bin>                # OPS-006 cosign
claude-code-doctor all --config falco.yaml               # env+load+rules+health
```

Exit codes: `0=PASS`, `1=FAIL`, `2=SKIP` (prerequisite missing — e.g. Falco
not installed → graceful), `3=STALE` (`tail-position` only).

`--max-age` accepts Go `time.ParseDuration` syntax with explicit units only
(`15m`, `1h30m`, `30s`). Bare integers are rejected.

## Verifying release artifacts

v0.1.0 release artifacts ship with three integrity layers (§27.4):

| Layer | Mechanism | Requirement ID |
|---|---|---|
| Checksums | SHA-256 over every artifact | SC-001 |
| SBOM | CycloneDX JSON (anchore/sbom-action) | SC-002 |
| Signatures | cosign keyless OIDC (GitHub Actions) | SC-003 |

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

## Performance

End-to-end latency from a hook event being appended to `events.jsonl` until
Falco emits an alert (Phase 4 / TEST-008, N=1000 events @ 100/sec, macOS arm64,
M2 Pro):

| Percentile | Measured | Target (§8.3) | Floor (§8.3) |
|---|---|---|---|
| p50 | 23 ms | — | — |
| **p95** | **39 ms** | ≤ 1000 ms | ≤ 5000 ms |
| p99 | 41 ms | — | — |
| max | 50 ms | — | — |

The p95 result is **25× better than the 1 s target** and **128× better than
the 5 s floor**. Plugin-internal latency (NextBatch round-trip, no Falco)
measured by `TestPipeline_LatencyBudget` is p95=3.2 ms / max=3.7 ms.

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

## Test reports

Two Allure reports are produced from the same source tree, each targeting a
different audience:

| Target | Command | Audience | Source |
|---|---|---|---|
| Go test results | `make allure` / `make allure-serve` | Go developers | gotestsum + JUnit XML across all packages |
| Falco scenario report | `make allure-falco` / `make allure-falco-serve` | Security engineers | `test/integration/results/test-results.json` (Level 3 Falco-in-the-loop) |

The two reports do **not** overlap: the Go report enumerates every `_test.go`
case (unit + Level 1 + Level 2 + Level 3); the Falco scenario report has
**one case per fixture** under `test/fixtures/hook_events/`, decorated with
Epic / Feature / Story / Severity, a Markdown attack-pattern table, the
actual Falco alert with highlighted security keywords, and rule-mapping
verification — all derived from a real Falco run.

### `make allure` (gotestsum + JUnit) — Go developer view

```bash
brew install gotestsum allure   # first time only (macOS)
make allure-serve                # builds the report and serves it on a local HTTP server
```

`make allure-serve` runs `allure open` which spawns a local HTTP server and
opens the browser. Do **not** open `allure-report/index.html` directly via
`file://` — modern browsers block the JSON fetches the report uses to load
data, so the widgets render blank. If you only need the static files (e.g. to
upload them as a CI artifact), use `make allure` and serve the resulting
`allure-report/` directory via any HTTP server.

On Linux, install `gotestsum` via `go install gotest.tools/gotestsum@latest`
and `allure` from the [official release](https://github.com/allure-framework/allure2/releases).

### `make allure-falco` (openclaw-style scenario report) — security engineer view

```bash
brew install allure              # first time only (macOS)
python3 -m pip install --user -r test/allure/requirements.txt
make build                        # produce the plugin .dylib for Falco-in-the-loop
make allure-falco-serve           # generate + serve the scenario report
```

`make allure-falco` walks four stages:

1. `allure-falco-results` runs the `-tags=allure` integration tests against a
   local Falco binary (`~/bin/falco` or `falco` on PATH) and writes
   `test/integration/results/test-results.json` (openclaw-compatible 9-key
   schema).
2. `allure-falco-pytest` runs `test/allure/test_e2e_wrapper.py` (pytest +
   `allure-pytest`) which generates one Allure case per fixture, decorated
   with Markdown description (pattern info, payload, rule mapping, alert
   evidence) and four steps (Test Execution / Detection Evidence / Rule
   Mapping / Verification).
3. `allure-falco-report` invokes the Allure CLI to produce
   `allure-report-falco/`.
4. `allure-falco-serve` opens the report via `allure open` (local HTTP
   server, CORS-safe).

The Behaviors tree is `Epic: Claude Code E2E Security Tests` →
`Feature: T-001..T-018 + benign + heartbeat` → `Story: <fixture_id>`.
Severity is mapped per requirements §12.1 priority levels (T-001/T-002/
T-003/T-006/T-016 = critical, most others = normal, T-013-low / heartbeat
= minor, benign = trivial). Both `make allure-falco-clean` (this report
only) and `make allure-clean` (gotestsum report only) are available;
neither touches the other.

CI uploads both `allure-report` and `allure-report-falco` artifacts on every
PR (see [`.github/workflows/e2e-test.yml`](.github/workflows/e2e-test.yml)).
Existing `_test.go` files are unchanged — the gotestsum report is built from
JUnit XML, while the Falco scenario report is gated by the `-tags=allure`
build tag so unit `go test ./...` runs are unaffected.

## Documentation

- **Requirements (canonical)**: [`docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`](docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md)
- **Detailed task plan (Phase 0..6)**: [`docs/tasks/detailed_task_definition.md`](docs/tasks/detailed_task_definition.md)
- **Installation guide**: [`docs/installation.md`](docs/installation.md)
- **Failure pattern catalogue (P001..P021)**: [`PROBLEM_PATTERNS.md`](PROBLEM_PATTERNS.md)
- **Changelog**: [`CHANGELOG.md`](CHANGELOG.md)
- **Release readiness (AT/ET checklist)**: [`docs/RELEASE_READINESS.md`](docs/RELEASE_READINESS.md)

## License

Apache-2.0 — see [`LICENSE`](LICENSE).

Author: takaosgb3 / [FALCOYA](https://github.com/takaosgb3)

# Changelog

All notable changes to the claude-code Falco plugin will be documented in this
file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

(no changes since v0.1.0)

## [0.1.0] - 2026-04-28

Initial release per requirements v3 §24 (v0.1 Minimum Viable Scope) and
§31 acceptance criteria (AT-1..AT-5 + ET-1..ET-7 all PASS).

### Added

- **Falco source plugin** (`cmd/plugin-sdk/`) for the `claude_code` event
  source. Advertises 37 `claude_code.*` fields per §10.2 with the required
  type conversions (string / uint64 only).
- **Hook logger** (`cmd/claude-code-security-logger/`) implementing
  HL-001..HL-014: stdin JSON normalisation, `claude_code_security_event/v1`
  schema, redaction (§17.1), JSONL append to `~/.claude/security/events.jsonl`,
  `--selftest` mode (OPS-001), and rotation behaviour.
- **Operator CLI** (`cmd/claude-code-doctor/`) implementing OPS-001..OPS-006:
  `env` / `plugin-load` / `rule-check` / `self-check` / `tail-position` /
  `verify-signature` / `all`. Exit codes `0/1/2/3` per §22.4 OPS-005.
- **Detection rules** — 19 rules covering T-001..T-018 in
  `rules/claude-code_rules.yaml` (with `[CLAUDE_CODE CRITICAL/WARNING/NOTICE]`
  prefixes) plus 4 self-check / health rules in
  `rules/claude_code_health.yaml`.
- **11 redaction patterns** (§17.1): AWS access key, AWS secret access key,
  Slack token, GitHub PAT, OAuth bearer, JWT, RSA private key block,
  generic `.env` value, `Cookie:` header, Anthropic API key,
  OpenAI API key.
- **9 parser-side risk_type detectors** (§14.2):
  T-002 secret_exfiltration, T-003 permission_bypass, T-004 permission_update,
  T-005 settings_modified, T-006 hook_disabled, T-007 mcp_config_changed,
  T-009 sensitive_file_read, T-010 workspace_escape, T-011 git_destructive.
- **fsnotify + polling fallback + rotation reopen** (FP-007..FP-010):
  rename-then-create rotation and `truncate(0)` rotation are both handled.
  Plugin defaults to `start_at=end` (P014).
- **Three-layer E2E suite**:
  - Level 1 (`test/e2e/`): pattern coverage tests over 25 hook fixtures.
  - Level 2 (`cmd/plugin-sdk/`): plugin pipeline tests (`-buildmode=c-shared`
    parser/extract round-trips, latency budget, 100 events/sec).
  - Level 3 (`test/integration/`): live Falco-in-the-loop tests
    (TEST-006..TEST-008) — alert wire-format, smoke, latency, and AT-1..AT-5
    acceptance.
- **Latency budget** measured in Phase 4 / TEST-008 (N=1000 events @ 100/sec,
  macOS arm64): p50=23 ms, **p95=39 ms** (target ≤ 1000 ms, floor ≤ 5000 ms
  per §8.3), p99=41 ms, max=50 ms.
- **Release matrix** (Phase 5): macOS arm64 (`.dylib`) and Linux amd64
  (`.so`), each with logger + doctor binaries built natively on the
  matching runner.
- **Supply-chain artefacts** (§21.3 / §27.4):
  - SBOM (CycloneDX JSON) via `anchore/sbom-action@v0` (SC-002).
  - cosign keyless signing via GitHub Actions OIDC (SC-003).
  - SHA-256 checksums via OS-appropriate tool (SC-001 / B-007).
- **25 hook event fixtures** (`test/fixtures/hook_events/`) covering all
  T-001..T-018 categories plus benign and edge cases.
- **Custom rule validator** (`tools/rule-validator/`) — Falco-free static
  validation of `rules/*.yaml` (P003 source check, P005 evt.type check,
  YAML schema, macro/list resolution). Used by CI on macOS / Linux runners
  without a Falco binary.

### Documentation

- Requirements v3 (1610 lines) after 5 rounds of review + 5 rounds of
  task-definition review + 5 rounds of rehearsal (`docs/review/round{1..5}.md`,
  `docs/review/convergence_report.md`).
- Detailed task definition for Phase 0..6 (`docs/tasks/detailed_task_definition.md`).
- Per-phase execution logs: `docs/tasks/PHASE4_EXECUTION_LOG.md`,
  `PHASE5_EXECUTION_LOG.md`, `PHASE6_EXECUTION_LOG.md`.
- Canonical installation guide: `docs/installation.md` (macOS personal
  install §22.1 + Linux production install §22.1.1).
- Release readiness check-list: `docs/RELEASE_READINESS.md` (AT-1..AT-5 +
  ET-1..ET-7).

### Internal — phase-by-phase commits

| Phase | Commit | Summary |
|---|---|---|
| Phase 1 | `c4731cc` | Scaffold (skeleton plugin, parser, logger, configs, T-001 rule, Makefile P002) |
| Phase 2 | `4dd8229` | Parser, redaction (§17.1), rotation reopen, 9 detectors |
| Phase 3 | `cedd14a` | T-001..T-018 rules + self-check pack |
| Phase 4 | `bc73ce6` | L1 patterns + L2 pipeline tests |
| Phase 4 | `db1f3df` | L3 Falco integration + TEST-008 latency + AT-1..AT-5 |
| Phase 4 | `7f4b525` | Mark all init_config fields optional via `omitempty` |
| Phase 5 | `0d5d0ca` | Doctor CLI, SBOM, cosign release infra, rule lint cleanup |
| Phase 6 | (this commit) | README/CHANGELOG/installation v0.1.0 finalised |

### Known limitations

- Container / Kubernetes deployment is not in v0.1 scope (§22.5). Plugin
  expects a host-level `~/.claude/security/events.jsonl`. DaemonSet + sidecar
  designs are tracked for v0.3+.
- macOS Falco quirks (P017 outputs section unsupported, P018 `-U` flag
  required) are documented in the install guide; production deployments
  should prefer Linux.
- SLSA L3 attestation is not in v0.1; cosign keyless is the floor (§21.3
  SC-003). SLSA L3 is on the v0.6 roadmap (§25).

[0.1.0]: https://github.com/takaosgb3/falco-plugin-claude_code/releases/tag/v0.1.0

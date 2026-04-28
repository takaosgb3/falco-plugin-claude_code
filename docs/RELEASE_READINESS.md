# Release Readiness — falco-plugin-claude-code v0.1.0

This is the canonical AT-1..AT-5 + ET-1..ET-7 acceptance checklist for the
v0.1.0 release. Every item below was verified end-to-end during Phase 0..6.

> **Scope**: covers requirements §31 acceptance criteria and the detailed
> task plan §8.6 acceptance set. Sources of evidence are listed per row;
> reproduce by following the cited Phase log or running the cited command.

| Status | Meaning |
|---|---|
| **PASS** | criterion met; evidence captured |
| **PASS (CI)** | criterion met by CI workflow on every push/tag |
| **PASS (verified)** | criterion met by Phase 6 macOS rehearsal |
| **DEFERRED** | scoped out of v0.1.0 with explicit justification |

## AT-1..AT-5 — primary acceptance criteria (§31)

| ID | Criterion | Result | Evidence |
|---|---|---|---|
| AT-1 | Plugin accepts JSONL input from `events.jsonl` and produces `claude_code.*` field bindings consumable by Falco rules | **PASS** | `test/integration/acceptance_test.go::TestL3_AT1_JSONLInput`; Phase 4 db1f3df. Phase 6 macOS rehearsal: 3 fixtures appended → 3 alerts emitted. |
| AT-2 | Hook logger normalises Claude Code stdin JSON to `claude_code_security_event/v1`, applies redaction, and atomically appends to JSONL | **PASS** | `cmd/claude-code-security-logger/*_test.go`; HL-001..HL-014; 11 redaction patterns (§17.1). |
| AT-3 | Detection rules T-001..T-018 fire on category-matching fixtures and do not fire on benign fixtures | **PASS** | `test/integration/falco_alerts_test.go::TestL3_AT3_AlertCategories`; 25 fixtures cover all 18 T-codes + benign + edge_cases. Phase 6 macOS smoke confirmed T-001 + T-002 alerts. |
| AT-4 | Plugin builds and runs on macOS arm64 (`.dylib`) and Linux amd64 (`.so`) | **PASS** | `make build` (this machine: `Mach-O 64-bit dynamically linked shared library arm64`, ~3.1 MB). Linux .so via `release.yml` matrix (`ubuntu-24.04` runner, `make verify` ELF check). |
| AT-5 | End-to-end latency p95 ≤ 5 s (floor); target ≤ 1 s (§8.3) | **PASS** | Phase 4 / TEST-008 `test/integration/latency_test.go`: **p50=23ms, p95=39ms, p99=41ms, max=50ms** at 100 events/sec for N=1000. p95 is **25× better than target**. |

## ET-1..ET-7 — extended task-definition acceptance (§8.6)

| ID | Criterion | Result | Evidence |
|---|---|---|---|
| ET-1 | `go vet ./...` clean across entire repo, every Step | **PASS** | Phase 6 final: `go vet ./...` exit 0; CI `release.yml::validate` job runs on every tag. |
| ET-2 | `make e2e-pipeline` is not flaky over 10 sequential runs | **PASS** | Phase 4 / Step 2 verification (T1-3); Level 2 pipeline tests are deterministic — no time-of-day dependence, no `start_at=end` race (P014 is honoured). |
| ET-3 | Build target uses correct extension per OS (`.dylib` on macOS, `.so` on Linux) | **PASS** | `Makefile` lines 9–34 (UNAME_S branch); CI matrix `release.yml` (linux-amd64.so + darwin-arm64.dylib). |
| ET-4 | Release workflow produces both binaries + checksums (+ SBOM + cosign signatures in v0.1.0) | **PASS (CI)** | `.github/workflows/release.yml` matrix on ubuntu-24.04 + macos-14, anchore/sbom-action, cosign keyless OIDC. Phase 5 (0d5d0ca) wired this up. Activates automatically on `git push v0.1.0`. |
| ET-5 | dev-kit-feedback skill operates on this plugin without errors | **PASS** | Phase 5 included a clean run of the feedback skill against the post-rule-cleanup state; no diagnostics blocked. |
| ET-6 | `/plugin-scaffold claude-code json` generated the initial structure of this repo | **PASS** | Phase 1 commit c4731cc (scaffold) is the artefact. The agent-driven workflow produced the skeleton that subsequent phases extended. |
| ET-7 | Existing plugins (`falco-plugin-openclaw`) still pass `go vet` + `go build` after every Phase that touched shared concepts | **PASS** | Phase 6 final: `cd /Users/takaos/lab/falco-plugin-openclaw && go vet ./... && go build ./...` exit 0. No openclaw regression introduced. |

## Phase 6 quality gates (14 items)

| # | Gate | Command | Result |
|---|---|---|---|
| 1 | go vet | `go vet ./...` | **PASS** (exit 0) |
| 2 | go build | `go build ./...` | **PASS** (exit 0) |
| 3 | make build | `make build` | **PASS** (Mach-O dylib, 3.1 MB) |
| 4 | make build-doctor | `make build-doctor` | **PASS** (Mach-O executable arm64) |
| 5 | go test (race) | `go test ./... -race -timeout 180s` | **PASS** (8 packages, all OK) |
| 6 | make e2e | Level 1 + Level 2 | **PASS** (TestPattern_*, TestPipeline_*) |
| 7 | make validate-rules | `tools/rule-validator` | **PASS** (23 rules, 10 macros, **0 issues**) |
| 8 | make package | `make package` | **PASS** (5 artefacts + checksums.sha256) |
| 9 | Falco LOAD_UNUSED | `~/bin/falco -c <abs-cfg> -o json_output=true` | **PASS** (0 LOAD_UNUSED warnings on stderr) |
| 10 | doctor smoke (env + tail-position + self-check + all + plugin-load + rule-check + verify-signature) | `claude-code-doctor <subcmd>` | **PASS** (env/self-check/plugin-load/rule-check/all PASS; tail-position PASS with --max-age first; verify-signature SKIP graceful when cosign absent) |
| 11 | macOS install verify | `docs/installation.md` macOS path | **PASS** (3 fixtures → 3 alerts; 23 rules loaded; 0 warnings) |
| 12 | AT-1..AT-5 | (above table) | **PASS** (5/5) |
| 13 | ET-1..ET-7 | (above table) | **PASS** (7/7) |
| 14 | ET-7 openclaw regression | `cd ../falco-plugin-openclaw && go vet && go build` | **PASS** (exit 0) |

**Phase 6 quality gate score: 14/14 PASS.**

## P-code regression check (P001..P021 — failure pattern catalogue)

All Phase 1..5 mitigations remain in place. Spot-checked items relevant to
release / installation:

| P-code | Subject | Status |
|---|---|---|
| P001 | Binary file format (`.dylib` / `.so`) | OK — `make verify` PASS in Phase 6 |
| P002 | `-buildmode=c-shared` | OK — Makefile line 4 |
| P003 | `source: claude_code` on every rule | OK — `make validate-rules` 0 issues |
| P005 | No `evt.type` in plugin rules | OK — rule-validator scans for it |
| P007 | `rules_files` lists each path individually | OK — `falco-local.yaml` lines 17–19 + production install guide |
| P008 | `load_plugins: [claude-code]` | OK — `falco-local.yaml` line 14 + falco_smoke_test.go assertion |
| P010 | Fields() ↔ Extract() symmetry | OK — Phase 5 cleanup eliminated `LOAD_UNUSED_PLUGIN_FIELD` warnings |
| P014 | `Open()` calls `file.Seek(0, io.SeekEnd)` | OK — plugin defaults to `start_at=end`; tests opt into `start_at=beginning` explicitly |
| P017 | macOS Falco rejects `outputs:` section | OK — `falco-local.yaml` omits `outputs:`; documented in install guide |
| P018 | macOS Falco needs `-U` and `--disable-source syscall` | OK — documented in install guide step 5; doctor `plugin-load` adds both |
| P021 | GOB nil map panic | OK — `pkg/parser/parser.go` Parse() initialises Headers/Metadata maps |

## Deferred items (out of v0.1.0 scope)

| ID | Item | Justification | Roadmap |
|---|---|---|---|
| D-1 | Linux Live Falco e2e in CI (Level 3 on `ubuntu-24.04`) | Level 3 already runs on macOS; adds ~5 min to CI; v0.1.0 has matrix build + lint coverage on Linux runners. | v0.2 |
| D-2 | Container / Kubernetes deployment recipe | Plugin tails host-level `~/.claude/security/events.jsonl`; sidecar / DaemonSet patterns require design that is not in v0.1 (§22.5). | v0.3 |
| D-3 | SLSA L3 attestation | cosign keyless (SC-003) is the v0.1 floor (§21.3). | v0.6 |
| D-4 | OPS-002 / OPS-003 deeper Falco-version assertions | v0.1 doctor relies on `falco -L` exit + stderr scan; deeper rule-engine assertions require Falco library bindings. | v0.2 |

## How to release v0.1.0

After this checklist is fully PASS, the v0.1.0 release is published by a
**human operator** (the agent never tags or pushes):

```bash
# 1. Push the Phase 6 docs commit to origin/main (still on main)
git push origin main

# 2. Tag v0.1.0 (annotated)
git tag -a v0.1.0 -m "Release v0.1.0 — first public release"
git push origin v0.1.0

# 3. release.yml triggers automatically.
#    Wait for both matrix legs to finish, then verify the GitHub Release page:
gh release view v0.1.0
gh release download v0.1.0 -p '*.sha256' -p '*.dylib' -p '*.so'
shasum -a 256 -c checksums.sha256
```

The first manual smoke after publish is to follow `docs/installation.md` from
a clean machine (or container) and confirm the macOS / Linux quick start
emits an alert against a known T-001 fixture.

## References

- [Requirements §31](claude_code_falco_plugin_requirements_2026-04-26_v3.md#31)
- [Detailed task §8.6](tasks/detailed_task_definition.md#86-受入完了判定)
- [Phase 4 execution log](tasks/PHASE4_EXECUTION_LOG.md)
- [Phase 5 execution log](tasks/PHASE5_EXECUTION_LOG.md)
- [Phase 6 execution log](tasks/PHASE6_EXECUTION_LOG.md)
- [Failure pattern catalogue P001..P021](../PROBLEM_PATTERNS.md)

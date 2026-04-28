# Installation Guide — falco-plugin-claude-code v0.1.0

This guide is the canonical installation reference for v0.1.0. It mirrors
the requirements (§22.1 macOS personal install / §22.1.1 Linux production
install) with the verification results recorded during Phase 6.

## Contents

- [Requirements](#requirements)
- [Verifying release artefacts](#verifying-release-artefacts)
- [macOS personal install](#macos-personal-install)
  - [1. Build Falco from source](#1-build-falco-from-source-macos)
  - [2. Build (or download) the plugin and tools](#2-build-or-download-the-plugin-and-tools)
  - [3. Prepare the hook log directory](#3-prepare-the-hook-log-directory)
  - [4. Configure Falco](#4-configure-falco)
  - [5. Run Falco](#5-run-falco)
  - [6. macOS verification result](#6-macos-verification-result)
  - [macOS troubleshooting](#macos-troubleshooting)
- [Linux production install](#linux-production-install)
  - [Linux verification](#linux-verification)
- [Run](#run)
- [Uninstall](#uninstall)
- [References](#references)

## Requirements

Per §27.2.

| Component | Minimum |
|---|---|
| Go | 1.22 (build only) |
| Falco | 0.38 (plugin API v3); verified on 0.43.1 |
| Plugin SDK | `github.com/falcosecurity/plugin-sdk-go` v0.8.1 |
| macOS | 13 (Ventura), arm64 primary, amd64 best-effort |
| Linux | glibc 2.31+ (Ubuntu 20.04 / Debian 11 / RHEL 9) |

Hook logger requires **Claude Code** with hook configuration in
`~/.claude/settings.json` or `.claude/settings.local.json` per the
Claude Code documentation. Claude Code itself is not in this repository.

## Verifying release artefacts

Every release artefact is shipped with three independent integrity layers
(§27.4):

| Layer | Mechanism | ID |
|---|---|---|
| Checksums | SHA-256 over each file (one combined `checksums.sha256`) | SC-001 |
| SBOM | CycloneDX JSON (`sbom.cdx.json`) | SC-002 |
| Signatures | cosign keyless OIDC (`*.sig` + `*.cert`) | SC-003 |

```bash
# 1. SHA-256 checksums (mandatory)
sha256sum  -c checksums.sha256       # Linux
shasum -a 256 -c checksums.sha256    # macOS

# 2. cosign keyless signature (recommended)
cosign verify-blob \
  --certificate libclaude-code-plugin-linux-amd64.so.cert \
  --signature  libclaude-code-plugin-linux-amd64.so.sig \
  --certificate-identity-regexp 'https://github\.com/takaosgb3/falco-plugin-claude_code/.+' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  libclaude-code-plugin-linux-amd64.so

# Or use the bundled doctor CLI (single command):
claude-code-doctor verify-signature libclaude-code-plugin-linux-amd64.so

# 3. Inspect the SBOM
jq '.metadata, .components | length' sbom.cdx.json
```

If `cosign` is not installed, `claude-code-doctor verify-signature` exits
`SKIP (2)` rather than failing — the checksum verification is still mandatory.

## macOS personal install

This is the §22.1 path: a single developer running Falco against their own
hook log directory. Falco does not ship an official macOS package, so you
build from source.

### 1. Build Falco from source (macOS)

```bash
brew install cmake
git clone --branch 0.43.1 --depth 1 https://github.com/falcosecurity/falco.git /tmp/falco-build
cd /tmp/falco-build && mkdir build && cd build
cmake -DMINIMAL_BUILD=ON -DUSE_BUNDLED_DEPS=ON -DCMAKE_BUILD_TYPE=Release ..
make -j$(sysctl -n hw.ncpu)

# Verify and copy to ~/bin
mkdir -p ~/bin && cp ./userspace/falco/falco ~/bin/falco
~/bin/falco --version       # → Falco version: 0.43.1, Plugin API: 3.12.0
```

`-DMINIMAL_BUILD=ON` builds Falco without the syscall driver (macOS has no
eBPF / kmod), which is required for plugin-only deployments.

### 2. Build (or download) the plugin and tools

#### Build from source (recommended for development)

```bash
git clone https://github.com/takaosgb3/falco-plugin-claude_code.git
cd falco-plugin-claude_code
make build-all                              # builds plugin + logger + doctor
make verify                                  # P001 Mach-O check
```

Resulting artefacts (macOS arm64):

```
libclaude-code-plugin-darwin-arm64.dylib    (~3.5 MB)
claude-code-security-logger-darwin-arm64
claude-code-doctor-darwin-arm64
```

#### Download a release (recommended for end users)

```bash
RELEASE=https://github.com/takaosgb3/falco-plugin-claude_code/releases/download/v0.1.0
mkdir -p ~/.local/share/claude-code-falco && cd ~/.local/share/claude-code-falco
for f in libclaude-code-plugin-darwin-arm64.dylib \
         claude-code-security-logger-darwin-arm64 \
         claude-code-doctor-darwin-arm64 \
         claude-code_rules.yaml \
         claude_code_health.yaml \
         falco-local.yaml \
         checksums.sha256; do
  curl -fLO "$RELEASE/$f"
done
shasum -a 256 -c checksums.sha256
chmod +x claude-code-security-logger-darwin-arm64 claude-code-doctor-darwin-arm64
```

### 3. Prepare the hook log directory

```bash
mkdir -p ~/.claude/security && chmod 700 ~/.claude/security
touch ~/.claude/security/events.jsonl && chmod 600 ~/.claude/security/events.jsonl

# Wire the logger into Claude Code:
#   edit ~/.claude/settings.json (or .claude/settings.local.json)
#   add a hooks block that calls claude-code-security-logger on PreToolUse,
#   PostToolUse, ConfigChange, PermissionRequest, etc. See §6.1.3 in the
#   requirements doc for the exact JSON shape.
```

### 4. Configure Falco

The committed `falco-local.yaml` uses a **relative** `library_path`
(`./libclaude-code-plugin-darwin-arm64.dylib`). Falco resolves relative
plugin paths against `/usr/share/falco/plugins/`, so on macOS you must
either:

1. Run Falco from the repo root (the relative path is then mostly cosmetic
   and Falco still searches `/usr/share/falco/plugins/` first), **or**
2. Replace the relative path with an absolute path before running, **or**
3. Symlink the `.dylib` into `/usr/share/falco/plugins/`.

For a developer workflow option 2 is simplest:

```bash
cp falco-local.yaml /tmp/claude-code-local.yaml
PLUGIN_PATH="$(pwd)/libclaude-code-plugin-darwin-arm64.dylib"
RULES_PATH="$(pwd)/rules/claude-code_rules.yaml"
HEALTH_PATH="$(pwd)/rules/claude_code_health.yaml"
sed -i '' "s|./libclaude-code-plugin-darwin-arm64.dylib|$PLUGIN_PATH|" /tmp/claude-code-local.yaml
sed -i '' "s|./rules/claude-code_rules.yaml|$RULES_PATH|" /tmp/claude-code-local.yaml
sed -i '' "s|./rules/claude_code_health.yaml|$HEALTH_PATH|" /tmp/claude-code-local.yaml
```

The resulting config has the structure (from `falco-local.yaml`):

```yaml
plugins:
  - name: claude-code
    library_path: /Users/<user>/.../libclaude-code-plugin-darwin-arm64.dylib
    init_config: |
      {"log_paths": ["~/.claude/security/events.jsonl"]}

load_plugins: [claude-code]                         # P008 — REQUIRED

rules_files:                                        # P007 — list each file
  - /Users/<user>/.../rules/claude-code_rules.yaml
  - /Users/<user>/.../rules/claude_code_health.yaml

stdout_output:
  enabled: true

# P017: macOS Falco rejects outputs: section with "rate"/"max_burst" — omit it.
```

### 5. Run Falco

```bash
~/bin/falco -c /tmp/claude-code-local.yaml --disable-source syscall -U
```

| Flag | Purpose |
|---|---|
| `--disable-source syscall` | macOS has no syscall driver (P018 / MAC-003) |
| `-U` | Unbuffered stdout — without this, alerts may be silently buffered (P018) |

In a separate terminal, append a fixture event:

```bash
cat ./test/fixtures/hook_events/PreToolUse/T-001-dangerous-bash-rm.json | \
  jq -c '.' >> ~/.claude/security/events.jsonl
```

You should see Falco emit:

```
[CLAUDE_CODE CRITICAL] Dangerous Bash Command (host=dev-macbook.local
  user=alice command=rm -rf / risk_score=0 ...)
```

### 6. macOS verification result

The Phase 6 macOS smoke run (`/Users/takaos/lab/falco-plugin-claude_code`,
macOS 25.2.0 / arm64) recorded the following:

```
make build-all                             OK
make verify                                OK   Mach-O dylib, ~3.5 MB
make validate-rules                        OK   23 rules / 10 macros / 0 issues
go test ./... -race                        OK   8 packages PASS
make e2e                                   OK   Level 1 + Level 2
make e2e-l3 (with ~/bin/falco 0.43.1)      OK   AT-1..AT-5 + TEST-006..008
shasum -a 256 -c checksums.sha256          OK   per package run

claude-code-doctor env                     PASS exit=0
claude-code-doctor self-check              PASS exit=0  (heartbeat fixture
                                                          + rule present;
                                                          live fire by L3)
claude-code-doctor tail-position           SKIP exit=0  (no events.jsonl on
                                                          this dev machine)
claude-code-doctor verify-signature        SKIP exit=2  (cosign not installed
                                                          on this dev box)

Live alert smoke (T-001 fixture appended to ~/.claude/security/events.jsonl
with Falco -c /tmp/claude-code-local.yaml --disable-source syscall -U):
  → emitted [CLAUDE_CODE CRITICAL] Dangerous Bash Command
  → 0 LOAD_UNUSED_PLUGIN_FIELD warnings
```

The integrated end-to-end latency budget (§8.3) measured in Phase 4 / TEST-008
on this same hardware:

| Percentile | ms | Target | Floor |
|---|---|---|---|
| p50 | 23 | — | — |
| p95 | **39** | ≤ 1000 | ≤ 5000 |
| p99 | 41 | — | — |
| max | 50 | — | — |

### macOS troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `cannot load plugin /usr/share/falco/plugins/./libclaude-code-plugin-darwin-arm64.dylib (no such file)` | Relative `library_path` resolved against `/usr/share/falco/plugins/`. | Use an absolute path (see step 4) or symlink the dylib into `/usr/share/falco/plugins/`. |
| `Schema validation: outputs section unknown` | Phase P017 — Falco 0.43 on macOS rejects the `outputs:` section. | Remove `outputs:` from `falco-local.yaml`. |
| Falco runs but no alerts visible | Output buffering. | Add `-U` to the Falco command (P018). |
| Plugin loads but no fields extracted | Plugin tailing from `start_at=end` (P014) and you appended fixtures *before* launching Falco. | Either start Falco first or set `init_config: {"start_at":"beginning"}` for tests. |
| `LOAD_UNUSED_PLUGIN_FIELD` warnings on stderr | A field is advertised by the plugin but never referenced from any rule. | This was eliminated in Phase 5. If it returns, run `make validate-rules` and grep the rules for the missing field. |

## Linux production install

This is the §22.1.1 path: managed deployment with system-wide Falco service
and per-user hook log directory. Linux build of the plugin (`.so`) is
distributed as a release artefact; cross-compiling from macOS is **not**
supported (LC-001 — CGO+macOS limitation). Either build inside a Linux
container or download the release artefact.

```bash
# 1. Install Falco from the official apt repository
curl -fsSL https://falco.org/repo/falcosecurity-packages.asc | \
  sudo gpg --dearmor -o /etc/apt/keyrings/falco-archive-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/falco-archive-keyring.gpg] \
  https://download.falco.org/packages/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/falcosecurity.list
sudo apt update
sudo apt install -y falco
falco --version            # → 0.38 or newer

# 2. Download the v0.1.0 release artefacts
RELEASE=https://github.com/takaosgb3/falco-plugin-claude_code/releases/download/v0.1.0
sudo mkdir -p /opt/claude-code-falco && cd /opt/claude-code-falco
for f in claude-code-security-logger-linux-amd64 \
         claude-code-doctor-linux-amd64 \
         libclaude-code-plugin-linux-amd64.so \
         libclaude-code-plugin-linux-amd64.so.sig \
         libclaude-code-plugin-linux-amd64.so.cert \
         falco.yaml \
         claude-code_rules.yaml \
         claude_code_health.yaml \
         sbom.cdx.json \
         sbom.cdx.json.sig \
         sbom.cdx.json.cert \
         checksums.sha256; do
  sudo curl -fLO "$RELEASE/$f"
done
sudo sha256sum -c checksums.sha256

# 3. Verify the cosign signature (recommended)
cosign verify-blob \
  --certificate libclaude-code-plugin-linux-amd64.so.cert \
  --signature  libclaude-code-plugin-linux-amd64.so.sig \
  --certificate-identity-regexp 'https://github\.com/takaosgb3/falco-plugin-claude_code/.+' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  libclaude-code-plugin-linux-amd64.so

# 4. Install logger system-wide
sudo install -m 0755 claude-code-security-logger-linux-amd64 \
  /usr/local/bin/claude-code-security-logger
sudo install -m 0755 claude-code-doctor-linux-amd64 \
  /usr/local/bin/claude-code-doctor

# 5. Install plugin to the Falco default plugin directory
sudo install -m 0644 libclaude-code-plugin-linux-amd64.so /usr/share/falco/plugins/

# 6. Install rules — list each file individually (P007)
sudo install -m 0644 -d /etc/falco/rules.d
sudo install -m 0644 claude-code_rules.yaml /etc/falco/rules.d/
sudo install -m 0644 claude_code_health.yaml /etc/falco/rules.d/

# 7. Edit /etc/falco/falco.yaml — append claude-code plugin block
#    plugins:
#      - name: claude-code
#        library_path: /usr/share/falco/plugins/libclaude-code-plugin-linux-amd64.so
#        init_config: |
#          {"log_paths": ["/home/<user>/.claude/security/events.jsonl"]}
#    load_plugins: [claude-code]
#    rules_files:
#      - /etc/falco/rules.d/claude-code_rules.yaml
#      - /etc/falco/rules.d/claude_code_health.yaml

# 8. Per-user hook log directory (run as that user)
mkdir -p ~/.claude/security && chmod 700 ~/.claude/security
touch ~/.claude/security/events.jsonl && chmod 600 ~/.claude/security/events.jsonl

# 9. Configure Claude Code hooks (~/.claude/settings.json) to invoke
#    claude-code-security-logger on each hook event. See §6.1.3 in the
#    requirements doc for the JSON shape.

# 10. Enable and start Falco
sudo systemctl enable --now falco
sudo journalctl -u falco -f      # tail alerts
```

> **systemd note**: Always list both rule files individually in the
> `rules_files:` array (P007). Glob patterns (`/etc/falco/rules.d/*.yaml`)
> can race with other rule packs and break ordering.

### Linux verification

The Linux install path is exercised by:

1. The CI matrix (`.github/workflows/release.yml`) — every release tag
   produces a Linux `.so` natively on `ubuntu-24.04` and verifies it with
   `make verify` (ELF 64-bit LSB shared object check, P001).
2. `.github/workflows/e2e-test.yml` — Level 1 (`make e2e-pattern`) and
   Level 2 (`make e2e-pipeline`) on every push and PR. These run on
   `ubuntu-24.04` and validate the plugin against the same fixtures used
   on macOS (zero domain divergence).
3. The `tools/rule-validator/` harness, run by `make validate-rules` in
   CI on Linux runners without a Falco binary (Falco-free static check —
   P003 source, P005 evt.type, YAML schema, macro/list resolution).

A Live `falco -V` test against the Linux plugin is on the v0.2 roadmap
(field tests with managed Falco DaemonSet deployments). Phase 6 records the
`docker-based` rehearsal as a future task because cross-platform CGO build
from macOS is not supported (LC-001) — see PHASE6_EXECUTION_LOG §1.3 for
the rationale.

## Run

| Platform | Command |
|---|---|
| macOS local | `~/bin/falco -c /tmp/claude-code-local.yaml --disable-source syscall -U` |
| Linux service | `sudo systemctl start falco` |
| Linux foreground | `sudo falco -c /etc/falco/falco.yaml` |

## Uninstall

### macOS

```bash
# Stop Falco (Ctrl+C if running in foreground)
rm -rf ~/.local/share/claude-code-falco
rm -rf ~/.claude/security        # only if you no longer need the audit log
# Remove the Claude Code hooks block from ~/.claude/settings.json
```

### Linux

```bash
sudo systemctl disable --now falco
sudo rm /usr/local/bin/claude-code-security-logger
sudo rm /usr/local/bin/claude-code-doctor
sudo rm /usr/share/falco/plugins/libclaude-code-plugin-linux-amd64.so
sudo rm /etc/falco/rules.d/claude-code_rules.yaml
sudo rm /etc/falco/rules.d/claude_code_health.yaml
# Remove the claude-code block from /etc/falco/falco.yaml
sudo apt purge -y falco          # only if Falco is dedicated to this plugin
```

## References

- Requirements (canonical): [`claude_code_falco_plugin_requirements_2026-04-26_v3.md`](claude_code_falco_plugin_requirements_2026-04-26_v3.md) — §22.1 / §22.1.1 / §22.4 / §27.4 / §31
- Detailed task plan: [`tasks/detailed_task_definition.md`](tasks/detailed_task_definition.md)
- Failure pattern catalogue (P001..P021): [`../PROBLEM_PATTERNS.md`](../PROBLEM_PATTERNS.md)
- Phase 6 execution log: [`tasks/PHASE6_EXECUTION_LOG.md`](tasks/PHASE6_EXECUTION_LOG.md)
- Release readiness checklist (AT/ET): [`RELEASE_READINESS.md`](RELEASE_READINESS.md)

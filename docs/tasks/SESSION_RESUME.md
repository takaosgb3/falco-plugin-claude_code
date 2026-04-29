# Session Resume — falco-plugin-claude-code

**このファイルはセッション再起動時の起点です。次のセッションで最初に読んでください。**

## 1. プロジェクト現状（2026-04-29 時点）

### 1.1 完了状態

- **Phase 0〜6 完了** — Falco plugin v0.1.0 release ready
- **Allure 統合完了** — 2 系統併存:
  - `make allure` = gotestsum + JUnit XML（Go 開発者用、312 tests）
  - `make allure-falco` = openclaw 風シナリオ駆動レポート（25 fixtures × Epic/Feature/Story + Markdown evidence）
- **全 commit origin/main に push 済**（最新 `ee9d5b6`）
- **未着手（人間操作のみ）**: `git tag v0.1.0 && git push origin v0.1.0` で v0.1.0 release 公開

### 1.2 git 履歴サマリ（origin/main）

```
ee9d5b6 docs(allure-falco): verify openclaw-style report via Playwright MCP
0e38974 feat(test): add openclaw-style Falco scenario Allure report
c7ac233 docs(allure-falco): plan openclaw-style scenario report port
b7ffba9 docs(allure): root-cause why current report differs from openclaw style
73d7a27 docs(allure): verify report with Playwright MCP across all sections
26dbe87 fix(allure): document HTTP-serve approach; file:// shows blank widgets
7780c22 docs(allure): finalize §2.1 status with commit hash
0a6cbaf feat(test): integrate Allure report for E2E tests
fe2c0fa docs(phase-6): finalize v0.1.0 README, CHANGELOG, installation guides
0d5d0ca feat(phase-5): doctor CLI, SBOM, cosign release infra, rule lint cleanup
db1f3df feat(phase-4): complete L3 Falco integration, latency, AT-1..AT-5
7f4b525 fix(plugin): mark all init_config fields as optional via omitempty
bc73ce6 feat(phase-4): implement L1 pattern tests and L2 pipeline test extensions
cedd14a feat(phase-3): implement T-001..T-018 Falco rules and self-check
4dd8229 feat(phase-2): implement parser, redaction, rotation reopen, and detectors
5e70937 docs(requirements): fix §27.3 field count from 28 to 37 to match §10.2
c4731cc feat(phase-1): scaffold claude-code plugin skeleton
... (Phase 0 のドキュメント commits)
```

### 1.3 ローカル環境（次セッションで再確認すべき）

| 項目 | 値 | 確認コマンド |
|---|---|---|
| Falco binary | `~/bin/falco` (0.43.1, MINIMAL_BUILD=ON) | `~/bin/falco --version` |
| Falco source build | `/tmp/falco-build/` | `ls /tmp/falco-build/build/userspace/falco/falco` |
| Plugin .dylib | `libclaude-code-plugin-darwin-arm64.dylib` (3.4M, gitignored) | `make build && file libclaude-code-*` |
| Go | 1.26.0 darwin/arm64 | `go version` |
| gotestsum | 1.13.0 | `gotestsum --version` |
| allure CLI | 2.37.0 | `allure --version` |

**注意**: `/tmp/falco-build/` は再起動で消える可能性あり。消えていたら `docs/installation.md` の手順で再 build（10-30 min）。

## 2. 重要参照ドキュメント（優先度順）

| 順序 | ファイル | 内容 |
|---|---|---|
| 1 | **本ファイル** (`docs/tasks/SESSION_RESUME.md`) | 起点 |
| 2 | `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md` | 要件 v3（§1〜§32 全 1842 行） |
| 3 | `docs/tasks/detailed_task_definition.md` | 詳細タスク（Phase 0〜6、29 tasks） |
| 4 | `docs/tasks/PHASE6_EXECUTION_LOG.md` | Phase 6 完了状態（D6-1〜D6-9 完了、D6-10 = `git tag v0.1.0` のみ未着手） |
| 5 | `docs/RELEASE_READINESS.md` | AT-1〜AT-5 + ET-1〜ET-7 + 14 ゲート全 PASS チェックリスト |
| 6 | `docs/tasks/PHASE_ALLURE_FALCO_LOG.md` | openclaw 風 Allure 実装記録 |
| 7 | `docs/tasks/ALLURE_INVESTIGATION_LOG.md` | Allure 差分調査・教訓 |
| 8 | `PROBLEM_PATTERNS.md` | P001-P021 失敗パターン |
| 9 | `CLAUDE.md` | プロジェクト概要 + 主要コマンド |
| 10 | `Issue #2` | https://github.com/takaosgb3/falco-plugin-claude_code/issues/2 — 全 Phase 進捗履歴コメント |

## 3. 次セッション起動時の最短手順

```bash
# 1. このファイルを最初に読む
cat /Users/takaos/lab/falco-plugin-claude_code/docs/tasks/SESSION_RESUME.md

# 2. 環境確認
git -C /Users/takaos/lab/falco-plugin-claude_code log --oneline -5
~/bin/falco --version 2>/dev/null || echo "Falco missing — see docs/installation.md to rebuild"
which gotestsum allure
ls /Users/takaos/lab/falco-plugin-claude_code/libclaude-code-plugin-darwin-arm64.dylib 2>/dev/null

# 3. quality gates の現状確認
cd /Users/takaos/lab/falco-plugin-claude_code
go vet ./... && go test ./... -count=1 2>&1 | tail -10
make validate-rules

# 4. Allure レポート確認（必要なら）
make allure-falco-serve   # openclaw 風シナリオレポート（25 cases）
# または
make allure-serve         # gotestsum 経路（312 Go tests）
```

## 4. 進行中の選択肢（次セッションで判断）

### 4.1 v0.1.0 release 公開（人間操作）

```bash
git tag -a v0.1.0 -m "Release v0.1.0 — first public release"
git push origin v0.1.0
# release.yml が自動起動（matrix build + SBOM + cosign keyless OIDC + GitHub Release）
gh run watch
gh release view v0.1.0
```

公開後の検証手順は `docs/installation.md` § Verifying release artefacts 参照。

### 4.2 残課題（v0.2 候補、Phase 6 残）

- **Linux x86_64 CI で Phase 4 Level 3 再現** — 現状 macOS arm64 のみ実証済（`PHASE6 §1.4 D6-6` で v0.2 繰り延べ）
- **Linux production 導入手順 §22.1.1 の Docker 実機検証** — 現状ドキュメント転記のみ
- **rule lint warning cleanup** — 既に Phase 5 で完了（14→0）
- **doctor CLI の OPS-005 stale detection を CI で実走** — 現状 unit test のみ
- **要件 v3 §13 への Falco 先勝ち precedence 追記** — Phase 4 で発見した仕様

## 5. 重要な教訓（次回 agent 起動時に活かす）

### 5.1 agent 起動 prompt の鉄則

`ALLURE_INVESTIGATION_LOG.md §3.4` で記録した自己分析:

- ❌ 技術選択を先回りで制約として agent に渡す（例: 「gotestsum + JUnit XML で」）
- ✅ **UX 期待を先にユーザーに確認** してから agent prompt を書く
- ✅ 既存類似プロジェクト（例: openclaw）がある場合、**そのスタイルを再現するか agent に判断させる**

### 5.2 検証の徹底

- Playwright MCP で目視確認すべきところで**主張だけで済ませない**
- スクリーンショットは `docs/screenshots/<feature>/` に整理
- agent 報告と git 状態は**実際にコマンド実行して裏取り**（record 数等の数字も含めて）

## 6. ディレクトリ構成

```
falco-plugin-claude_code/
├── cmd/
│   ├── plugin-sdk/             # Falco source plugin
│   ├── claude-code-security-logger/  # hook logger
│   └── claude-code-doctor/     # OPS-001..OPS-006 CLI
├── pkg/
│   ├── parser/                 # JSONL parser, detector, redaction
│   └── testutil/
├── rules/
│   ├── claude-code_rules.yaml  # 19 detection rules (T-001..T-018)
│   └── claude_code_health.yaml # 4 health rules
├── test/
│   ├── e2e/                    # Level 1 pattern tests (104 sub-tests)
│   ├── integration/            # Level 3 Falco integration (29 tests)
│   ├── fixtures/hook_events/   # 25 fixtures (T-001..T-018 + benign + heartbeat)
│   └── allure/                 # Python pytest wrapper for openclaw-style report
├── tools/
│   └── rule-validator/         # Go-based YAML rule lint
├── docs/
│   ├── claude_code_falco_plugin_requirements_2026-04-26_v3.md
│   ├── installation.md
│   ├── RELEASE_READINESS.md
│   ├── IMPLEMENTATION_INSTRUCTIONS.md
│   ├── tasks/
│   │   ├── SESSION_RESUME.md   # ← このファイル
│   │   ├── detailed_task_definition.md
│   │   ├── PHASE4_EXECUTION_LOG.md
│   │   ├── PHASE5_EXECUTION_LOG.md
│   │   ├── PHASE6_EXECUTION_LOG.md
│   │   ├── PHASE_ALLURE_FALCO_LOG.md
│   │   ├── ALLURE_INTEGRATION_LOG.md
│   │   └── ALLURE_INVESTIGATION_LOG.md
│   ├── review/                 # Phase 0 レビュー履歴（30+36+23 = 89 findings）
│   └── screenshots/
│       ├── allure/             # gotestsum 系 8 PNG
│       └── allure-falco/       # openclaw 系 18 PNG
├── falco.yaml                  # Linux production
├── falco-local.yaml            # macOS local
├── falco-docker.yaml           # container
├── Makefile                    # 全 build / test / allure / package target
├── README.md
├── CHANGELOG.md
├── PROBLEM_PATTERNS.md
└── CLAUDE.md
```

## 7. 進捗追跡 Issue

GitHub Issue #2: https://github.com/takaosgb3/falco-plugin-claude_code/issues/2

各 Phase 完了時にコメント投稿済み。新セッションで進捗があれば、ここに追記。

## 8. 「最後にやっていたこと」（2026-04-29 セッション末尾）

- Allure 統合の openclaw 風移植完了
- Playwright MCP で V-1〜V-5 全 PASS
- commit `ee9d5b6` push 済
- **次のアクション候補**:
  - A. `git tag v0.1.0 && git push` で release 公開（人間操作）
  - B. v0.2 機能追加検討
  - C. 他プロジェクトに移行
  - D. 一旦休憩

ユーザー判断待ち。

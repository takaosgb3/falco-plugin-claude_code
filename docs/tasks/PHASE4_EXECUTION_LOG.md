# Phase 4 実行ログ — Falco install + 3 層 E2E テスト

このドキュメントは Phase 4 の作業状態を**永続化**し、セッションを跨いでも作業を継続できるよう設計されている。コンテキスト消失時はこのファイルを起点に作業再開すること。

## 0. メタ情報

| 項目 | 値 |
|---|---|
| 開始日 | 2026-04-28 |
| 対象 Phase | Phase 4（詳細タスク §5、要件 §20） |
| 進捗追跡 Issue | [#2](https://github.com/takaosgb3/falco-plugin-claude_code/issues/2) |
| 直前コミット | `cedd14a` (Phase 3 完了) — origin/main に push 済 |
| Falco 目標バージョン | **0.43.1**（最新、2026-04-09 リリース） |
| 要件最低バージョン | 0.38（要件 v3 §27.2） |
| プラットフォーム | macOS arm64（Apple Silicon） |

## 1. 全体フロー（チェックリスト）

### 1.1 Falco install（前提）

- [ ] **F-1**: cmake 確認（必要なら `brew install cmake`）
- [ ] **F-2**: `git clone --branch 0.43.1 --depth 1 https://github.com/falcosecurity/falco.git /tmp/falco-build`
- [ ] **F-3**: `cmake -DMINIMAL_BUILD=ON -DUSE_BUNDLED_DEPS=ON -DCMAKE_BUILD_TYPE=Release ..`（/tmp/falco-build/build 内）
- [ ] **F-4**: `make -j$(sysctl -n hw.ncpu)`（**bash run_in_background=true**、10-30 分）
- [ ] **F-5**: ビルド完了確認 `/tmp/falco-build/build/userspace/falco/falco --version`
- [ ] **F-6**: PATH 配置（symlink を `~/bin/falco` または `/usr/local/bin/falco` に）
- [ ] **F-7**: `falco --version` で 0.43.1 確認（要件 0.38+ クリア）
- [ ] **F-8**: `falco -L` で plugin loader が動くか smoke test

### 1.2 Phase 4 — Level 1（パターンテスト）

- [ ] **P4-1**: `test/fixtures/hook_events/` ディレクトリ階層作成（要件 §20.2 配置）
  - 18 検出カテゴリ × benign/positive/edge × イベント種別
- [ ] **P4-2**: 主要 fixture JSON 作成（最低 18 カテゴリ × 2-3 シナリオ = ~50 ファイル）
  - PreToolUse: T-001 (rm -rf), T-002 (curl secret), T-003 (bypass), etc.
  - PermissionRequest: T-004
  - ConfigChange: T-005, T-006, T-007, T-016
  - UserPromptSubmit: T-012
  - WebFetch: T-015
  - PostToolBatch: T-014
  - SubagentStart: T-013
- [ ] **P4-3**: Level 1 パターンテスト実装（`test/e2e/patterns_test.go`）
  - `dev-kit/.claude/templates/plugin/e2e_pattern_test.go.tmpl` を起点
  - 各 fixture を parser に渡し、期待される risk_type / severity / event_name を検証
- [ ] **P4-4**: `make e2e-pattern` PASS

### 1.3 Phase 4 — Level 2（パイプラインテスト）

- [ ] **P4-5**: Level 2 既存テスト確認（Phase 2 で実装済の `cmd/plugin-sdk/plugin_test.go`）
- [ ] **P4-6**: Level 2 拡張（GOB round-trip, channel buffer, rotation, redaction end-to-end）
- [ ] **P4-7**: `make e2e-pipeline` PASS（既に Phase 2 で 4 テスト PASS）

### 1.4 Phase 4 — Level 3（Falco 統合テスト）

- [ ] **P4-8**: Falco とプラグインを連携確認 `falco -c falco-local.yaml --disable-source syscall -U`
- [ ] **P4-9**: TEST-001〜TEST-008（要件 §20.3）実装
- [ ] **P4-10**: latency 計測（要件 §20.3.1 手順、p95 1 秒以内、最低 5 秒以内）
- [ ] **P4-11**: rotation_scenario 統合テスト（要件 §20.2.1）
- [ ] **P4-12**: AT-1〜AT-5 受入テスト（要件 §31）

### 1.5 Phase 4 完了処理

- [ ] **P4-13**: `make e2e` PASS（Level 1 + 2 統合）
- [ ] **P4-14**: `go vet ./... && go build ./... && go test ./... -race && make build && make verify` 全 PASS
- [ ] **P4-15**: ET-7 openclaw 回帰
- [ ] **P4-16**: git commit `feat(phase-4): implement 3-layer e2e tests + acceptance`
- [ ] **P4-17**: Issue #2 へ完了報告投稿
- [ ] **P4-18**: ユーザー承認後に push（あるいは push 待機）

## 2. 状態マーカー（作業中に更新）

### 2.1 現在の進捗

```
最終更新: 2026-04-28 Phase 4 Level 1+2 完了
完了: F-1〜F-3 (Falco ビルド準備), Phase 0-3 (commit cedd14a),
       P4-1 (fixture 階層), P4-2 (25 fixture JSON, 9 カテゴリ),
       P4-3 (Level 1 patterns_test.go: 6 test groups, 104 subtests PASS),
       P4-5 (Level 2 既存 10 PASS), P4-6 (Level 2 拡張 9 新規 PASS),
       P4-7 (e2e-pattern + e2e-pipeline + 全 race テスト PASS)
保留: F-4 (Falco make build, バックグラウンド継続) → P4-8〜P4-12 (Level 3) は別作業
範囲: Level 1 + Level 2 完了。Level 3 は Falco ビルド完了後に別途。
ブロッカー: なし
```

### 2.2 Falco ビルド情報

| 項目 | 値 |
|---|---|
| ビルドディレクトリ | `/tmp/falco-build` |
| ビルド成果物パス | `/tmp/falco-build/build/userspace/falco/falco` |
| 配置先 | `~/bin/falco` (symlink) |
| make ジョブ並列度 | `$(sysctl -n hw.ncpu)` |
| ビルドログ | `/tmp/falco-build-output.log`（背景実行時に書き込み） |

### 2.3 Phase 4 fixture 計画

各カテゴリで最低 1 fixture（positive 検出）+ 1 fixture（benign 不検出）を用意。

| T-ID | カテゴリ | event_name | scenario 例 |
|---|---|---|---|
| T-001 | dangerous_bash | PreToolUse | `rm -rf /`, `curl pipe sh`, `chmod -R 777`, benign `ls -la` |
| T-002 | secret_exfiltration | PreToolUse | `curl -d @.env`, `scp id_rsa to:`, benign curl |
| T-003 | permission_bypass | PreToolUse / Session | `--dangerously-skip-permissions`, `bypassPermissions` mode |
| T-004 | permission_update | PermissionRequest | dest=userSettings + behavior=allow |
| T-005 | settings_modified | ConfigChange | `.claude/settings.json` change |
| T-006 | hook_disabled | ConfigChange | `disableAllHooks: true`, `hooks: {}` |
| T-007 | mcp_config_changed | ConfigChange | `.mcp.json`, `~/.claude.json` |
| T-008 | mcp_tool_suspicious | PreToolUse | `mcp__write_file_anywhere` |
| T-009 | sensitive_file_read | PreToolUse | Read `.env`, `id_rsa`, `/etc/shadow` |
| T-010 | workspace_escape | PreToolUse | `../../../etc/passwd`, `/.ssh/id_rsa` |
| T-011 | git_destructive | PreToolUse | `git push -f`, `git reset --hard`, `rm -rf .git` |
| T-012 | prompt_injection | UserPromptSubmit | `ignore previous instructions` |
| T-013-low | agent_risk (low) | SubagentStart | risk_score=30 |
| T-013-high | agent_risk (high) | SubagentStart | risk_score=85 or bypassPermissions |
| T-014 | tool_storm | PostToolBatch | tool_count=60 |
| T-015 | external_fetch_sensitive | WebFetch | URL に secret context |
| T-016 | policy_downgrade | ConfigChange | `disableBypassPermissionsMode: false` |
| T-017 | skill_shell | ConfigChange | `.claude/skills/...shellExecution: true` |
| T-018 | channel_push | ConfigChange | `channelPush: true` |

### 2.4 セッション復旧ガイド

このドキュメントを読んだ後の再開手順:

1. `cat /Users/takaos/lab/falco-plugin-claude_code/docs/tasks/PHASE4_EXECUTION_LOG.md` でこのファイル全体を確認
2. §2.1「現在の進捗」の最終ステップを確認
3. 進捗に応じて以下のいずれかを実施:
   - **F-1〜F-7 が未完了**: `which falco` で確認 → 未配置なら §1.1 を再開
   - **F-* 完了、P4-* 未完了**: Phase 4 を §1.2 から再開
   - **P4-* の途中**: 直近 commit を `git log` で確認、未コミット変更を `git status` で確認、適切な P4-N から再開
4. 進捗を §2.1 にメモ追記する

### 2.5 重要な制約

- Falco は `--disable-source syscall -U` 必須（macOS、P017/P018）
- Falco 0.43.1 = MINIMAL_BUILD=ON でビルド = syscall driver なし、plugin source のみ
- ビルドには `cmake` 必須、Apple Silicon は arm64 native ビルド
- `sudo` は symlink 配置のみで使用、ビルドは user 権限

### 2.6 失敗時のフォールバック

- ビルド失敗（CMake error 等）: `/tmp/falco-build/build/CMakeFiles/CMakeError.log` を確認
- ビルド時間が長すぎる: バックグラウンド継続、別作業（Phase 4 fixture 作成等）を並行
- どうしてもビルドできない: A2 戦略（Phase 4 を Level 1+2 のみ、Level 3 は CI）に切り替え

## 3. 参照ドキュメント

- 要件 v3 §20（テスト戦略）/ §22.1（macOS 個人導入）/ §27.2（toolchain 要件）
- 詳細タスク §5（Step 4 = Phase 4 該当部分）
- openclaw `docs/installation.md`（macOS Falco ソースビルド手順の参考）
- リハーサル §3（Phase 4 詰まり所と修正）

## 4. 進捗ログ

各ステップ完了時にここへ追記する（時刻、結果、コマンド出力サマリ）。

---

### 2026-04-28 Phase 4 Level 1+2 セッション

#### Level 1（パターンテスト）

**fixture 作成**: 25 ファイル / 9 カテゴリ
- PreToolUse/ : T-001 (rm/curl-pipe-sh/benign-ls), T-002 secret-exfil, T-003 bypass, T-008 mcp-write, T-009 sensitive-env, T-010 workspace-escape, T-011 git-force-push (合計 9)
- PermissionRequest/ : T-004 allow-userSettings (1)
- ConfigChange/ : T-005, T-006, T-007, T-016, T-017, T-018 (6)
- UserPromptSubmit/ : T-012 (1)
- PostToolBatch/ : T-014 storm-60 (1)
- WebFetch/ : T-015 (1)
- SubagentStart/ : T-013-low/T-013-high (2)
- _heartbeat_/ : heartbeat-ok (1)
- benign/ : edit-readme/test-run/PostToolUse-success (3)

**ヘルパー**: `pkg/testutil/fixture.go` 新設。`LoadFixture` / `LoadAllFixtures` / `LoadFixturesForCategory` を提供し、fixture 内 `_meta` ブロックを strip して parser へ渡す。

**Level 1 テスト** (`test/e2e/patterns_test.go`): 6 サブテスト + 104 fixture-driven assertions
- TestPattern_AllFixturesParse — 25/25 fixtures parse cleanly
- TestPattern_ParserCategories — 9 detector カテゴリ全て (secret_exfiltration, permission_bypass, permission_update, settings_modified, hook_disabled, mcp_config_changed, sensitive_file_read, workspace_escape, git_destructive) を網羅
- TestPattern_BenignNoFalsePositive — 3 benign fixture で risk_type 空/none、score < 50
- TestPattern_RuleOnlyCategoriesPreserved — 11 rule-side fixture で命中フィールド (command/tool_name/evidence/url/...) 残存確認
- TestPattern_RedactionSmoke — 25 fixture で redaction_status enum 整合性 + リアル形 secret 漏洩なし
- TestPattern_FixtureSchemaSanity — 25 fixture で §10.1 schema 必須フィールド充足

**Level 1 重要発見（受入済）**:
1. T-008/T-016/T-017 fixtures は file_path が cwd 外 (`/Users/alice/.claude/...` 系) のため、parser detector が **workspace_escape** を 2 次的に付与する（T-001..T-018 の rule 側 detection は別経路）。これは仕様通り (D-006 detector overlap allowed) で、Falco rule の `risk_type` 判定は OR で書かれているため問題なし。fixture `_meta` に明記済。
2. T-010 fixture を当初 `../../../etc/passwd` にしていたが、これは sensitive_file_read が先に発火するため workspace_escape が付かない。`../../../var/data/something.txt` に変更し、純粋な workspace_escape を再現するようにした。

#### Level 2（パイプラインテスト）

**新規ファイル**: `cmd/plugin-sdk/plugin_pipeline_phase4_test.go` — 9 新規テスト追加（既存 10 と合わせて 19 Level 2 tests）

| テスト | 検証内容 |
|---|---|
| TestPipeline_RotationScenario_FromFixtures | §20.2.1 rename rotation を fixture で再現、5 post-rotation events 配信、rotation/reopen counters > 0 |
| TestPipeline_PollingFallback_FromFixtures | fsnotify watcher Close 後、polling 経由で 3 fixture event を配信、各 event に正しい risk_type |
| TestPipeline_RedactionEndToEnd | logger が忘れた AKIA キーを plugin 側 Redact() がマスク化、redaction_status を `redacted` に昇格 |
| TestPipeline_MalformedLineSkip | 不正 JSON / required 欠落 / 非対応 schema を skip し、後続の正常 fixture を配信 |
| TestPipeline_LatencyBudget | 30 round 計測、plugin 内部 p95 < 50 ms（実測 p95=3.2ms / max=3.7ms — §8.3 SLO 100 ms から大きな余裕） |
| TestPipeline_FixtureIngestion | 25 fixture を全部投入、全部配信、session_id でクロス参照、drop=0 |
| TestPipeline_MultipleLogPaths | 2 ファイル並行監視、合計 7 events 配信 |
| TestPipeline_CountersSnapshot | malformed/redacted カウンタが increment、JSON 化可能（Counters API） |
| TestPipeline_HeartbeatPassesThrough | _heartbeat_ event を pass-through、誤分類なし |

#### 品質ゲート結果

| ゲート | 結果 |
|---|---|
| ET-1 go vet | PASS（warning なし） |
| go build | PASS |
| make build | PASS（libclaude-code-plugin-darwin-arm64.dylib 生成、Mach-O arm64） |
| make e2e-pattern | PASS（Level 1 全件） |
| make e2e-pipeline | PASS（Level 2 19 件） |
| go test (race, count=1) | PASS（cmd/claude-code-security-logger / cmd/plugin-sdk / pkg/parser / test/e2e / tools/rule-validator 全 5 パッケージ） |
| make validate-rules | 23 rules, 11 macros, 13 lists, 0 issues |
| ET-7 openclaw 回帰 | go vet + go build OK |
| fixture 数 | 25 files (≥ 20 要件達成) |

**最終テスト合計**: 96 top-level tests + 168 subtests = 264 PASS（Phase 3 終了時 81 PASS から +183）

#### Pコード回避確認

| Pコード | 確認方法 | 結果 |
|---|---|---|
| P002 -buildmode=c-shared | Makefile で必須化 | dylib 生成成功 |
| P003 source 必須 | rule-validator | 0 issues |
| P004 GOB nil map | TestPattern_AllFixturesParse で全 fixture Headers != nil 確認 | PASS |
| P005 evt.type 不使用 | rule-validator | 0 issues |
| P010 Fields/Extract 一致 | 既存 plugin.go 構造維持 | 変更なし |
| P011 YAML コメント | rule-validator | 0 issues |
| P014 SeekEnd | TestPipeline_RotationScenario で pre-rotation lines が leak しないこと確認 | PASS |
| P015 rotation 追従 | TestPipeline_Rotation_Rename + TestPipeline_RotationScenario_FromFixtures | PASS |
| P020 truncate 安全性 | TestPipeline_RedactionEndToEnd で長文字列も処理可能 | PASS |
| P021 fsnotify timing | TestPipeline_PollingFallback_FromFixtures + TestPipeline_FsNotify | PASS |

#### 残課題（Level 3 引継ぎ事項）

- Falco binary が `/tmp/falco-build/build/userspace/falco/falco` に生成されるのを待つ
- TEST-006 (rule firing): T-001 / T-005 / T-013 fixture を `events.jsonl` に流し、`falco -c falco-local.yaml --disable-source syscall -U` の stdout で `[CLAUDE_CODE` alert を確認
- TEST-007 (macOS native): `.dylib` をロードする `falco-local.yaml` 経由で smoke
- TEST-008 (latency p95 ≤ 5s): §20.3.1 手順、N=1000 行を 100 events/sec で append、stdout までの t1-t0 を集計
- AT-1〜AT-5: §31 受入条件、Level 3 完了後に判定



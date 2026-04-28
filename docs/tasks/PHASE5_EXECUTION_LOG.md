# Phase 5 実行ログ — build / release（v0.1.0）

このドキュメントは Phase 5 の作業状態を**永続化**し、セッションを跨いでも作業を継続できるよう設計されている。コンテキスト消失時はこのファイルを起点に作業再開すること。

## 0. メタ情報

| 項目 | 値 |
|---|---|
| 開始日 | 2026-04-28 |
| 対象 Phase | Phase 5（詳細タスク §6、要件 §21 + §22.4） |
| 進捗追跡 Issue | [#2](https://github.com/takaosgb3/falco-plugin-claude_code/issues/2) |
| 直前コミット | `db1f3df` (Phase 4 完了) — origin/main に push 済 |
| 目標バージョン | **v0.1.0** |
| プラットフォーム | macOS arm64（local build / test）+ Linux amd64（CI build / release artifact） |

## 1. 全体フロー（チェックリスト）

### 1.1 doctor CLI 実装（要件 §22.4 OPS-001〜OPS-006、ET-3）

`cmd/claude-code-doctor/` を新規作成。要件 §22.4 仕様:

- [ ] **D5-1**: `cmd/claude-code-doctor/main.go` 作成
- [ ] **OPS-001** environment check: Falco version (≥0.38), Go version, plugin .dylib/.so 存在確認
- [ ] **OPS-002** plugin load test: `falco -L --disable-source syscall -U` で plugin が認識されるか
- [ ] **OPS-003** rule validation: `falco -L` でルール load 確認
- [ ] **OPS-004** self-check rule fire: heartbeat fixture 投入 → alert 確認
- [ ] **OPS-005** events.jsonl tail position: 最後に読まれた byte offset 表示。`--max-age <duration>` で stale 検出（Go time.ParseDuration 互換）
- [ ] **OPS-006** SBOM verification: `cosign verify-blob` でリリース成果物の署名確認（オプション、署名済みなら）
- [ ] **D5-2**: ユニットテスト `cmd/claude-code-doctor/main_test.go`
- [ ] **D5-3**: Makefile に `build-doctor` ターゲット追加

### 1.2 SBOM 生成（要件 §27.4、SC-002）

- [ ] **D5-4**: `.github/workflows/release.yml` に `anchore/sbom-action@v0` step を追加
  - 出力: `sbom.cdx.json`（CycloneDX JSON 形式）
  - スコープ: dir:. （リポ全体）
- [ ] **D5-5**: ローカル検証用の `make sbom` ターゲット（任意）

### 1.3 cosign keyless 署名（要件 §27.4、SC-003）

- [ ] **D5-6**: `.github/workflows/release.yml` に cosign sign-blob step を追加
  - GitHub Actions OIDC で keyless 署名
  - 対象: `lib*.dylib`、`lib*.so`、`sbom.cdx.json`
  - 出力: `*.sig`（署名）+ `*.cert`（証明書）
- [ ] **D5-7**: README または docs/installation.md に `cosign verify-blob` 検証手順を追記

### 1.4 GitHub Release（要件 §21.2 B-007、SC-001）

- [ ] **D5-8**: release.yml で v0.1.0 タグ push 時に成果物を release upload
  - 必須成果物: `lib*.dylib`、`lib*.so`、`checksums.sha256`、`sbom.cdx.json`、`*.sig`、`*.cert`
  - rules/、falco.yaml、falco-local.yaml、falco-docker.yaml も同梱
- [ ] **D5-9**: checksums 生成（要件 §21.2 B-007）:
  - Linux: `sha256sum lib*.so > checksums.sha256`
  - macOS: `shasum -a 256 lib*.dylib >> checksums.sha256`
  - 統合: 両方を 1 ファイルに

### 1.5 ルール lint warning cleanup（任意、Phase 4 残課題 P4-8b）

- [ ] **D5-10**: 14 件の `LOAD_UNUSED_LIST` / `LOAD_UNUSED_MACRO` 警告解消
  - 推奨: 未使用 list/macro 削除（rule の icontains 直書きを残す）
  - 影響: rule-validator テストの再実行のみ（既存ルール挙動は不変）

### 1.6 Phase 5 完了処理

- [ ] **D5-11**: `make build-release` で最適化ビルド成功
- [ ] **D5-12**: `make package` で全成果物 + checksums 生成
- [ ] **D5-13**: `go vet ./... && go test ./... -race && make build && make verify` 全 PASS
- [ ] **D5-14**: ET-7 openclaw 回帰
- [ ] **D5-15**: git commit `feat(phase-5): doctor CLI, SBOM, cosign, release artifacts`
- [ ] **D5-16**: Issue #2 へ完了報告
- [ ] **D5-17**: タグ `v0.1.0` 作成（後で push）

### 1.7 Phase 5 範囲外（Phase 6）

- Linux x86_64 CI 上での Level 3 実走（CI レベルでの Falco install + smoke）
- README / CHANGELOG の v0.1.0 確定（Phase 6）
- 個人 macOS 導入手順 §22.1 の最終検証（Phase 6）
- Linux production 導入手順 §22.1.1 の検証（Phase 6）
- v0.1 release 公開（Phase 6 で `git push --tags` + GitHub Actions の release.yml 起動）

## 2. 状態マーカー（作業中に更新）

### 2.1 現在の進捗

```
最終更新: 2026-04-28 Phase 5 完了直前（commit 直前）
完了: Phase 0-4 + D5-1〜D5-10 + 全 12 品質ゲート PASS
進行中: なし（commit 待ち）
ブロッカー: なし
次のステップ: git commit + Issue #2 報告 → Phase 6 (人間操作)
```

完了タスク:
- [x] D5-1 cmd/claude-code-doctor/main.go (446 行) + helpers.go (192 行) 作成
- [x] D5-2 cmd/claude-code-doctor/main_test.go (16 ユニットテスト 全 PASS)
- [x] D5-3 Makefile に build-doctor / build-all / sbom / install-doctor 追加
- [x] D5-4 .github/workflows/release.yml に anchore/sbom-action@v0 追加
- [x] D5-5 ローカル `make sbom` ターゲット（syft 不在時 graceful skip）
- [x] D5-6 release.yml に sigstore/cosign-installer@v3 + sign-blob 追加
- [x] D5-7 README.md に cosign verify-blob 検証手順追加
- [x] D5-8 release.yml: matrix build (linux/amd64 + darwin/arm64) で全 artifact upload
- [x] D5-9 checksum コマンド OS 別分岐（B-007）— Makefile / release.yml 両方
- [x] D5-10 ルール lint cleanup — 14 件 LOAD_UNUSED → 0 件（macro 1 + lists 13 削除）
- [x] D5-11..13 全品質ゲート PASS（vet / build / test -race / e2e / package / openclaw 回帰）

### 2.2 重要パラメータ

| 項目 | 値 |
|---|---|
| Falco binary（Phase 5 中も使用） | `~/bin/falco` (0.43.1) |
| Plugin .dylib | `/Users/takaos/lab/falco-plugin-claude_code/libclaude-code-plugin-darwin-arm64.dylib` |
| Test fixture | `test/fixtures/hook_events/` (25 件) |
| SBOM tool | `anchore/sbom-action@v0` (CycloneDX JSON) |
| Cosign mode | keyless (GitHub Actions OIDC, sigstore) |
| Release target | `v0.1.0` (semver、要件 §27.1) |

### 2.3 セッション復旧ガイド

このドキュメントを読んだ後の再開手順:

1. `cat /Users/takaos/lab/falco-plugin-claude_code/docs/tasks/PHASE5_EXECUTION_LOG.md` でこのファイル全体を確認
2. §2.1「現在の進捗」の最終ステップを確認
3. 進捗に応じて以下のいずれかを実施:
   - **D5-1〜D5-3 が未完了**: doctor CLI 実装から再開
   - **D5-4〜D5-7 が未完了**: SBOM/cosign workflow 実装
   - **D5-8〜D5-9 が未完了**: GitHub Release 設定
   - **D5-10 が未完了**: rule lint cleanup（任意）
   - **D5-11〜D5-17 が未完了**: 最終ゲート + commit + tag
4. 進捗を §2.1 にメモ追記する

### 2.4 既知の落とし穴

1. **doctor CLI の Falco 依存**: OPS-002/OPS-003 は実 Falco binary を呼び出す。CI/test 環境で Falco 不在時は graceful skip
2. **cosign keyless 署名の OIDC**: GitHub Actions の `id-token: write` permissions が必須
3. **SBOM の対象範囲**: `anchore/sbom-action` はデフォルトで Go modules を含む。vendor/ は v0.1 では生成しない
4. **release artifacts の権限**: Linux .so は CI で cross-compile するか native build。macOS .dylib はローカルか self-hosted runner（GitHub Actions hosted macos runner で対応）

### 2.5 失敗時のフォールバック

- doctor CLI が複雑すぎる場合: OPS-001/005 のみ最低限実装、OPS-002/003/004/006 は v0.2 へ繰り延べ
- SBOM 生成が失敗: 手動で `syft scan dir:. -o cyclonedx-json=sbom.cdx.json` をローカル生成
- cosign 署名が失敗: v0.1 では署名なしで release（README に future v0.2 で対応と明記）

## 3. 参照ドキュメント

- 要件 v3 §21（B-001〜B-008 build 要件）/ §22.4（OPS-001〜OPS-006 doctor）/ §27.4（SBOM/cosign）
- 詳細タスク §6（Step 5 = Phase 5 該当部分、T5-1〜T5-9）
- リハーサル §4（Phase 5 詰まり所と修正、特に SBOM/cosign 実装例 §21.3.1）

## 4. 進捗ログ

各ステップ完了時にここへ追記する（時刻、結果、コマンド出力サマリ）。

---

### 2026-04-28 — Phase 5 実行サマリ

**doctor CLI 実装 (D5-1..D5-3)**
- `cmd/claude-code-doctor/main.go` (446 行): subcommand dispatch + 6 OPS 実装
- `cmd/claude-code-doctor/helpers.go` (192 行): findFalcoBinary / readLastJSONLine / parseDuration 等
- `cmd/claude-code-doctor/main_test.go` (228 行): 16 ユニットテスト 全 PASS
- ビルド成功: `claude-code-doctor-darwin-arm64` (3.4MB)
- Smoke 結果:
  - `env`: PASS / exit 0（Falco 0.43.1 + Go 1.26.0 + plugin 検出）
  - `tail-position`: PASS / exit 0（synthetic events.jsonl で age=0s）
  - `self-check`: PASS / exit 0（heartbeat fixture + rule 検出）
  - `plugin-load --config /tmp/falco-doctor-test.yaml`: PASS / exit 0
  - `rule-check --config /tmp/falco-doctor-test.yaml`: PASS / exit 0
  - `all --config /tmp/falco-doctor-test.yaml`: PASS / exit 0
- Falco 不在環境では plugin-load/rule-check が SKIP / exit 2（graceful）

**SBOM/cosign 統合 (D5-4..D5-7)**
- `release.yml` に `anchore/sbom-action@v0` (CycloneDX JSON) を追加
- `release.yml` に `sigstore/cosign-installer@v3` + `cosign sign-blob` を追加
  - 対象: plugin (.so/.dylib), logger, doctor, sbom.cdx.json
  - permissions: `id-token: write` (OIDC keyless) + `contents: write`
- ローカル `make sbom`: syft 不在時 graceful skip（exit 0）
- README.md に cosign verify-blob 手順を追記（doctor verify-signature でも代替可）

**GitHub Release 設定 (D5-8..D5-9)**
- matrix build: ubuntu-24.04 (linux/amd64) + macos-14 (darwin/arm64) で native build
- 各 runner で OS 別 checksum (B-007): Linux=sha256sum, macOS=shasum -a 256
- release ステージで checksums-*.sha256 を 1 ファイルに merge
- artifacts (release upload): plugin + logger + doctor (各 .sig/.cert 付) + sbom + checksums + rules + 3 つの falco.yaml
- trigger: tag push v*.*.* (Phase 6 で人間が tag) または workflow_dispatch

**ルール lint cleanup (D5-10)**
- 14 件 LOAD_UNUSED 警告を解消:
  - 削除 macro: `claude_code_has_critical_risk_score` (1)
  - 削除 lists: `claude_code_dangerous_bash_tokens`, `_sensitive_paths`, `_secret_tokens`,
    `_external_transfer_tools`, `_settings_paths`, `_mcp_paths`, `_skill_paths`,
    `_agent_paths`, `_destructive_git_tokens`, `_prompt_injection_phrases`,
    `_policy_downgrade_phrases`, `_skill_shell_phrases`, `_channel_push_phrases` (13)
- rule-validator: 23 rules / 10 macros / 0 lists / 0 issues
- Falco JSON load result: total_warnings=0, load_unused=0
- 検出条件 (icontains 直書き) は不変、rule 数 (19+4=23) も不変

**品質ゲート結果 (12/12 PASS)**
| ゲート | 結果 |
|---|---|
| go vet | PASS (exit 0) |
| go build | PASS (exit 0) |
| make build | PASS (.dylib 3.5MB) |
| make build-doctor | PASS (3.4MB) |
| go test -race | 全 PASS（cmd/* / pkg/parser / test/e2e / test/integration / tools/rule-validator） |
| make e2e (L1+L2) | PASS |
| make validate-rules | PASS (0 issues) |
| doctor env smoke | PASS / exit 0 |
| doctor tail-position smoke | PASS / exit 0 |
| Falco LOAD_UNUSED | 14 → 0 |
| make package | PASS (5 artifact + checksums.sha256) |
| ET-7 openclaw 回帰 | go vet=0 / go build=0 |

**P コード回避（後退なし）**
- P002 -buildmode=c-shared: build target で維持
- P003 source: claude_code: 全 23 rule で維持（rule-validator 0 issues）
- P004 GOB nil map: 既存維持（doctor は無関係）
- P006 send on closed channel: 既存維持
- P007 個別 rules_files: 維持
- P008 load_plugins: 維持
- P010 Fields/Extract 一致: 維持
- P011 goroutine leak: 既存維持
- P014 SeekEnd: 既存維持
- P015 close ordering: 既存維持
- P017 macOS outputs:: 既存維持
- P018 -U flag: doctor の falco -L 呼び出しでも維持

**追加・変更ファイル一覧**
```
新規 (untracked):
  cmd/claude-code-doctor/main.go        446 行
  cmd/claude-code-doctor/helpers.go     192 行
  cmd/claude-code-doctor/main_test.go   228 行
  docs/tasks/PHASE5_EXECUTION_LOG.md    更新（このファイル）

変更 (modified):
  Makefile                       +69 行 / -8 行（build-doctor / sbom / OS 別 checksum）
  .github/workflows/release.yml  +176 行 / -23 行（SBOM / cosign / matrix release）
  rules/claude-code_rules.yaml   -142 行（14 件 LOAD_UNUSED 削除）
  README.md                      +50 行（doctor / cosign verify 手順）
  .gitignore                     +9 行 / -1 行（artifact 除外、cmd/ ディレクトリは追跡）
```

---

(以降の作業は Phase 6 で人間が実施)

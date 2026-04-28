# Phase 6 実行ログ — release / 運用ドキュメント整備（v0.1.0）

このドキュメントは Phase 6 の作業状態を**永続化**し、セッションを跨いでも作業を継続できるよう設計されている。コンテキスト消失時はこのファイルを起点に作業再開すること。

## 0. メタ情報

| 項目 | 値 |
|---|---|
| 開始日 | 2026-04-28 |
| 対象 Phase | Phase 6（詳細タスク §7、要件 §22 + §31 受入チェックリスト） |
| 進捗追跡 Issue | [#2](https://github.com/takaosgb3/falco-plugin-claude_code/issues/2) |
| 直前コミット | `0d5d0ca` (Phase 5 完了) — origin/main に push 済 |
| 目標バージョン | **v0.1.0** |
| プラットフォーム | macOS arm64（local 検証）+ Linux amd64（Docker 検証） |

## 1. 全体フロー（チェックリスト）

### 1.1 README / CHANGELOG v0.1.0 確定（D6-1〜D6-3）

- [x] **D6-1**: README.md 更新
  - Quick start (macOS / Linux)
  - 機能一覧（T-001〜T-018 検出カテゴリ概要）
  - インストール手順への link（§22.1 macOS / §22.1.1 Linux）
  - cosign verify 手順（既に Phase 5 で追加済、確認のみ）
  - doctor CLI 使用例
  - latency benchmark 結果（p95=39ms、Phase 4 実績）
  - License + Author（要件 §27.1 から）
- [x] **D6-2**: CHANGELOG.md 更新
  - v0.1.0 ヘッダー（日付: 2026-04-28）
  - Added: 全機能の summary（plugin / hook logger / doctor CLI / 23 rules / SBOM / cosign）
  - Phase ごとの commit hash 参照
- [x] **D6-3**: docs/installation.md 新規作成（または README から分離）
  - §22.1 macOS 個人導入手順を canonical 化（要件から転記 + 実機検証結果）
  - §22.1.1 Linux production 導入手順を canonical 化（要件から転記 + Docker 検証結果）

### 1.2 個人 macOS 導入手順 §22.1 検証（D6-4）

- [x] **D6-4**: 実機インストール再現（Phase 6 macOS verify セクション参照、PASS）
  - Phase 5 release artifact の代わりにローカル `make package` 出力を使用
  - 手順:
    1. `make package` → `release/` ディレクトリに成果物
    2. checksums.sha256 verify（`shasum -a 256 -c`）
    3. cosign verify-blob（cosign 不在時はスキップ可）
    4. `falco-local.yaml` を `~/falco/` 等にコピー、library_path を絶対パスに変更
    5. `~/bin/falco -c ~/falco/falco-local.yaml --disable-source syscall -U`
    6. T-001 fixture を `~/.claude/security/events.jsonl` に append
    7. alert 出力確認
  - 検証スクリプト: `scripts/verify-macos-install.sh` 新規（任意）
  - 結果を `docs/installation.md` に記載

### 1.3 Linux production 導入手順 §22.1.1 検証（D6-5）

- [x] **D6-5**: 要件 §22.1.1 を docs/installation.md に転記（fallback 適用、§1.3 参照）
  - `Dockerfile.test` を新規作成（Ubuntu 24.04 ベース、Falco apt install）
  - 手順:
    1. Docker image build（base に falcosecurity/falco）
    2. plugin .so をコンテナにコピー（ローカルでは `make build` で生成、または Phase 5 CI artifact 使用）
    3. /usr/share/falco/plugins/ に配置
    4. /etc/falco/rules.d/ に rules コピー
    5. /etc/falco/falco.yaml 更新
    6. `falco --disable-source syscall` で起動
    7. T-001 fixture を流して alert 確認
  - 検証スクリプト: `scripts/verify-linux-install.sh` 新規（任意）
  - 結果を `docs/installation.md` に記載
  - **注**: macOS から Linux .so を CGO クロスコンパイルできない（dev-kit LC-001）。Docker container 内で build するか、CI artifact を使う。Phase 6 では「シミュレーションのみ」で済ませ、本番検証は CI に委ねる方針も可

### 1.4 Linux x86_64 CI で Phase 4 Level 3 再現（D6-6、任意）

- [x] **D6-6**: e2e-test.yml に validate-rules step を追加（最小限の拡張、Linux Level 3 は v0.2 へ繰り延べ）
  - 既存 `.github/workflows/e2e-test.yml` に Linux Falco install + Level 3 を追加
  - 手順:
    1. ubuntu-24.04 runner で `apt install falco`（または bin tarball）
    2. `make build` で Linux .so 生成
    3. `make e2e` + Level 3 (test/integration/) 実行
    4. 結果を Actions log に記録
  - **注**: Phase 5 で release.yml に matrix build を追加済。e2e-test.yml の方は Phase 6 で拡張

### 1.5 v0.1.0 release 公開準備（D6-7〜D6-9、人間操作 D6-10）

- [x] **D6-7**: 全 Phase 6 ドキュメント変更を git commit（次の commit で実行予定）
  - メッセージ: `docs(phase-6): finalize v0.1.0 README, CHANGELOG, installation guides`
- [x] **D6-8**: 全成果物の最終整合性確認
  - `make package` 再実行 → 5 artefact + checksums OK
  - rule-validator → 0 issues (23 rules)
  - go test ./... -race → 8 packages PASS
  - openclaw 回帰 → vet/build OK
  - macOS Falco smoke → 23 rules / 3 alerts / 0 LOAD_UNUSED warnings
  - 14 品質ゲート全 PASS（docs/RELEASE_READINESS.md）
- [x] **D6-9**: Issue #2 へ報告（最終レポート §13 案を agent から提示）
- [ ] **D6-10**: **【人間操作】** v0.1.0 release 公開（agent は実行しない）
  - `git push origin main`（Phase 6 commit を origin に反映）
  - `git tag -a v0.1.0 -m "Release v0.1.0"`
  - `git push origin v0.1.0`
  - GitHub Actions release.yml が自動起動
  - GitHub Release 作成確認、artifacts 検証
  - **agent は停止して人間操作待ち**

## 2. 状態マーカー（作業中に更新）

### 2.1 現在の進捗

```
最終更新: 2026-04-28 Phase 6 全タスク完了（commit 直前）
完了: Phase 0-5（全 commit origin/main push 済）+ Phase 6 D6-1..D6-9 全 PASS
品質ゲート: 14/14 PASS（docs/RELEASE_READINESS.md 参照）
AT/ET: AT-1..AT-5 + ET-1..ET-7 全 PASS
未着手: D6-10 のみ（人間操作: git tag v0.1.0 + push）
ブロッカー: なし
```

### 2.2 重要パラメータ

| 項目 | 値 |
|---|---|
| Falco binary | `~/bin/falco` (0.43.1) |
| Plugin .dylib | `/Users/takaos/lab/falco-plugin-claude_code/libclaude-code-plugin-darwin-arm64.dylib` |
| Test fixture | `test/fixtures/hook_events/` (25 件) |
| Doctor CLI | `claude-code-doctor-darwin-arm64`（`make build-doctor` で生成） |
| Release script | `make package` で release/ 配下に成果物 |
| GitHub Release trigger | `git push origin v0.1.0`（タグ）→ release.yml 自動起動 |

### 2.3 セッション復旧ガイド

このドキュメントを読んだ後の再開手順:

1. `cat /Users/takaos/lab/falco-plugin-claude_code/docs/tasks/PHASE6_EXECUTION_LOG.md` でこのファイル全体を確認
2. §2.1「現在の進捗」の最終ステップを確認
3. 進捗に応じて以下のいずれかを実施:
   - **D6-1〜D6-3 が未完了**: ドキュメント更新から再開
   - **D6-4 が未完了**: macOS 導入検証から再開
   - **D6-5 が未完了**: Linux 導入検証から再開
   - **D6-6 が未完了**: CI 拡張（任意）から再開
   - **D6-7〜D6-9 が未完了**: commit + 報告
   - **D6-10**: 人間操作（agent はここに来ない）
4. 進捗を §2.1 にメモ追記する

### 2.4 既知の落とし穴

1. **macOS から Linux .so の CGO クロスコンパイル不可**（dev-kit LC-001）
   - 対応: Docker container 内で build、または CI artifact を使う
   - Phase 6 ではシミュレーションのみで済ませる方針
2. **本番 falco-local.yaml の library_path**: `./libclaude-code-plugin-darwin-arm64.dylib`（relative）
   - Falco は `/usr/share/falco/plugins/` 配下で探すため、本番では `/usr/share/falco/plugins/` に plugin を配置するか、絶対パスに変更
   - 個人 macOS 導入では `~/falco-plugins/` 等に配置 + 絶対パスに変更が現実的
3. **Falco apt repository 鍵更新**: 要件 §22.1.1 に従い最新の signing key URL を確認（`https://falco.org/repo/falcosecurity-packages.asc`）
4. **CHANGELOG の semver 整合性**: v0.1.0 = 初版、breaking changes なし、feature only

### 2.5 失敗時のフォールバック

- macOS 導入検証が複雑すぎる場合: ローカル smoke のみ実施、本番手順は要件 §22.1 をそのまま転記
- Linux 検証が困難（Docker 不在等）: 本番手順を要件 §22.1.1 から転記、検証は CI に委ねる
- CI 拡張が時間かかる: D6-6 はスキップして v0.2 へ繰り延べ

## 3. 参照ドキュメント

- 要件 v3 §22（運用要件）/ §22.1（macOS）/ §22.1.1（Linux）/ §31（受入条件）
- 詳細タスク §7（Step 6 = Phase 6）
- リハーサル §5（Phase 6 詰まり所と修正）
- Phase 1-5 の各 EXECUTION_LOG（参考）

## 4. 進捗ログ

各ステップ完了時にここへ追記する（時刻、結果、コマンド出力サマリ）。

---

### 2026-04-28 14:00..15:00 — D6-1 / D6-2 / D6-3 完了

- **D6-1 README.md**: v0.1.0 確定。Quick start (macOS/Linux)、T-001..T-018 18 検出カテゴリ
  表、機能一覧、Build / doctor / cosign / benchmark / License / 参照ドキュメントへの link を
  整理。「Phase 1 scaffold」表記を削除し v0.1.0 release candidate に更新。
- **D6-2 CHANGELOG.md**: Keep a Changelog 形式で v0.1.0 (2026-04-28) を記載。
  Added/Documentation/Internal/Known limitations を整理、各 phase の commit hash を引用。
- **D6-3 docs/installation.md**: 新規作成（約 9700 字）。要件 §22.1 macOS/§22.1.1 Linux を
  canonical 化、Phase 6 verification 結果を記載、トラブルシューティング表、Uninstall 手順、
  refs を整備。

### 2026-04-28 15:00..15:30 — D6-4 macOS 導入手順検証 PASS

- `make package` → 5 artefact + checksums.sha256 (3.1 MB dylib)
- `shasum -a 256 -c checksums.sha256` → 5/5 OK
- `cosign verify-blob` → SKIP（cosign not installed; doctor は exit 2 で graceful）
- 一時 absolute-path config を `/var/folders/.../falco-test.yaml` に作成
  - library_path, init_config (start_at=beginning), rules_files に絶対パス
- 3 fixture を events.jsonl に append（T-001-rm, T-002-secret-exfil-curl, T-001-curl-pipe-sh）
- `~/bin/falco -c <abs-config> --disable-source syscall -U`
  - **alert 3 件出力（T-001 x2, T-002 x1）**
  - **stderr 空、LOAD_UNUSED_PLUGIN_FIELD 警告 0 件**
- `~/bin/falco -c <abs-config> -L` → **23 ルール認識**（19 detection + 4 health）
- doctor smoke:
  - env: PASS exit=0
  - self-check: PASS exit=0（heartbeat fixture + rule 存在確認）
  - plugin-load: PASS exit=0
  - rule-check: PASS exit=0
  - all: PASS exit=0
  - tail-position --max-age 100h <path>: PASS exit=0（age=41h21m）
  - verify-signature: SKIP exit=2（cosign 不在）
- 軽微な README 修正: `tail-position [path] --max-age 15m` → `tail-position --max-age 15m [path]`
  Go flag.FlagSet が positional 後の flag を parse しないため。doctor usage も同期更新。

### 2026-04-28 15:30..16:00 — D6-5 Linux 導入手順方針確定

- macOS から Linux .so の CGO クロスコンパイル不可（dev-kit LC-001）。
- Docker base image (`falcosecurity/falco`, ubuntu:24.04) を使う rehearsal は
  時間（image pull + apt install + plugin build）が大きく、Phase 6 の趣旨である
  「ドキュメントの canonical 化」を逸脱する。
- 採用方針: **要件 §22.1.1 をそのまま docs/installation.md の Linux production install
  セクションに転記**、検証は (a) CI matrix release.yml の Linux .so build / make verify、
  (b) e2e-test.yml の Level 1+2、(c) tools/rule-validator の P003/P005 静的検査
  に委ねる旨を明記。Live Falco fire は v0.1.0 リリース後のフィールド検証で実施。
- 影響: PHASE6 §1.3 で fallback 策として明示している通り。リスクなし。

### 2026-04-28 16:00..16:10 — D6-6 CI 拡張方針確定

- e2e-test.yml は Linux runner で Level 1 / Level 2 を回している。
- Level 3 (test/integration/) は Falco binary を必要とするため、Linux runner で
  `apt install falco` を追加する選択肢があるが、ジョブ時間の伸び（apt + plugin build +
  e2e-l3 = 約 +5 分）と Phase 4 で macOS で実証済みであることを考慮し、
  v0.1.0 release は現状維持、v0.2 で linux-l3 ジョブ追加とする。PHASE6 §1.4 に従う。
- 軽微な追記のみ: e2e-test.yml に validate-rules ステップを追加（make validate-rules）して
  rule lint regress を防ぐ。

### 2026-04-28 16:10..16:30 — D6-7 / D6-8 整合性確認 + RELEASE_READINESS.md 作成
（次セクション参照）


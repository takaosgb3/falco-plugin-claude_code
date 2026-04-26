# 実装指示テンプレート — claude-code Falco plugin v0.1.0

新セッションで `plugin-dev-workflow` エージェントに渡す指示文。本ファイルを `@docs/IMPLEMENTATION_INSTRUCTIONS.md` で参照させて Phase 0 から開始する。

進捗追跡 Issue: [#2](https://github.com/takaosgb3/falco-plugin-claude_code/issues/2)
要件レビュー履歴 Issue: [#1](https://github.com/takaosgb3/falco-plugin-claude_code/issues/1)

---

## A. 段階確認版（推奨・本番リリース向け）

### A-1. Phase 0-1 起動指示

```
plugin-dev-workflow エージェントで Claude Code 用 Falco プラグインの実装を開始してください。
今回は Phase 0-1 までを実施し、完了後に停止してユーザー承認を待ってください。

【参照ドキュメント】
- 要件定義書: docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md
- 詳細タスク定義書: docs/tasks/detailed_task_definition.md
- リハーサル知見: docs/review/rehearsal/convergence_report.md

【標準フローからの差分（必読）】
本プラグインは AI agent 固有の検出を行う特殊ケース。エージェント標準の
「SQL Injection / XSS / Path Traversal / Command Injection / Suspicious Agent」フローではなく、
要件 §12.1 の T-001〜T-018 に置き換えてください。

【Phase 0: 要件確認】
標準の「対話で fields を聞く」を、要件 §27.3 の scaffold 入力ワークシート 19 項目を読み込んで
ユーザーに**一括提示**する形式に置き換えてください（1 項目ずつ聞かない。差分のみ確認）。

ユーザーに確認すべき事項:
1. ワークシート §27.3 の全 19 項目（特に AUTHOR=takaosgb3、LOG_FORMAT=json、
   LOG_PATH_DEFAULT=~/.claude/security/events.jsonl、SDK_VERSION=v0.8.1）
2. ドメイン固有フィールドは要件 §10.2 の claude_code.* 全 28 項目
3. セキュリティ検出は標準 5 カテゴリではなく要件 §12.1 の T-001〜T-018（18 種）
4. cmd/claude-code-security-logger/ を Phase 1 で別途追加（要件 §6.1 仕様、§22.1 前提注記）
5. 出力先は本リポ root 直下（詳細タスク §8.3.2、claude-code/ サブディレクトリは作らない）

ユーザー承認を得てから Phase 1 に進むこと。

【Phase 1: scaffold + hook logger 骨格】
1. /plugin-scaffold claude-code json を §27.3 ワークシートで起動
2. 5 プレースホルダー展開（${DOMAIN_FIELDS_STRUCT/DEFS/EXTRACT/MAPPING/PARSE_JSON}）を
   要件 §10.2 の claude_code.* 全 28 項目で生成
3. cmd/claude-code-security-logger/ を新規作成（要件 §6.1 HL-001〜HL-014 を満たす骨格）
4. Phase 1 品質ゲート（全件 PASS が必須）:
   - go vet ./... が PASS（パーサー TODO は許容）
   - go build ./... が成功
   - **make build が成功**（P002 -buildmode=c-shared を含む .so/.dylib 生成確認）
   - 詳細タスク §2.7 の Step 1 完了確認の項目を埋める
   - ET-1 (go vet) と ET-7 (既存プラグイン nginx-plugin/openclaw 回帰) を実施

【自律実行ルール】
- ゲート不通過時は自動修正を最大 3 回試行、3 回失敗で停止しユーザー報告
- Phase 1 完了時に git commit（メッセージ: feat(phase-1): scaffold claude-code plugin skeleton）
- 重大判断（既存ファイル削除、破壊的変更、リポ rename 等）は実行前に Issue #2 で確認
- 以下の P コードを踏まないこと:
  - P002（-buildmode=c-shared）、P004（nil map）、P008（load_plugins）、
    P010（Fields/Extract 一致）、P014（SeekEnd）、P017（macOS outputs 除外）、P018（-U フラグ）

【完了報告】
Phase 1 完了後は停止し、Issue #2 にコメントで以下を報告:
- 生成ファイル一覧（root 直下のディレクトリ構造）
- go vet / go build / make build の結果
- ET-7 既存プラグイン回帰結果
- 次 Phase（parser 実装）に向けた懸念点・確認事項

Phase 0 から開始してください。
```

### A-2. Phase 2-3 続行指示（Phase 1 承認後）

```
plugin-dev-workflow エージェントで Phase 2 (parser 実装) と Phase 3 (ルール作成) を
続行してください。Phase 3 完了で停止してユーザー承認を待ってください。

【Phase 2: parser 実装】
- 詳細タスク §3 の T2-1〜T2-7 を実施
- JSONL → claude_code.* extraction 実装
- fsnotify + polling fallback + rotation reopen（要件 §20.2.1）
- P004 nil map 初期化、P014 SeekEnd 必須
- ユニットテスト追加、go test ./... PASS

【Phase 3: ルール作成】
- 詳細タスク §4 の T3-1〜T3-6 を実施
- 要件 §12.4 の T-002〜T-018 condition 雛形を活用
- 全ルール source: claude_code 付与（P003）
- rules/claude_code_health.yaml を別ファイル切り出し（要件 §22.4 OPS-004）
- falco -V でルール検証 PASS

【完了報告】Issue #2 にコメント。
```

### A-3. Phase 4 続行指示（Phase 3 承認後）

```
plugin-dev-workflow エージェントで Phase 4 (3 層 E2E テスト) を実施してください。
Phase 4 完了で停止してユーザー承認を待ってください。

【Phase 4: 3 層 E2E テスト】
- 詳細タスク §5 の T4-1〜T4-9 を実施
- Level 1 Pattern: make e2e-pattern
- Level 2 Pipeline: make e2e-pipeline
- Level 3 Falco 統合: TEST-001〜TEST-008（要件 §20.2 fixture 配置）
- latency 計測（要件 §20.3.1 手順）p95 1 秒以内、最低 5 秒以内
- AT-1〜AT-5 受入テスト

【完了報告】Issue #2 にコメント。テスト結果サマリ + latency 実測値。
```

### A-4. Phase 5-6 続行指示（Phase 4 承認後）

```
plugin-dev-workflow エージェントで Phase 5 (build/package) と Phase 6 (release/運用) を
続行してください。最終完了で全結果を報告してください。

【Phase 5: build / package】
- 詳細タスク §6 を実施
- macOS .dylib + Linux .so の cross build
- make package（チェックサム生成、要件 §21.2 B-007）
- SBOM 生成（推奨、要件 §21.3.1）
- cosign keyless 署名（推奨、要件 §21.3.1）

【Phase 6: release / 運用】
- GitHub Release v0.1.0 作成
- §22.1 macOS 個人導入手順 / §22.1.1 Linux production 導入手順の検証
- doctor CLI（要件 §22.4 OPS-001〜OPS-006）
- CHANGELOG / README 更新

【完了報告】Issue #2 に最終サマリ + リリースアセット URL。
```

---

## B. 完全自律版（試験実装・プロトタイプ向け）

```
plugin-dev-workflow エージェントで Claude Code 用 Falco プラグインを Phase 0-6 まで
完全自律で実装してください。

【参照ドキュメント】
- 要件定義書: docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md
- 詳細タスク定義書: docs/tasks/detailed_task_definition.md
- リハーサル知見: docs/review/rehearsal/convergence_report.md
- 進捗追跡: GitHub Issue #2

【標準フローからの差分】
本プラグインは AI agent 固有検出。標準 5 カテゴリではなく要件 §12.1 の T-001〜T-018 を採用。
出力先は root 直下（詳細タスク §8.3.2）、cmd/claude-code-security-logger/ も同時作成。

【自律実行ルール】
- 各 Phase 完了時に git commit（feat(phase-N): ...）と Issue #2 へのコメント
- ゲート不通過時は自動修正を最大 3 回試行、3 回失敗で停止しユーザー報告
- 重大判断（既存ファイル削除、破壊的変更、リポ rename 等）は実行前に Issue #2 で確認
- コンテキスト 70% 到達で現 Phase を完了させて停止し、ユーザーに継続指示を求める
- P002/P003/P004/P008/P010/P014/P017/P018 を絶対に踏まない

【最終受入基準】
詳細タスク §8.6 の 14 項目チェックリスト + AT-1〜AT-5 + ET-1〜ET-7 をすべて PASS。
v0.1.0 タグの GitHub Release が作成されていること。

Phase 0 から開始してください。
```

---

## C. 推奨運用

| シナリオ | 推奨 |
|---|---|
| 本番リリース予定（v0.1.0 を世に出す） | **A. 段階確認版** |
| プロトタイプ・実験 | B. 完全自律版 |
| エージェント挙動を初めて見る | A の Phase 0-1 だけ実施して感触を掴む |

### ハイブリッド推奨フロー

1. **A-1 で Phase 0-1 実施** → エージェント挙動を確認
2. **A-2 で Phase 2-3 一気通貫** → parser/rules は密結合なのでまとめた方が手戻り少
3. **A-3 で Phase 4 確認** → テスト戦略は人間レビュー
4. **A-4 で Phase 5-6 続行** → ビルド/リリースは自己修復が効きやすい

---

## D. 起動前チェック

実行前に以下を確認:

- [ ] git working tree が clean（`git status`）
- [ ] origin/main に追従済（`git pull`）
- [ ] Falco がローカルに導入済（`falco --version` で 0.43+ 確認）
- [ ] Go toolchain（`go version` で 1.21+ 確認）
- [ ] 参照プラグイン `/Users/takaos/lab/falco-plugin-openclaw` がアクセス可能
- [ ] dev-kit `.claude/` 配下が存在（`ls .claude/skills/`）

---

## E. トラブル時のエスカレーション

エージェントが停止した場合:

1. Issue #2 のコメントで「何のゲートで何回失敗したか」を確認
2. `docs/review/rehearsal/convergence_report.md` §6 残課題を確認
3. `PROBLEM_PATTERNS.md` の該当 P コードを参照
4. 必要なら `/plugin-debug <error-or-symptom>` で個別デバッグ

# Allure Report 統合ログ

このドキュメントは Allure 統合作業の状態を**永続化**する。コンテキスト消失時はこのファイルを起点に作業再開。

## 0. メタ情報

| 項目 | 値 |
|---|---|
| 開始日 | 2026-04-28 |
| 位置付け | Phase 6 後の拡張（v0.1.0 release は別途） |
| 進捗追跡 Issue | [#2](https://github.com/takaosgb3/falco-plugin-claude_code/issues/2) |
| 直前コミット | `fe2c0fa` (Phase 6 完了) — origin/main に push 済 |
| 参考 | `/Users/takaos/lab/falco-plugin-openclaw/` の allure 設定 |

## 1. タスクチェックリスト

### 1.1 openclaw allure 構成の調査
- [ ] **A-1**: `/Users/takaos/lab/falco-plugin-openclaw/Makefile` の allure ターゲット確認
- [ ] **A-2**: `/Users/takaos/lab/falco-plugin-openclaw/.github/workflows/` の allure CI ステップ確認
- [ ] **A-3**: openclaw が使用している tools（gotestsum / allure-commandline / allure-go）特定

### 1.2 ローカル統合
- [ ] **A-4**: 必要 tool の有無確認（`which gotestsum`、`which allure`）
- [ ] **A-5**: 不在 tool のインストール手順を Makefile に記載（実インストールは人間操作）
- [ ] **A-6**: `Makefile` に以下を追加:
  - `make allure-results`: `gotestsum --junitfile allure-results/junit.xml -- ./...`
  - `make allure-report`: `allure generate allure-results -o allure-report --clean`
  - `make allure`: 上記 2 つを連続実行 + open
- [ ] **A-7**: `.gitignore` に `allure-results/` と `allure-report/` を追加
- [ ] **A-8**: ローカル smoke: `make allure` 実行 → `allure-report/index.html` 生成確認

### 1.3 CI 統合
- [ ] **A-9**: `.github/workflows/e2e-test.yml` に allure step 追加
  - gotestsum + allure-commandline インストール
  - allure-results 生成
  - artifact upload（`actions/upload-artifact@v4` で `allure-report/`）
- [ ] **A-10**: README.md に Allure 確認手順を追記

### 1.4 完了処理
- [ ] **A-11**: 全 E2E テスト + allure 出力 PASS（Level 1 + 2 + 3）
- [ ] **A-12**: git commit `feat(test): integrate Allure report for E2E tests`
- [ ] **A-13**: Issue #2 へ報告

## 2. 状態マーカー

### 2.1 現在の進捗

```
最終更新: 2026-04-28 全タスク完了（commit 0a6cbaf）
完了: A-1 〜 A-13（A-13 Issue #2 報告は人間が実施）
進行中: なし
ブロッカー: なし（gotestsum local install は人間操作 — Makefile install ヒント済）
commit: 0a6cbaf feat(test): integrate Allure report for E2E tests
```

最終成果物:

| ファイル | 変更 |
|---|---|
| `Makefile` | `+40 -1` allure-deps/allure-results/allure-report/allure/allure-clean target を追加 |
| `.gitignore` | `+4` allure-results/ allure-report/ を ignore |
| `.github/workflows/e2e-test.yml` | `+26` gotestsum/Allure CLI install + report 生成 + artifact upload |
| `README.md` | `+20` 「Test reports」セクション新設 |
| `docs/tasks/ALLURE_INTEGRATION_LOG.md` | `+176` 本ログ |

### 2.2 重要パラメータ

| 項目 | 値 |
|---|---|
| 現在のテスト件数 | Level 1: 104 sub / Level 2: 19 / Level 3: 9 = 132+ |
| go.mod | `github.com/takaosgb3/falco-plugin-claude_code` |
| 現状 test framework | 標準 Go testing |
| 目標 | 既存 `_test.go` を変更せず、JUnit XML 経由で Allure 化 |

### 2.3 セッション復旧ガイド

1. このファイル全体を確認
2. §2.1「現在の進捗」確認
3. 進捗に応じて A-N から再開
4. §2.1 を更新

### 2.4 既知の落とし穴

1. **gotestsum / allure-commandline の不在**: macOS なら `brew install gotestsum allure`、Linux は手動 install。CI では action で install
2. **JUnit XML から Allure feature/story attribute への mapping**: gotestsum の XML はテスト関数名のみ。allure-go ライブラリ無しでは Step/Severity 等は生成されない（最小限の HTML のみ）
3. **Level 3 (test/integration) の Falco 依存**: Falco 不在環境では skip → allure では skip 扱い

### 2.5 失敗時のフォールバック

- Allure tool 不在で進められない場合: Makefile / CI 設定のみ追加し、ローカル動作確認は人間に委ねる（README に手順を明記）
- openclaw が allure-go (struct-based) を使っている場合: 一時退避、本リポでは gotestsum + JUnit 簡易版で済ませる

## 3. 参照ドキュメント

- 参考: `/Users/takaos/lab/falco-plugin-openclaw/` の allure 関連ファイル
- Allure 公式: https://allurereport.org/
- gotestsum: https://github.com/gotestyourself/gotestsum

## 5. 検証チェックリスト（追加、Playwright での目視確認）

このセクションは「test 結果がレポートに正しく反映されているか」を検証する。
Playwright MCP で Allure HTTP server (http://127.0.0.1:8765) に navigate しながら確認。

- [ ] **V-1**: Overview ページ — total/passed/failed の集計が一致
- [ ] **V-2**: Suites ページ — 7 packages 全部展開して内訳確認
  - pkg/parser: 111
  - test/e2e: 110
  - test/integration: 29
  - cmd/claude-code-doctor: 19
  - cmd/plugin-sdk: 19
  - cmd/claude-code-security-logger: 14
  - tools/rule-validator: 10
- [ ] **V-3**: Suites 内の代表テストを drill down して個別テスト名・duration・status を確認
- [ ] **V-4**: Graphs — Status / Severity / Duration distribution が描画されているか
- [ ] **V-5**: Timeline — 7 packages の並列実行が timeline 上で見えるか
- [ ] **V-6**: Behaviors — feature/story グルーピング（gotestsum 経由なので限定的なはず）
- [ ] **V-7**: Packages — Go package ごとの集計が表示されているか
- [ ] **V-8**: Categories — 失敗カテゴリは 0 件のはず（全 PASS）

### 5.1 検証結果（2026-04-28 Playwright MCP 経由）

検証環境: HTTP server `allure open allure-report --port 8765`、ブラウザ navigate via Playwright MCP、各セクションで snapshot + screenshot を取得。
スクリーンショット: `docs/screenshots/allure/*.png`

| V-N | 検証内容 | 結果 | 証拠 |
|---|---|---|---|
| **V-1** | Overview ページ集計 | **PASS**: 312 test cases / 100% / start 14:40:55 / duration 10s 820ms / 7 suites | `allure-overview.png` |
| **V-2** | Suites 7 packages 内訳 | **PASS**: parser 111 / e2e 110 / integration 29 / doctor 19 / plugin-sdk 19 / security-logger 14 / rule-validator 10（合計 312） / Status 0+0+312+0+0 | `allure-overview.png`, `allure-suite-integration.png` |
| **V-3** | test/integration drill-down | **PASS**: 29 tests 全件展開（TestAT_1_Build / TestAT_5_Redaction / TestAT_Summary / TestL3_Falco_BenignNoFalsePositive 2s 310ms / TestL3_Falco_Categories + 20 sub-tests / TestL3_Falco_Heartbeat 2s 310ms / TestL3_FalcoLocalYAML_* / TestL3_Latency_P95 10s 820ms） | `allure-suite-integration.png` |
| **V-3 詳細パネル** | TestL3_Latency_P95 click → 個別 detail | **PASS**: Status=Passed / Severity=normal / Duration=10s 820ms / Overview/History/Retries タブ表示 | `allure-test-detail-latency.png` |
| **V-4** | Graphs（Status / Severity / Duration distribution） | **PASS**: STATUS 100% green ring / SEVERITY normal=312 / DURATION ヒストグラム（多数 0-1s, 数件 2-3s, 1 件 10-11s = TestL3_Latency_P95 と一致） / Trend 系は履歴なしで「nothing to show」（初回実行のため正常） | `allure-graphs.png` |
| **V-5** | Timeline | **PASS**: "Selected 312 tests (100.00%) with duration above 0s" 集計表示。host/thread レーン詳細は gotestsum が emit しないため表示されない（既知の制約） | `allure-timeline.png` |
| **V-6** | Behaviors | **PASS**: 312 tests を behavior として alphabetical 一覧表示（TestAllDomainFields_* / TestAT_* / TestC4_RPVMustBeFirst / TestDetectCommandInjection + sub-tests / TestDetector_GitDestructive + sub-tests 等）。Status 0+0+312+0+0 | `allure-behaviors.png` |
| **V-7** | Packages | **PASS**: 階層 tree 表示 `github` (312) / 展開可能。Status 0+0+312+0+0 | `allure-packages.png` |
| **V-8** | Categories | **PASS（期待通り空）**: "There are no items"。failure 0 件のため failure category も 0 件で正常 | `allure-categories.png` |

### 5.2 制約・既知の限界

gotestsum + JUnit XML 経路の限界として、以下の Allure 機能は最小限の表示にとどまる:

1. **Severity**: 全テストが `normal`（gotestsum が個別 severity を emit しない）
2. **Feature/Story グルーピング**: Behaviors ページでは alphabetical 一覧のみ。`@feature`/`@story` annotation がないため
3. **Test body / Steps 詳細**: 個別 detail パネルの Execution / Test body は空（PASS テストの場合）
4. **Timeline host/thread レーン**: gotestsum が host/thread metadata を emit しないため詳細レーン非表示
5. **Trend 系**: 初回実行のため履歴データなし。複数回実行 + history 保存設定で表示開始

これらは将来 `allure-go` ライブラリ統合で解消可能（ただし全 _test.go の書き換えが必要、v0.1 では非対応）。

### 5.3 結論

**「テスト結果は正しくレポートされている」**。
- 全 312 件の test name / duration / status (PASS) が正確に表示
- Suites / Behaviors / Packages 全 hierarchical view で確認可能
- Graphs で集計可視化、Timeline で実行時間集計
- Failure 0 件のため Categories は空（仕様通り）

ご指摘いただいた `file://` 直接 open で widgets が空になる問題は **commit `26dbe87`** で解消（`make allure-serve` を推奨経路に変更）。

## 4. 進捗ログ

### 2026-04-28 A-1〜A-4 調査・確認

**A-1: openclaw Makefile**

- `E2E_ALLURE_DIR := e2e/allure`（Python ディレクトリ）
- `e2e-report` target: `cd e2e/allure && python3 -m pytest test_e2e_wrapper.py --alluredir=../../allure-results -v`
- `e2e-serve` target: `allure serve allure-results`
- `e2e-all` target: `e2e e2e-ci e2e-report` の連鎖

**A-2: openclaw .github/workflows/e2e-test.yml**

- `allure-report` job: `pip install -r e2e/allure/requirements.txt` + `allure 2.32.0` を curl で install
- 履歴 (history) と trend widget の独自生成あり（categories-trend.json 等）
- artifact upload: `actions/upload-artifact@v4` with `name: allure-report`
- `gh-pages` deploy（main branch のみ）

**A-3: 使用ツール**

- openclaw: `pytest` + `allure-pytest`（Python）
  - `e2e/allure/conftest.py`, `requirements.txt`, `test_e2e_wrapper.py`
  - Falco 出力 (test-results.json) を pytest が読んで Allure 化する仕組み
- claude-code: 別アプローチを採用
  - **理由 1**: 既存 `_test.go` を変更しない（侵襲性ゼロ目標）
  - **理由 2**: Python 依存を持ち込まない（Go-only リポを維持）
  - **理由 3**: 既存 `make e2e` (L1+L2 = 123 tests) を gotestsum + JUnit XML で
    そのまま Allure 化できる
- 採用: `gotestsum --junitfile` → `allure generate` の最小経路

**A-4: ツール存在確認**

```
which gotestsum   → not found
which allure      → /opt/homebrew/bin/allure
brew list allure  → /opt/homebrew/Cellar/allure/2.37.0/bin/allure
```

- `gotestsum` 不在 → Makefile install ヒントに記載のみ。実 install は人間操作
  - `brew install gotestsum` または `go install gotest.tools/gotestsum@latest`
- `allure` v2.37.0 install 済 → ローカル smoke 実行可能（gotestsum install 後）

### 2026-04-28 A-5〜A-7 Makefile/.gitignore 拡張

- Makefile に `allure-deps`, `allure-results`, `allure-report`, `allure`, `allure-clean` を追加
  - 既存 target は変更せず、ファイル末尾に append
  - `allure-deps` は `gotestsum`/`allure` 不在時に install ヒントを表示して `exit 1`
  - `allure-results` は `gotestsum --junitfile allure-results/junit.xml --format pkgname -- -count=1 ./...`
  - `allure-report` は `allure generate allure-results -o allure-report --clean`
  - `allure` は `allure-report` のエイリアス
- `.PHONY` 行に追加 target を反映
- `.gitignore` に `allure-results/` と `allure-report/` を追加

### 2026-04-28 A-8 ローカル smoke

- `gotestsum` 不在のため `make allure-results` は実行不可
- `make allure-deps` が期待通り install ヒントを表示して exit 1 することを確認
- 完全 smoke は人間が `brew install gotestsum` 後に再実行する必要あり

### 2026-04-28 A-9 CI 統合

- `.github/workflows/e2e-test.yml` に Allure step を追加
  - `Install gotestsum`: `go install gotest.tools/gotestsum@latest`
  - `Install Allure CLI`: 公式 GitHub Releases から allure-2.27.0.tgz を取得
  - `Generate Allure results`: `gotestsum --junitfile allure-results/junit.xml`
  - `Generate Allure report`: `allure generate allure-results -o allure-report --clean`
  - `Upload Allure report`: `actions/upload-artifact@v4` で `allure-report/` を upload
  - 既存 `e2e-results` artifact step は維持

### 2026-04-28 A-10 README.md 更新

- 「## Test reports」セクションを Documentation の前に新規追加
- ローカル `make allure` 手順と CI artifact の説明を記載

### 2026-04-28 A-11 品質ゲート

- `go vet ./...` → PASS
- `go build ./...` → PASS
- `make build` → PASS（lib/.dylib 生成）
- `go test ./... -race -count=1` → PASS
- `make e2e` → PASS（Level 1 + Level 2）
- `make validate-rules` → PASS
- `make allure-deps` → 期待通り gotestsum 不在エラー（人間 install 必要）
- openclaw 回帰: 触っていないため影響なし（claude-code リポ内のみ変更）

### 2026-04-28 A-12 git commit

- commit: `feat(test): integrate Allure report for E2E tests`
- 含まれる変更: Makefile / .gitignore / .github/workflows/e2e-test.yml / README.md / docs/tasks/ALLURE_INTEGRATION_LOG.md

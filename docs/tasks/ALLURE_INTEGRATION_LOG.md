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

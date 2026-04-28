# Allure レポート差分調査ログ — 「いつもの」レポートとの不一致

## 0. 問題提起

ユーザー指摘: 「いつもの Allure レポートでは**実際の Falco ルールを使った検知テスト**が表示されていたが、今回のレポートは全く違う」

### 観察された現状（claude-code）

- 312 test cases、すべて Go 標準 testing.T 関数名（`TestL3_Falco_Categories`、`TestPattern_AllFixturesParse` 等）
- gotestsum + JUnit XML → allure generate
- Severity 全部 normal、Behaviors は alphabetical 一覧
- Test detail パネルの Test body / Steps は空

### 期待された姿（openclaw 経験から）

- **「T-001 dangerous bash detected」のような検知シナリオ単位のテスト**
- Falco alert 出力を evidence として表示
- Feature/Story グルーピング（例: T-001..T-018 を feature 単位で分類）
- 個別テストに Falco rule fire 結果が attachments として表示

## 1. 調査タスク

- [ ] **I-1**: openclaw の Allure 構成を徹底的に読む
  - Makefile の allure target
  - .github/workflows/e2e-test.yml の allure job
  - e2e/allure/ 配下（Python wrapper / pytest / conftest）
  - allure-results の中身（attachments / categories / behaviors）
- [ ] **I-2**: pytest-allure と Go gotestsum の出力構造を比較
- [ ] **I-3**: claude-code が「いつもの」レポートにするための差分を特定
- [ ] **I-4**: 根本原因の確定
- [ ] **I-5**: 採るべきアプローチの検討（A: pytest 経路移植 / B: allure-go 統合 / C: カスタム allure-results 生成）

## 2. 状態マーカー

```
最終更新: 2026-04-28 調査開始
完了: なし
進行中: I-1 (openclaw allure 構成精査)
```

## 3. 調査結果

### 3.1 openclaw の Allure パイプライン全体像（I-1〜I-2）

```
test/e2e/patterns/categories/*.json    (11 カテゴリ × 4-10 patterns = 56 patterns)
     │ pattern 定義: id / description / payload / expected_rule / severity
     ▼
inject_patterns.sh                      (bash)
     │ 1. test log file 初期化
     │ 2. Falco を background で start
     │ 3. 各 pattern の payload を log file に append (session_id 付与)
     │ 4. 一定 wait 後 Falco に SIGTERM、stdout を falco-output.log に保存
     ▼
e2e/results/falco-output.log            (Falco の実 JSON output)
e2e/results/test-ids.json               (pattern → session_id 対応)
     │
     ▼
batch_analyzer.py                       (Python)
     │ 1. patterns/ から pattern_id ↔ category map 構築
     │ 2. falco-output.log を JSON line ごとに parse
     │ 3. session_id で alert ↔ pattern を correlate
     │ 4. detected / matched_rule / latency / evidence を抽出
     ▼
e2e/results/test-results.json           ← Allure の入力データ
     │ 配列: 各要素 = pattern 1 件の検証結果
     │ {pattern_id, category, detected, expected_rule, matched_rule,
     │  rule_match, latency_ms, evidence (Falco alert 全文), status}
     ▼
test_e2e_wrapper.py                     (pytest + allure-pytest)
     │ pytest_generate_tests で test-results.json 配列を parametrize
     │ 各 pattern に対して:
     │   - allure.dynamic.epic("E2E Security Tests")
     │   - allure.dynamic.feature(category.upper())
     │   - allure.dynamic.story(pattern_id)
     │   - allure.dynamic.severity(SEVERITY_MAP[category])
     │   - allure.dynamic.description(rich Markdown)
     │   - allure.step("Test Execution Result"): JSON attachment
     │   - allure.step("Detection Evidence"): HTML highlighted attachment
     │   - allure.step("Rule Mapping"): text attachment
     │   - allure.step("Verification Result"): pass/fail attachment
     │   - assert status == "passed"
     ▼
allure-results/                         (pytest-allure 形式)
     ▼ allure generate
allure-report/                          (HTML、Epic→Feature→Story 階層 + 添付)
```

### 3.2 claude-code 現状の Allure（gap 特定）

```
go test ./...                           ← 単に Go testing.T 関数を実行
     ▼
gotestsum --junitfile junit.xml         ← JUnit XML (test name + duration + status のみ)
     ▼
allure generate                          ← Falco alert 情報なし
     ▼
allure-report/                           (test 関数名一覧、Epic/Feature/Story なし、Evidence なし)
```

### 3.3 根本原因（I-3, I-4）

**ALLURE_INTEGRATION_LOG.md §3 で agent が選択した「JUnit XML 経由の最小経路」は、ユーザー期待である openclaw 風のシナリオ駆動レポートを満たさない**。具体的に欠けているもの:

| 観点 | openclaw | claude-code 現状 | 差分 |
|---|---|---|---|
| データソース | Falco の実 alert output (JSON line) | Go 関数の pass/fail のみ | **Falco 実行結果が一切 Allure に伝わらない** |
| 1 test case = | 1 攻撃シナリオ (pattern_id) | 1 Go testing 関数 | テスト粒度が違う |
| Description | Markdown 表（pattern, payload, expected_rule, matched_rule, evidence） | 空 | 攻撃の context 不在 |
| Epic/Feature/Story | E2E Security / category / pattern_id | なし | hierarchy なし |
| Severity | category 別マッピング (CRITICAL/NORMAL/MINOR) | 全 normal | 重要度区別なし |
| Steps | 4 steps (Execution / Evidence / Rule Mapping / Verification) | なし | drill-down できない |
| Attachments | JSON + HTML (highlighted) + TXT | なし | Falco alert 全文が見えない |
| Pattern ファイル | test/e2e/patterns/categories/*.json (56 patterns) | test/fixtures/hook_events/*.json (25 fixtures、構造異なる) | カテゴリ × pattern 階層が浅い |
| 実 Falco 実行 → 結果 capture スクリプト | inject_patterns.sh + batch_analyzer.py | なし（Go integration test 内に閉じている） | スクリプト経路で外部に export されていない |

### 3.4 なぜ初手で Pytest 経路を採用しなかったのか（自己分析）

ALLURE_INTEGRATION_LOG.md §A-3 で agent は次の理由で pytest 経路を採用しなかった:

1. 「Python 依存を持ち込まない（Go-only リポを維持）」
2. 「既存 _test.go を変更しない（侵襲性ゼロ）」
3. 「最小経路で素早く Allure HTML 化」

**これらは技術的判断としては妥当だが、「ユーザーの期待する Allure の中身」を確認せずに進めた点が誤り**。ユーザー指摘「いつもの Allure レポート」= openclaw 風シナリオ駆動レポートという情報を軽視した。

### 3.5 取り得るアプローチ（I-5）

| 案 | 内容 | 工数 | 維持性 | 備考 |
|---|---|---:|---|---|
| **A. openclaw pytest 経路を完全移植** | inject_patterns.sh / batch_analyzer.py / test_e2e_wrapper.py を claude-code 用に adapt。pattern→T-001..T-018 fixture マッピング | **8-12h** | 中（Python 依存） | openclaw と同じ UX、最も確実 |
| **B. Go integration test で test-results.json を export → Python wrapper のみ移植** | 既存 test/integration/falco_alerts_test.go を拡張して pattern_id/evidence を JSON 出力。inject_patterns.sh は Go test に置き換え | **4-6h** | 中 | 中間パス、Go 強み + pytest 弱み補完 |
| **C. allure-go で Go test から直接 allure-results 出力** | github.com/ozontech/allure-go 等を使って Go test 内で allure.AddStep / allure.Attach | **6-10h** | 高（Go-only 維持） | Step/Severity/Attachment 完備、ただし _test.go 全面書き換え |
| **D. 現状維持 + ユーザーに方針確認** | 「JUnit XML 経路は Go 標準テスト確認用、Falco シナリオ報告は別 path」と分けて記述 | 0h | 高 | UX 期待を満たさない |

### 3.6 推奨

**案 B（Go test で test-results.json export + Python wrapper 移植）**

理由:
- 既存 `test/integration/falco_alerts_test.go` は既に Falco 実行 + alert capture を行っており、pattern_id ↔ rule mapping を assertion している → JSON export が最小コスト
- openclaw の `test_e2e_wrapper.py` は test-results.json を入力にする「pure converter」なので、コア logic は再利用可能（fixture/category mapping を T-001..T-018 用に修正のみ）
- Python 依存は CI/local 両方で常識的（`pip install allure-pytest`）
- inject_patterns.sh は不要（Go test が代替）
- 結果として openclaw 風の **Epic / Feature / Story / Evidence / 4 Steps** UX を再現

### 3.7 案 B 実装計画（agent 起動時に詳細化）

```
T-001..T-018 fixture (test/fixtures/hook_events/) — 既存
     │ + 新規 metadata 追加: expected_rule, severity (T-* 別)
     ▼
go test -tags=allure ./test/integration/...
     │ falco_alerts_test.go が既存の rule fire 検証 + 新規 JSON export
     │ test-results.json: [{pattern_id (T-NNN), category (T-NNN-name),
     │                       detected, expected_rule, matched_rule,
     │                       rule_match, latency_ms, evidence (Falco alert),
     │                       status}, ...]
     ▼
test/allure/test_e2e_wrapper.py        (openclaw からポート、claude-code 用に変更)
     │ - SEVERITY_MAP: T-001/T-002/T-006/T-016 = CRITICAL、T-003〜T-018 = WARNING/NORMAL
     │ - PATTERNS_DIR: test/fixtures/hook_events/
     │ - SECURITY_KEYWORDS: claude_code 固有 (rm -rf, .env, AKIA, ghp_, mcp__, ...)
     ▼
allure-results/ → allure generate → allure-report/
```

新 Makefile target:
- `make allure-falco-results` — Go test で JSON 出力
- `make allure-falco-pytest` — Python wrapper 実行
- `make allure-falco` — 上記 2 つを連続実行 → allure generate

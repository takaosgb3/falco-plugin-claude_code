# Phase Allure-Falco 実行ログ — openclaw 風シナリオ駆動レポート移植

このドキュメントは「openclaw 風 Allure レポート」実装の作業状態を**永続化**する。
コンテキスト消失時はこのファイルを起点に作業再開すること。

## 0. メタ情報

| 項目 | 値 |
|---|---|
| 開始日 | 2026-04-28 |
| 位置付け | ALLURE_INVESTIGATION_LOG §3.6 推奨案 B の実装 |
| 進捗追跡 Issue | [#2](https://github.com/takaosgb3/falco-plugin-claude_code/issues/2) |
| 直前コミット | `b7ffba9` (ALLURE 調査) — origin/main に push 済 |
| 前提 ALLURE 統合 commits | `0a6cbaf` (initial integration) + `26dbe87` (file:// fix) + `73d7a27` (verification) |
| 参考プロジェクト | `/Users/takaos/lab/falco-plugin-openclaw/e2e/` |

## 1. ゴール（DoD: Definition of Done）

ローカル `make allure-falco` 実行で、以下を満たす Allure HTML レポートが生成されること:

1. **Epic / Feature / Story 階層**
   - Epic: `Claude Code E2E Security Tests`
   - Feature: T-001..T-018 各カテゴリ + benign + heartbeat（計 20 features）
   - Story: 個別 fixture id（例: `T-001-dangerous-bash-rm`、`T-002-secret-exfil-curl` 等）
2. **Severity 別マッピング**（要件 §12.1 priority 準拠、後述 §3.2）
3. **Markdown Description**（pattern info 表 + payload + matched rule + Falco alert evidence）
4. **4 Steps 完備**（Test Execution / Detection Evidence (HTML highlighted) / Rule Mapping / Verification）
5. **Falco alert 全文 attachment**（実測した alert 文字列を含む）
6. CI で artifact upload（既存 e2e-test.yml の workflow を活用）
7. 既存 gotestsum + JUnit 経路との**併存**（後方互換: `make allure` も残す）

## 2. アーキテクチャ

```
test/fixtures/hook_events/<event_name>/<fixture_id>.json
  既存 25 fixture (T-001..T-018 + benign + _heartbeat_)
  各 fixture は schema §10.1 準拠 + _meta block (expected_detection 等)
            │
            ▼
test/integration/falco_alerts_test.go
  Phase 4 で実装済（実 Falco 起動 + alert capture + assertion）
            │
            ▼
            │ +新規: TestL3_Falco_Categories の終了時に
            │       result を struct/JSON で集計
            ▼
test/integration/results/test-results.json
  形式: openclaw test-results.json と互換
  [{pattern_id, category, detected, expected_rule, matched_rule,
    rule_match, latency_ms, evidence (Falco alert 全文), status}, ...]
            │
            ▼
test/allure/test_e2e_wrapper.py
  openclaw test_e2e_wrapper.py からポート + 修正:
  - SEVERITY_MAP を T-001..T-018 用に再定義 (§3.2)
  - PATTERNS_DIR を test/fixtures/hook_events/ に変更
  - SECURITY_KEYWORDS を claude_code 固有に変更 (§3.3)
  - pattern info 取得を fixture _meta block から行う
            │
            ▼
allure-results/                     ← Epic / Feature / Story 完備
            │
            ▼
allure generate
            ▼
allure-report/                       ← 目的の openclaw 風レポート
```

## 3. 実装詳細

### 3.1 Go integration test の JSON export 拡張

ファイル: `test/integration/falco_alerts_test.go`（既存を拡張）

新規追加:
- `type FalcoTestResult struct` — openclaw 互換 schema
- `WriteTestResultsJSON(t *testing.T, results []FalcoTestResult)` — `test/integration/results/test-results.json` 書き出し
- `TestL3_Falco_Categories` の `t.Run` 内部で各 sub-case の result を append
- `TestMain` 終了時にまとめて JSON export

Schema 例:
```json
{
  "pattern_id": "T-001-dangerous-bash-rm",
  "category": "T-001",
  "detected": true,
  "expected_rule": "[CLAUDE_CODE CRITICAL] Dangerous Bash Command",
  "matched_rule": "[CLAUDE_CODE CRITICAL] Dangerous Bash Command",
  "rule_match": true,
  "latency_ms": 0,
  "evidence": "<Falco alert 全文 line>",
  "status": "passed"
}
```

build tag で controlable:
- `go test -tags=allure ./test/integration/...` のときのみ JSON export
- 通常の `go test` は副作用なし

### 3.2 SEVERITY_MAP（要件 §12.1 priority 準拠）

| T-* カテゴリ | priority (§12.1) | Allure Severity |
|---|---|---|
| T-001 dangerous_bash | CRITICAL | `critical` |
| T-002 secret_exfiltration | CRITICAL | `critical` |
| T-003 permission_bypass | CRITICAL | `critical` |
| T-006 hook_disabled | CRITICAL | `critical` |
| T-016 policy_downgrade | CRITICAL | `critical` |
| T-004 permission_update | WARNING | `normal` |
| T-005 settings_modified | WARNING | `normal` |
| T-007 mcp_config_changed | WARNING | `normal` |
| T-008 mcp_tool_suspicious | WARNING | `normal` |
| T-009 sensitive_file_read | WARNING | `normal` |
| T-010 workspace_escape | WARNING | `normal` |
| T-011 git_destructive | WARNING | `normal` |
| T-012 prompt_injection | WARNING | `normal` |
| T-014 tool_storm | WARNING | `normal` |
| T-015 external_fetch_sensitive | WARNING | `normal` |
| T-017 skill_shell | WARNING | `normal` |
| T-018 channel_push | WARNING | `normal` |
| T-013-low agent_risk_low | NOTICE | `minor` |
| T-013-high agent_risk_high | WARNING | `normal` |
| benign | — | `trivial` |
| heartbeat | NOTICE | `minor` |

### 3.3 SECURITY_KEYWORDS（evidence highlight 用、claude_code 固有）

```python
SECURITY_KEYWORDS = [
    # T-001 dangerous bash
    "rm -rf", "rm -f", "/etc/passwd", "/etc/shadow", "chmod -R 777",
    "mkfs.", "dd if=", "shutdown -h", "/dev/sda",
    "curl pipe sh", "| sh", "| bash", "$(whoami)", "`whoami`",
    # T-002 secret exfiltration
    "AKIA", "ghp_", "github_pat_", "xoxb-", "sk-ant-", "sk-",
    "AWS_SECRET_ACCESS_KEY", "id_rsa", "credentials", ".env",
    "BEGIN PRIVATE KEY", "BEGIN RSA PRIVATE KEY",
    # T-003 permission bypass
    "bypassPermissions", "--dangerously-skip-permissions",
    # T-006 hook disabled
    "disableAllHooks",
    # T-009 sensitive file
    ".aws/credentials", ".kube/config", "/root/",
    # T-010 workspace escape
    "../../", "../../../",
    # T-011 git destructive
    "git push -f", "git reset --hard", "git clean -fdx", "rm -rf .git",
    # T-012 prompt injection
    "ignore previous instructions", "ignore all previous",
    # claude_code generic
    "PreToolUse", "PermissionRequest", "ConfigChange", "claude_code",
]
```

### 3.4 fixture pattern info 取得

各 fixture の `_meta` block に既に存在:
```json
{
  "_meta": {
    "fixture_id": "T-001-dangerous-bash-rm",
    "category": "T-001",
    "expected_detection": "rule",
    "expected_event_name": "PreToolUse",
    "notes": "..."
  }, ...
}
```

これを Python wrapper から読み込んで Markdown description に展開。

### 3.5 ファイル構成

新規:
- `test/integration/results/.gitkeep` — directory placeholder
- `test/allure/test_e2e_wrapper.py` — openclaw からポート（修正版）
- `test/allure/conftest.py` — openclaw からポート
- `test/allure/requirements.txt` — `allure-pytest>=2.13`、`pytest>=7`

修正:
- `test/integration/falco_alerts_test.go` — JSON export 関数 + build tag
- `Makefile` — 新ターゲット
- `.gitignore` — `test/integration/results/test-results.json`
- `.github/workflows/e2e-test.yml` — Python install + pytest 実行

### 3.6 Makefile 新ターゲット

```makefile
allure-falco-results:
	go test -tags=allure -count=1 ./test/integration/... -run TestL3_Falco_Categories
	@echo "JSON written to test/integration/results/test-results.json"

allure-falco-pytest: allure-falco-results
	cd test/allure && python3 -m pip install --user -r requirements.txt
	cd test/allure && python3 -m pytest test_e2e_wrapper.py \
		--test-results=../../test/integration/results/test-results.json \
		--alluredir=../../allure-results-falco -v

allure-falco-report: allure-falco-pytest
	allure generate allure-results-falco -o allure-report-falco --clean
	@echo "openclaw-style Allure: make allure-falco-serve"

allure-falco-serve: allure-falco-report
	allure open allure-report-falco

allure-falco: allure-falco-serve

allure-falco-clean:
	rm -rf allure-results-falco allure-report-falco
	rm -f test/integration/results/test-results.json
```

`make allure` (gotestsum 経路) は併存維持。

### 3.7 CI 拡張

`.github/workflows/e2e-test.yml` に追加:

```yaml
      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.12"

      - name: Install allure-pytest
        run: pip install -r test/allure/requirements.txt

      - name: Generate Falco scenario test-results.json
        run: go test -tags=allure -count=1 ./test/integration/... -run TestL3_Falco_Categories
        continue-on-error: true

      - name: Run pytest wrapper
        if: always()
        run: |
          cd test/allure && python3 -m pytest test_e2e_wrapper.py \
            --test-results=../../test/integration/results/test-results.json \
            --alluredir=../../allure-results-falco -v

      - name: Generate Falco-scenario Allure report
        if: always()
        run: allure generate allure-results-falco -o allure-report-falco --clean

      - name: Upload Falco-scenario Allure report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: allure-report-falco
          path: allure-report-falco/
          retention-days: 30
```

既存の `allure-report` artifact は維持。両方 upload。

## 4. タスクチェックリスト

### 4.1 Go 側 (B-1〜B-5)
- [ ] **B-1**: `test/integration/falco_alerts_test.go` に build-tag `allure` 追加して JSON export 実装
- [ ] **B-2**: `FalcoTestResult` struct と `writeTestResultsJSON` 関数
- [ ] **B-3**: `TestL3_Falco_Categories` の各 sub-case で result 収集
- [ ] **B-4**: `TestL3_Falco_BenignNoFalsePositive` も収集
- [ ] **B-5**: `TestL3_Falco_Heartbeat` も収集

### 4.2 Python 側 (B-6〜B-9)
- [ ] **B-6**: openclaw `test_e2e_wrapper.py` を `test/allure/` にコピー
- [ ] **B-7**: SEVERITY_MAP / SECURITY_KEYWORDS / PATTERNS_DIR を claude_code 用に修正
- [ ] **B-8**: pattern info 取得を fixture `_meta` block から行うよう変更
- [ ] **B-9**: openclaw `conftest.py` をコピー、`requirements.txt` 作成

### 4.3 Makefile / CI / Doc (B-10〜B-13)
- [ ] **B-10**: Makefile に allure-falco* ターゲット追加（既存 allure target は維持）
- [ ] **B-11**: `.gitignore` に results/test-results.json と allure-*-falco/ 追加
- [ ] **B-12**: `.github/workflows/e2e-test.yml` に Python + pytest step 追加
- [ ] **B-13**: README.md の「Test reports」セクションに `make allure-falco` を案内

### 4.4 検証 (B-14〜B-17)
- [ ] **B-14**: `make allure-falco-results` で test-results.json 生成、形式確認
- [ ] **B-15**: `make allure-falco-pytest` で allure-results-falco 生成、Epic/Feature/Story 確認
- [ ] **B-16**: `make allure-falco` でブラウザ表示、**Playwright MCP で全セクション再 verify**
  - V-1〜V-8 を再実行
  - 加えて: 個別 test case の Markdown description 表示、4 Steps の attachment 内容、Severity 色分け、Behaviors の Epic/Feature/Story 階層
- [ ] **B-17**: スクリーンショット 8 枚（最低）を `docs/screenshots/allure-falco/` に保存

### 4.5 完了処理 (B-18〜B-20)
- [ ] **B-18**: 全品質ゲート PASS
- [ ] **B-19**: git commit `feat(test): add openclaw-style Falco scenario Allure report`
- [ ] **B-20**: Issue #2 へ報告

## 5. 状態マーカー

### 5.1 現在の進捗
```
最終更新: 2026-04-28 計画策定完了
完了: ALLURE_INVESTIGATION_LOG §3 (root cause + approach 評価)
進行中: B-1（agent 起動準備）
ブロッカー: なし
```

### 5.2 既知の落とし穴

1. **fixture 内 `_meta` block を parser に渡す前に strip** が必要（既存 `pkg/testutil/fixture.go` 参照）
   - Allure wrapper では `_meta` を残したまま読み込む（pattern info 用）
2. **Falco preempt 動作**（Phase 4 既知）
   - T-013-high → T-003 に preempt、T-017 → T-010 に preempt 等
   - test-results.json では `matched_rule` が expected と異なる場合があり、`rule_match: false` でも `status: passed` (preemption は許容)
   - openclaw の `format_rule_match_status()` で「preempted」状態の表現が必要かも（要検討）
3. **CI 上の Falco install** は既存 e2e-test.yml で済んでいるか？
   - 確認必要: 現状 e2e-test.yml は Linux Falco を install していない（Phase 6 D6-6 で v0.2 へ繰り延べ済）
   - Phase Allure-Falco では Linux CI で実 Falco を起動するため、追加で `apt install falco` が必要
   - **代替**: CI では allure-falco job を skip し、ローカルでのみ実行可能とする方針も
4. **既存 ALLURE_INTEGRATION_LOG / make allure (gotestsum) は併存**
   - 削除しない、`make allure` は Go 開発者用、`make allure-falco` はセキュリティ E2E 用
   - README で 2 つの違いを明記

### 5.3 セッション復旧ガイド

1. このファイル全体を確認
2. §5.1「現在の進捗」確認
3. 進捗に応じて B-N から再開
4. §5.1 と §6 進捗ログを更新

### 5.4 失敗時のフォールバック

- pytest install が CI で失敗 → CI では allure-falco job を `if: matrix.os == 'macos-14'` に絞る、または ローカル only
- pattern info 取得が複雑すぎる → fixture _meta から最小限（pattern_id + category + expected_event_name）のみ取得
- Falco preempt の表現が pytest wrapper で扱えない → `_meta.preempted_by` を fixture に追加して表示

## 6. 進捗ログ

(作業実行中に追記)

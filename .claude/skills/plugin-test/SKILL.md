---
name: plugin-test
description: Falcoプラグインのテスト作成・実行・レポートを支援する。ユーザーが「テストを作成」「テスト実行」「E2Eテスト」「パターンテスト」「パイプラインテスト」「Level 1/2/3テスト」「make e2e」「go test」「テストカバレッジ」「偽陽性テスト」「benign/edge_casesパターン」「テスト結果」と言った場合にトリガーする。3層E2Eテストアーキテクチャ（Level 1 Pattern/Level 2 Pipeline/Level 3 Falco Integration）を管理。plugin_test.go.tmplとe2e_pattern_test.go.tmplからテスト生成。ルール作成にはplugin-rules、ビルドにはplugin-buildを使用すること。
argument-hint: "[action] [test-type]"
---

# テスト作成・実行支援

$ARGUMENTS についてテストの作成・実行を支援します。

## 3層 E2E テストアーキテクチャ

```
Level 1: パターンカバレッジテスト（Falco 不要）
  - セキュリティ検出パターンの網羅テスト
  - テンプレート: e2e_pattern_test.go.tmpl
  - 実行: make e2e-pattern

Level 2: プラグインパイプラインテスト（Falco 不要、CGO_ENABLED=1 必要）
  - プラグインのライフサイクル・データフロー全体テスト
  - テンプレート: plugin_test.go.tmpl
  - 実行: make e2e-pipeline

Level 3: Falco 統合テスト（Falco 必要）
  - Falco にプラグインをロードして実際のアラート発火を検証
  - 実行: make e2e-ci (Linux) / make e2e-native (macOS)
```

## 引数

- **action**: 実行アクション
  - `create`: テストファイルの生成
  - `run`: テストの実行
  - `report`: テスト結果の集計・レポート
- **test-type**: テスト種別
  - `unit`: ユニットテスト（パーサー、セキュリティ検出）
  - `pipeline`: Level 2 パイプラインテスト
  - `e2e`: E2Eテストパターン（Level 1）
  - `all`: 全テスト種別（デフォルト）

## 実行手順

### 1. create: テストファイルの生成

#### 1.1 ユニットテスト生成

テンプレート（`.claude/templates/plugin/parser_test.go.tmpl`）を使用して `pkg/parser/parser_test.go` を生成。

テスト項目:

- **正常系テスト**: 各ログフォーマットのパース確認
  - Combined Log Format
  - Common Log Format
  - JSON Format（対応する場合）
- **異常系テスト**:
  - 空行の処理
  - 不正フォーマットのログ行
  - 巨大行（10KB超）の処理
  - 不完全なログ行
- **セキュリティパターンテスト**:
  - SQL Injection 検出
  - XSS 検出
  - Path Traversal 検出
  - Command Injection 検出
  - Suspicious Agent 検出
- **URLエンコードテスト**:
  - 1段階エンコード（`%27`）
  - 2段階エンコード（`%2527`）
  - 3段階エンコード（`%252527`）

テスト構造例:

```go
func TestParseCombined(t *testing.T) {
    parser := NewParser(Config{Format: "combined"})
    entry, err := parser.Parse(`192.168.1.1 - - [23/Feb/2026:10:15:30 +0900] "GET /api/users HTTP/1.1" 200 1234 "-" "Mozilla/5.0"`)
    assert.NoError(t, err)
    assert.Equal(t, "192.168.1.1", entry.RemoteAddr)
    assert.Equal(t, "GET", entry.Method)
    assert.Equal(t, "/api/users", entry.Path)
    assert.Equal(t, 200, entry.Status)
}

func TestParseInvalidLine(t *testing.T) {
    parser := NewParser(Config{Format: "combined"})
    _, err := parser.Parse("invalid log line")
    assert.Error(t, err)
}

func TestDetectSQLInjection(t *testing.T) {
    detector := NewSimpleSecurityDetector()
    threat := detector.DetectSecurityThreat("GET /api?id=' OR '1'='1")
    assert.Equal(t, ThreatSQLi, threat)
}

func TestDetectURLEncoded(t *testing.T) {
    detector := NewSimpleSecurityDetector()
    // 1段階エンコード
    threat := detector.DetectSecurityThreat("GET /api?id=%27%20OR%201%3D1")
    assert.Equal(t, ThreatSQLi, threat)
}
```

#### 1.2 Level 2 パイプラインテスト生成

テンプレート（`.claude/templates/plugin/plugin_test.go.tmpl`）を使用して `cmd/plugin-sdk/plugin_test.go` を生成。

テスト項目:

- **ライフサイクルテスト**: Init/Open/Close のライフサイクル検証
  - デフォルト設定・カスタム設定での Init
  - バッファサイズ境界値テスト
  - Open() ファイル自動作成、SeekEnd(P014)
  - Close() リソース解放
- **取り込みテスト**: ログ取り込みパイプライン検証
  - 基本ログ取り込み（combined format）
  - 複数ファイル監視
  - GOB ラウンドトリップ
  - Headers 非 nil (P004)
- **性能テスト**: スループット・安定性検証
  - 100 events/sec スループット
  - バッファオーバーフロー非ハング
- **エラー耐性テスト**: エラーハンドリング検証
  - 不正 JSON 設定
  - ファイル削除時

ヘルパー関数（6関数）:
- `initPlugin(t, logPaths)` — プラグイン初期化
- `openAndCleanup(t, plugin)` — Open + Cleanup 登録（`source.Instance` → `*MyInstance` 型アサーション）
- `writeToLog(t, path, line)` — ログ書き込み + fsnotify 待機
- `waitForEvent(t, ch, timeout)` — イベント待機
- `gobEncode(t, event)` — GOB エンコード
- `gobDecode(t, data)` — GOB デコード

実行コマンド:
```bash
# Level 2 テスト
make e2e-pipeline
# または直接実行
go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -count=1 -timeout 120s
```

#### 1.3 E2Eテストパターン生成

テンプレート（`.claude/templates/plugin/e2e_pattern.json.tmpl`）を使用してE2Eパターンを生成。

生成先: `test/e2e/patterns/categories/`

```
test/e2e/patterns/categories/
├── sqli.json              ← SQL Injection パターン（最低4個）
├── xss.json               ← XSS パターン（最低4個）
├── path_traversal.json    ← Path Traversal パターン（最低4個）
├── cmd_injection.json     ← Command Injection パターン（最低4個）
└── suspicious_agent.json  ← Suspicious Agent パターン（最低4個）
```

各カテゴリのパターン形式:

```json
{
  "category": "sqli",
  "patterns": [
    {
      "id": "SQLI_BASIC_001",
      "description": "Basic SQL injection with single quote",
      "payload": "' OR '1'='1",
      "expected_rule": "[${PLUGIN_NAME_UPPER} SQLi] SQL Injection Attempt",
      "attack_type": "sqli",
      "severity": "high",
      "encoding": "none"
    },
    {
      "id": "SQLI_ENCODED_001",
      "description": "URL-encoded SQL injection",
      "payload": "%27%20OR%20%271%27%3D%271",
      "expected_rule": "[${PLUGIN_NAME_UPPER} SQLi] SQL Injection Attempt",
      "attack_type": "sqli",
      "severity": "high",
      "encoding": "url"
    },
    {
      "id": "SQLI_UNION_001",
      "description": "UNION SELECT injection",
      "payload": "1 UNION SELECT username,password FROM users--",
      "expected_rule": "[${PLUGIN_NAME_UPPER} SQLi] SQL Injection Attempt",
      "attack_type": "sqli",
      "severity": "high",
      "encoding": "none"
    },
    {
      "id": "SQLI_DOUBLE_ENCODED_001",
      "description": "Double-encoded SQL injection",
      "payload": "%2527%20OR%201%3D1",
      "expected_rule": "[${PLUGIN_NAME_UPPER} SQLi] SQL Injection Attempt",
      "attack_type": "sqli",
      "severity": "high",
      "encoding": "double-url"
    }
  ]
}
```

要件: 5カテゴリ x 4パターン以上 = 最低20パターン（SC-032, SC-033）

**重要（C-002）**: Suspicious Agentは他の4カテゴリとは検出フローが異なる。`DetectSecurityThreat()` は4カテゴリのみ処理し、`DetectSuspiciousAgent()` はUser-Agentフィールドに対して個別に呼び出す。

### 2. run: テスト実行

#### 2.1 ユニットテスト実行

```bash
# 全テスト実行
go test ./... -v

# パーサーテストのみ
go test ./pkg/parser/... -v

# カバレッジ付き
go test ./pkg/parser/... -v -coverprofile=coverage.out
go tool cover -func=coverage.out
```

#### 2.2 Level 2 パイプラインテスト実行

```bash
# Level 2 テスト（Falco 不要、macOS でも実行可能）
make e2e-pipeline

# 直接実行
go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -count=1 -timeout 120s
```

#### 2.3 Level 1 パターンテスト実行

```bash
# Level 1 テスト（Falco 不要）
make e2e-pattern

# 直接実行
go test ./test/e2e/ -v -race -run TestPattern -count=1
```

#### 2.4 Level 1 + Level 2 combined

```bash
# Falco 不要の全 E2E テスト
make e2e
```

#### 2.5 Level 3 Falco 統合テスト実行

```bash
# Linux CI/CD 環境
make e2e-ci

# macOS ネイティブビルド
make e2e-native
```

### 3. report: テスト結果の集計

テスト実行結果を以下の形式で報告:

```
## テスト結果サマリー

| テスト種別 | 合計 | 成功 | 失敗 | スキップ |
|-----------|------|------|------|---------|
| ユニットテスト | XX | XX | 0 | 0 |
| E2Eパターン | XX | XX | 0 | 0 |

### カバレッジ
- パーサー: XX%
- 全体: XX%

### 失敗テスト（あれば）
- なし
```

## コンテキスト補完情報

### 参照すべきドキュメント

- `pkg/parser/nginx_test.go`: テストの参照実装
- `test/e2e/security-rules/`: E2Eパターンの参照実装
- Serena memory: `e2e_test`
- `.claude/templates/plugin/parser_test.go.tmpl`: パーサーユニットテストテンプレート
- `.claude/templates/plugin/plugin_test.go.tmpl`: Level 2 パイプラインテストテンプレート
- `.claude/templates/plugin/e2e_pattern.json.tmpl`: E2Eパターンテンプレート

### テストディレクトリ構造（Go慣習準拠）

```
pkg/parser/
├── parser.go           # パーサー実装
├── parser_test.go      # ユニットテスト（同一パッケージ）
├── config.go           # 設定
└── regex_simple.go     # セキュリティ検出

test/
├── e2e/
│   ├── patterns/
│   │   └── categories/    # カテゴリ別E2Eパターン
│   └── scripts/
│       └── run-e2e.sh     # E2E実行スクリプト
└── fixtures/
    └── sample_logs/       # テスト用サンプルログ
```

### 過去の失敗パターン（要注意）

1. **CI-005: テスト配置のGo慣習** — ユニットテストはテスト対象と同じパッケージに配置
2. **LC-004: E2Eパターン数の整合性** — 5カテゴリ x 4パターン = 最低20パターン

## 成功基準

| ID | 基準 | 検証方法 |
|----|------|----------|
| SC-030 | ユニットテストがすべてパス | `go test ./pkg/...` |
| SC-031 | パーサーの正常系・異常系テストが存在 | テストファイル検査 |
| SC-032 | E2Eテストパターンが最低20個以上（5カテゴリ x 4パターン以上） | パターンファイル検査 |
| SC-033 | 各攻撃カテゴリに最低4パターン以上 | パターンファイル検査 |
| SC-034 | Level 2 パイプラインテストがすべてパス | `make e2e-pipeline` |
| SC-035 | Level 2 テストに最低14テストケースが含まれる | テストファイル検査 |

## 重要な注意事項

- ユニットテストは `pkg/parser/parser_test.go` に配置（Go慣習準拠）
- E2Eテストパターンは `test/e2e/patterns/categories/` に配置
- テスト実行は `go test ./...` で全テストを実行可能であること
- E2Eテストの完全な実行にはLinux環境+Falcoが必要
- macOS環境ではユニットテストのみ実行可能
- 参照実装のテストは7ファイルに分割されているが、初期生成は単一ファイルで可。成長に伴い分割を推奨

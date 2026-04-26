---
name: plugin-parser
description: ログフォーマットを分析し、パーサーの実装とセキュリティ検出パターンの設定を支援する。ユーザーが「ログをパースしたい」「パーサーを実装」「ログフォーマットの分析」「セキュリティ検出パターン」「SQLi/XSS/PathTraversal/CMDi検出」「JSONパーサー」「auto検出モード」「parseJSON/parseCombined」「正規表現パターン」と言った場合にトリガーする。combined/common/json/auto/customの5フォーマット対応。入力サイズ超過時はtruncate方式（先頭10KB検出）。プラグインの初期生成にはplugin-scaffold、ルール定義にはplugin-rulesを使用すること。
argument-hint: "[log-format] [sample-log-file]"
---

# ログパーサー実装支援

$ARGUMENTS についてログパーサーの実装を支援します。

## 引数

- **log-format**: ログフォーマット名
  - `auto`: 自動検出（JSON は `{` 先頭判定、それ以外はテキストパーサーにフォールバック）
  - `combined`: Apache/nginx Combined Log Format
  - `common`: Common Log Format
  - `json`: JSON フォーマット（デフォルト実装あり: `json.Unmarshal` + `parseTimestamp` ヘルパー）
  - `custom`: カスタムフォーマット（正規表現指定）
- **sample-log-file** (任意): サンプルログファイルのパス

## 実行手順

### 1. Phase 1: ログフォーマット分析

#### 1.1 サンプルログの読み込み・解析

サンプルログファイルが提供された場合:

```bash
# サンプルログの確認（先頭10行）
head -10 ${SAMPLE_LOG_FILE}

# フォーマット自動検出
# JSON判定: 先頭行が { で始まるか
head -1 ${SAMPLE_LOG_FILE} | grep -q '^{' && echo "JSON format detected"

# Combined/Common判定: 正規表現マッチ
head -1 ${SAMPLE_LOG_FILE} | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' && echo "Access log format detected"
```

#### 1.2 フォーマット別パース戦略

| フォーマット | パース方式 | 参照 |
|-------------|----------|------|
| auto | JSON 自動検出 + テキストフォールバック | `parser.go.tmpl` parseAuto() |
| combined | 正規表現（scaffold が生成） | scaffold スキルが domain に応じて生成 |
| common | 正規表現（scaffold が生成） | 同上 |
| json | `encoding/json` Unmarshal + parseTimestamp() | `parser.go.tmpl` parseJSON()（デフォルト実装あり） |
| custom | ユーザー定義正規表現 | ユーザー入力 |

**注意**: 入力サイズ超過時（10KB超）の挙動は「切り詰めて続行」（truncate）方式。`DetectSecurityThreat()` は先頭 10KB 内の脅威パターンを検出する（P020参照）。

#### 1.3 フィールド一覧の抽出

サンプルログからフィールドを自動検出し、ユーザーに確認:

- リモートアドレス（IP）
- タイムスタンプ
- HTTPメソッド
- リクエストパス
- クエリストリング
- HTTPバージョン
- ステータスコード
- レスポンスサイズ
- リファラー
- User-Agent
- カスタムフィールド（ログソース固有）

### 2. Phase 2: パーサーコード生成

#### 2.1 生成ファイル

テンプレート（`.claude/templates/plugin/`）を使用して以下を生成:

1. `pkg/parser/parser.go` ← `parser.go.tmpl`
2. `pkg/parser/config.go` ← `config.go.tmpl`
3. `pkg/parser/regex_simple.go` ← `regex_simple.go.tmpl`
4. `pkg/parser/parser_test.go` ← `parser_test.go.tmpl`

#### 2.2 パーサー構造体の設計

```go
// LogEntry構造体 — パーサーの出力
type LogEntry struct {
    RemoteAddr     string
    RemoteUser     string
    TimeLocal      time.Time
    Method         string
    Path           string
    QueryString    string
    HTTPVersion    string
    Status         int
    BodyBytes      int
    Referer        string
    UserAgent      string
    Request        string              // Method + Path + QueryString + HTTPVersion
    Timestamp      time.Time
    SecurityThreat SecurityThreatType
    Headers        map[string]string
    Raw            string
}
```

**重要**: LogEntry（パーサー出力）と PluginEvent（プラグインイベント）は**別の構造体**。フィールド名・型が異なる場合がある:

- `LogEntry.HTTPVersion` → `PluginEvent.Protocol`（フィールド名変更）
- `LogEntry.Status(int)` → `PluginEvent.Status(uint64)`（型変換）
- `LogEntry.BodyBytes(int)` → `PluginEvent.BytesSent(uint64)`（名前+型変換）
- なし → `PluginEvent.LogPath`（追加フィールド）

`parseLine()` 関数で LogEntry → PluginEvent へのマッピングを行う。

#### 2.3 リクエスト詳細解析

```go
// parseRequest — "GET /path?query HTTP/1.1" を分解
func parseRequest(request string) (method, path, queryString, httpVersion string) {
    parts := strings.SplitN(request, " ", 3)
    if len(parts) >= 1 { method = parts[0] }
    if len(parts) >= 2 {
        uri := parts[1]
        if idx := strings.Index(uri, "?"); idx >= 0 {
            path = uri[:idx]
            queryString = uri[idx+1:]
        } else {
            path = uri
        }
    }
    if len(parts) >= 3 { httpVersion = parts[2] }
    return
}
```

### 3. Phase 3: セキュリティ検出パターン設定

#### 3.1 検出カテゴリの選択

ユーザーに検出したいカテゴリを確認:

| カテゴリ | 検出メソッド | 対象フィールド |
|---------|-------------|-------------|
| SQL Injection | `DetectSecurityThreat()` | request_uri, path, query_string |
| XSS | `DetectSecurityThreat()` | request_uri, path, query_string |
| Path Traversal | `DetectSecurityThreat()` | request_uri, path |
| Command Injection | `DetectSecurityThreat()` | request_uri, path, query_string |
| Suspicious Agent | `DetectSuspiciousAgent()` | User-Agent（専用メソッド） |

**重要（C-002）**: `DetectSecurityThreat()` 集約関数は上位4カテゴリのみ処理。`DetectSuspiciousAgent()` はUser-Agent専用の個別メソッドとして実装する。

#### 3.2 SimpleSecurityDetector の実装方針

```go
// 文字列マッチングベース（ReDoS安全）
// ❌ 正規表現は使用しない
// ✅ strings.Contains / strings.ToLower を使用

// URLデコード処理: 最大3段階
// Stage 1: %27 → '
// Stage 2: %2527 → %27 → '
// Stage 3: %252527 → %2527 → %27 → '

// 入力サイズ制限: 10KB（NFR-021）
const maxInputSize = 10 * 1024
```

#### 3.3 検出パターン例

```go
// SQL Injection パターン
sqliPatterns := []string{
    "' or ", "' and ", "union select", "union all select",
    "1=1", "1'='1", "drop table", "insert into",
    "sleep(", "benchmark(", "waitfor delay",
}

// URLエンコードパターン（P006対策）
sqliEncodedPatterns := []string{
    "%27", "%2527",      // ' (single quote)
    "%20or%20", "%20and%20",
}
```

## コンテキスト補完情報

### 参照すべきドキュメント

- `pkg/parser/nginx.go`: パーサーの参照実装
- `pkg/parser/regex_simple.go`: SimpleSecurityDetector（推奨実装）
- `pkg/parser/regex_safe.go`: SafeSecurityDetector（オプション）
- `pkg/parser/config.go`: パーサー設定の参照実装
- `pkg/parser/nginx_test.go`: テストの参照実装

### 過去の失敗パターン（要注意）

1. **P006: URLエンコード考慮漏れ** — `%27`（single quote）等のURLエンコードパターンを必ず検出対象に含める。最大3段階のデコードを実施
2. **C-001: LogEntryとPluginEventの構造体差異** — フィールド名・型が異なる場合があるため、`parseLine()` でのマッピングを正確に実装
3. **C-002: 検出メソッドの使い分け** — `DetectSecurityThreat()` は4カテゴリ、`DetectSuspiciousAgent()` はUser-Agent専用
4. **ReDoS安全性** — 正規表現の代わりに `strings.Contains` / `strings.ToLower` を使用

## 成功基準

| ID | 基準 | 検証方法 |
|----|------|----------|
| SC-010 | サンプルログが正しくパースできる | パーサーテスト |
| SC-011 | URLデコードが3段階まで対応 | テストケース |
| SC-012 | ReDoS脆弱性がない | 正規表現の静的解析（正規表現不使用を確認） |
| SC-013 | セキュリティ検出パターンが機能する | テストケース |

## 重要な注意事項

- ユニットテストは `pkg/parser/parser_test.go` として生成（テスト対象と同じパッケージ — Go慣習準拠）
- E2Eテストのみ `test/e2e/` に配置
- SimpleSecurityDetector を推奨（ReDoS安全）。SafeSecurityDetector はオプション
- SSRF検出は参照実装に存在しない。初期リリースでは上記5カテゴリに限定
- 入力サイズ制限（10KB）を超えるログ行は検出をスキップ
- `Headers map[string]string` は必ず `make(map[string]string)` で初期化（P004）

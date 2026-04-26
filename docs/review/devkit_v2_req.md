# falco-plugin-dev-kit v2 要件定義書

| 項目 | 内容 |
|------|------|
| 文書ID | REQ-DEVKIT-V2-001 |
| 作成日 | 2026-03-06 |
| ステータス | Draft v5.6 (実装リハーサルレビュー RH-001〜RH-024 反映) |
| 作成元 | falco-plugin-openclaw 開発経験からのフィードバック |

---

## 1. この文書の目的

### 1.1 何を改善するのか

falco-plugin-dev-kit は、Falco プラグインのコードを自動生成するツールキットです。
テンプレート（`.tmpl` ファイル）からプラグインのソースコードを生成し、スキル（`SKILL.md`）の
手順に従って開発を進めます。

**問題**: 現在のテンプレートから生成されるコードには、以下の問題があります。

1. **パーサーが未接続** — 生成直後のコードは、ログを読み取れるが解析できない（TODO のまま）
2. **HTTP 専用** — HTTP アクセスログ以外（AI ログ、IoT ログ等）に対応できない
3. **macOS で動かない** — Makefile が Linux 固定でビルドに失敗する
4. **テストが足りない** — パーサーの単体テストしかなく、プラグイン全体の動作テストがない
5. **CI/CD が最小限** — E2E テストやマルチプラットフォームリリースに対応していない

### 1.2 どのように改善するのか

falco-plugin-openclaw（このツールキットで生成した実績あるプラグイン）の開発で得た知見を、
テンプレートとスキルにフィードバックします。

### 1.3 改善後どうなるのか

テンプレートから生成したコードが、**最小限の手修正で動作する**状態になります。
具体的には:

- 生成直後に `go vet ./...` がパスする
- `go test ./...` で全テストが通る（パーサー＋パイプライン＋E2E パターン）
- macOS でも Linux でも `make build` が成功する
- HTTP 以外のログソースにも対応できる

### 1.4 改善対象

| 対象 | 説明 | ファイル数 |
|------|------|-----------|
| テンプレート | `.claude/templates/plugin/*.tmpl` | 既存 15 + 新規 8 |
| スキル定義 | `.claude/skills/*/SKILL.md` | 既存 5 |
| エージェント | `.claude/agents/plugin-dev-workflow.md` | 既存 1 |
| 問題パターン集 | `PROBLEM_PATTERNS.md` | 既存 1 |

---

## 2. 改善項目一覧

全改善項目を「何を」「どう変えて」「どうなるか」の形で整理します。

### 改善項目マップ

```
A. テンプレートの改善（生成されるコードを直接改善する）
   A1. plugin.go.tmpl  — パーサー接続、ドメイン非依存化、パス展開
   A2. parser.go.tmpl   — JSON パーサー実装、フォーマット自動検出、LogEntry 非依存化
   A3. regex_simple.go.tmpl — 入力サイズ超過時の挙動修正、URL デコード重複排除
   A4. Makefile.tmpl    — OS 自動検出、E2E ターゲット追加
   A5. ci.yml.tmpl      — 3 ワークフロー分離
   A6. falco.yaml.tmpl  — 3 環境対応
   A7. テスト系テンプレート（新規 3 ファイル）
   A8. ドキュメント系テンプレート（新規 2 ファイル + README 更新）
   A9. config.go.tmpl   — ドメイン非依存化、フィールド拡張

B. スキル・エージェントの改善（生成手順を改善する）
   B1〜B5. 各スキルの更新
   B6. ワークフローエージェントの更新

C. 新規追加
   C1. dev-kit-feedback スキル（新規）
   C2. PROBLEM_PATTERNS.md への知見追加

E. 非機能要件（横断的な制約条件 — 個別タスクではなく全改善項目に適用）
   E1〜E3. 後方互換性、ドメイン非依存性、ドキュメント整合性
   E4〜E8. 互換性、性能、セキュリティ、テンプレート変数仕様、受け入れテスト
```

---

## 3. テンプレートの改善（カテゴリ A）

### A1. plugin.go.tmpl の改善

---

#### A1-1: parseLine() と parser パッケージの接続【P0 Critical】

**現状の問題**:
テンプレートから生成されたコードでは、`parseLine()` が TODO のままです。
`parser` パッケージの import もありません。
つまり、生成直後のプラグインは**ログを読み取れるが、解析できません**。
開発者は毎回手作業で parser を接続する必要があります。

**現在のコード** (`plugin.go.tmpl` 277-302行目):

```go
func parseLine(line, path string) *PluginEvent {
    // TODO: Replace with actual parser call
    // entry, err := parser.Parse(line)
    // if err != nil { return nil }

    event := &PluginEvent{
        // Map LogEntry fields to PluginEvent fields
        // RemoteAddr:  entry.RemoteAddr,    // ← 全てコメントアウト
        // ...
        LogPath:   path,
        Raw:       line,
        Timestamp: time.Now(),
        Headers:   make(map[string]string),
    }
    return event
}
```

また、`readLoop()` と `readNewLines()` に parser の引数がなく、
`MyPlugin` 構造体にも parser フィールドがありません。

**改善後のコード**:

```go
import (
    // ... 既存 import ...
    "github.com/${AUTHOR}/${PLUGIN_NAME}/pkg/parser"  // ← 追加
)

type MyPlugin struct {
    plugins.BasePlugin
    config PluginConfig
    parser *parser.Parser    // ← 追加
}

func (p *MyPlugin) Init(config string) error {
    // ... 既存の設定パース処理 ...

    // パーサー初期化（追加）
    p.parser = parser.New(parser.Config{
        LogFormat:        "${LOG_FORMAT}",
        SecurityPatterns: true,
    })
    return nil
}

func (p *MyPlugin) Open(params string) (source.Instance, error) {
    // ... 既存のファイル監視セットアップ ...
    go instance.readLoop(p.parser)   // ← parser を渡す
    return instance, nil
}

func (inst *MyInstance) readLoop(p *parser.Parser) {
    // ... 既存ループ、readNewLines 呼び出しに parser を渡す ...
    inst.readNewLines(event.Name, p)
}

func (inst *MyInstance) readNewLines(path string, p *parser.Parser) {
    // ... 既存のファイル読み取り ...
    event := parseLine(line, path, p)  // ← parser を渡す
    // ...
}

func parseLine(line, path string, p *parser.Parser) *PluginEvent {
    entry, err := p.Parse(line)
    if err != nil {
        debugLog("Parse error: %v (line: %.100s)", err, line)
        return nil
    }

    event := &PluginEvent{
        // 共通フィールド（全プラグイン共通）
        LogPath:   path,
        Raw:       entry.Raw,
        Timestamp: entry.Timestamp,
        Headers:   entry.Headers,              // P004: parser 側で初期化済み

        // HTTP ドメイン固有フィールド（Step 1 では直接マッピング）
        RemoteAddr:  entry.RemoteAddr,
        RemoteUser:  entry.RemoteUser,
        TimeLocal:   entry.TimeLocal.Format("${TIME_FORMAT}"),
        Method:      entry.Method,
        Path:        entry.Path,
        QueryString: entry.QueryString,
        Protocol:    entry.HTTPVersion,
        Status:      uint64(entry.Status),
        BytesSent:   uint64(entry.BodyBytes),
        Referer:     entry.Referer,
        UserAgent:   entry.UserAgent,
    }
    return event
}
```

**注: Step 4（A1-2）での変更予定**:
Step 4 のドメイン非依存化では、上記の HTTP 固有フィールドマッピングが
WF-Phase 0 で収集したフィールド定義に基づいて動的生成されるように変更します。
例えば AI ログプラグインの場合、`SessionID: entry.Fields["session_id"]` のような
マッピングが自動生成されます。Step 1 の時点では HTTP 直接マッピングで実装します。

**なぜ重要か**:
この改善がないと、テンプレートから生成したコードは**何も解析できないプラグイン**です。
開発者は毎回、parser との接続コードを一から書く必要があり、
C-001（LogEntry と PluginEvent の構造体分離）の型変換ミスが頻発します。

---

#### A1-2: PluginEvent のドメイン非依存化【P1 High】

**現状の問題**:
`PluginEvent` が HTTP アクセスログ固有のフィールドをハードコーディングしています。
AI エージェントログや IoT センサーログなど、HTTP 以外のプラグインを作る場合、
構造体もフィールド定義も Extract() もすべて書き直す必要があります。

**現在のコード** (`plugin.go.tmpl` 58-74行目):

```go
type PluginEvent struct {
    RemoteAddr  string            // ← HTTP 固有
    RemoteUser  string            // ← HTTP 固有
    TimeLocal   string            // ← HTTP 固有
    Method      string            // ← HTTP 固有
    Path        string            // ← HTTP 固有
    QueryString string            // ← HTTP 固有
    Protocol    string            // ← HTTP 固有
    Status      uint64            // ← HTTP 固有
    BytesSent   uint64            // ← HTTP 固有
    Referer     string            // ← HTTP 固有
    UserAgent   string            // ← HTTP 固有
    LogPath     string            // 共通
    Raw         string            // 共通
    Timestamp   time.Time         // 共通
    Headers     map[string]string // 共通
}
```

**改善の方向性**:

スキャフォールディング（WF-Phase 0）でユーザーからフィールド定義を収集し、
それに基づいて `PluginEvent`、`Fields()`、`Extract()` を動的に生成します。

テンプレートを2層構造にします:

1. **共通部分**（全プラグインで必ず含む）:
   - `Timestamp`, `LogPath`, `Raw`, `Headers`
   - `NextBatch()`, `Close()` のボイラープレート

2. **ドメイン固有部分**（WF-Phase 0 で収集したフィールドから生成）:
   - 例: HTTP なら `Method`, `Path`, `Status` 等
   - 例: AI ログなら `SessionID`, `Tool`, `Args`, `ThreatLevel` 等

**Fields() と Extract() の非依存化方針**:

PluginEvent の非依存化に伴い、`Fields()` (16 フィールド定義) と `Extract()` (switch/case 分岐)
も同時に動的生成する必要があります。方針は以下の通りです:

- WF-Phase 0 で収集したフィールド定義（フィールド名、型、説明）をテンプレート変数として受け取る
- テンプレート展開時に、PluginEvent 構造体、Fields() のフィールド定義、Extract() の switch/case を
  フィールド定義に基づいて自動生成する
- 共通フィールド（LogPath, Raw, Timestamp, Headers）は固定、ドメイン固有フィールドのみ動的生成

**テンプレート展開機構**:

現在のテンプレートシステムは `${VARIABLE}` 形式の単純な文字列置換です。
ドメイン非依存化では複数行のコードブロック（構造体フィールド、Fields() 定義、Extract() case 文）を
動的生成する必要があるため、以下のプレースホルダーを導入します:

| プレースホルダー | 用途 | 展開先テンプレート | 生成例（HTTP） |
|----------------|------|-----------------|--------------|
| `${DOMAIN_FIELDS_STRUCT}` | 構造体のドメイン固有フィールド | `plugin.go.tmpl`（PluginEvent）、`parser.go.tmpl`（LogEntry） | `RemoteAddr string` 等 |
| `${DOMAIN_FIELDS_DEFS}` | Fields() のフィールド定義配列 | `plugin.go.tmpl` | `{Type: "string", Name: "http.remote_addr", ...}` 等 |
| `${DOMAIN_FIELDS_EXTRACT}` | Extract() の switch/case 文 | `plugin.go.tmpl` | `case "http.remote_addr": req.SetValue(event.RemoteAddr)` 等 |
| `${DOMAIN_FIELDS_MAPPING}` | parseLine() のフィールドマッピング | `plugin.go.tmpl` | `RemoteAddr: entry.RemoteAddr,` 等 |
| `${DOMAIN_FIELDS_PARSE_JSON}` | parseJSON() のドメイン固有フィールド設定 | `parser.go.tmpl` | `if v, ok := raw["remote_addr"].(string); ok { entry.RemoteAddr = v }` 等 |

scaffold スキルの WF-Phase 0 でユーザーからフィールド定義（名前、型、Falco フィールド名、JSON キー名）を
収集し、WF-Phase 1 のテンプレート展開時に上記プレースホルダーに対応するコードブロックを生成して挿入します。

**設計の選択肢**:

| 方式 | 型安全性 | 汎用性 | IDE サポート | 採用 |
|------|---------|--------|-------------|------|
| 具体的フィールド生成（テンプレート展開で型付きフィールドを生成） | 高い | 高い | 充実 | 推奨 |
| 汎用マップ (`Fields map[string]interface{}`) | 低い | 高い | 限定的 | 代替案 |

推奨方式は、テンプレート展開時に具体的な型付きフィールドを生成する方式です。
openclaw の実装もこの方式を採用しており、型安全性と IDE サポートの両方を確保しています。

**なぜ重要か**:
現状のテンプレートは HTTP 専用であり、HTTP 以外のプラグインを作るたびに
PluginEvent の全面書き直しが必要です。openclaw 開発では 13 フィールドすべてを
一から定義し直しました。

---

#### A1-3: `~/` パス展開ロジックの追加【P2 Medium】

**現状の問題**:
`Open()` 内に `~/` のホームディレクトリ展開がありません。
設定ファイルで `"log_paths": ["~/.myapp/logs/app.log"]` と書いても、
ファイルが見つからずプラグインが正常に動作しません。

**現在のコード**: パス展開なし。`logPath` をそのまま使用。

**改善後のコード** (Open() 内に追加):

```go
for _, logPath := range p.config.LogPaths {
    // パストラバーサル防止（E6 セキュリティ要件）
    if strings.Contains(logPath, "..") {
        return nil, fmt.Errorf("path traversal not allowed: %s", logPath)
    }

    // ~/パス展開（追加）
    if strings.HasPrefix(logPath, "~/") {
        if home, err := os.UserHomeDir(); err == nil {
            logPath = filepath.Join(home, logPath[2:])
        }
    }
    // ... 既存のファイルオープン処理 ...
}
```

---

#### A1-4: Extract() の冗長な nil チェック削除【P3 Low】

**現状の問題**:
`Extract()` で `evt.EventData()` を呼んで nil チェック後、別途 `evt.Reader()` で
GOB デコードしています。`EventData()` の呼び出しは不要です。

**現在のコード** (`plugin.go.tmpl` 332-343行目):

```go
func (p *MyPlugin) Extract(req sdk.ExtractRequest, evt sdk.EventReader) error {
    var event PluginEvent
    data := evt.EventData()    // ← 不要な呼び出し
    if data == nil {           // ← 不要なチェック
        return fmt.Errorf("nil event data")
    }
    decoder := gob.NewDecoder(evt.Reader())
    // ...
}
```

**改善後のコード**:

```go
func (p *MyPlugin) Extract(req sdk.ExtractRequest, evt sdk.EventReader) error {
    var event PluginEvent
    decoder := gob.NewDecoder(evt.Reader())  // 直接デコード
    if err := decoder.Decode(&event); err != nil {
        return fmt.Errorf("failed to decode event: %w", err)
    }
    // ...
}
```

---

### A2. parser.go.tmpl の改善

---

#### A2-1: JSON パーサーのデフォルト実装【P1 High】

**現状の問題**:
`parseJSON()` が「未実装」エラーを返すだけです。
JSON 形式のログを扱うプラグインを作る場合、パーサーを一から書く必要があります。

**現在のコード** (`parser.go.tmpl` 198-201行目):

```go
func (p *Parser) parseJSON(line string) (*LogEntry, error) {
    // TODO: Implement JSON parsing based on specific log format
    return nil, fmt.Errorf("JSON format not yet implemented - customize for your log source")
}
```

**改善後のコード**:

```go
func (p *Parser) parseJSON(line string) (*LogEntry, error) {
    var raw map[string]interface{}
    if err := json.Unmarshal([]byte(line), &raw); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }

    entry := &LogEntry{
        Headers: make(map[string]string),
    }

    // 汎用フィールドマッピング（JSON キー → LogEntry フィールド）
    if v, ok := raw["timestamp"].(string); ok {
        entry.Timestamp = p.parseTimestamp(v)
    }
    if v, ok := raw["level"].(string); ok {
        entry.Headers["level"] = v
    }
    if v, ok := raw["message"].(string); ok {
        entry.Headers["message"] = v
    }
    // ドメイン固有フィールド（テンプレート展開で生成）
    ${DOMAIN_FIELDS_PARSE_JSON}

    return entry, nil
}

// parseTimestamp は複数のタイムスタンプフォーマットを試行する
func (p *Parser) parseTimestamp(s string) time.Time {
    formats := []string{
        time.RFC3339,
        time.RFC3339Nano,
        "2006-01-02T15:04:05",
        "2006-01-02 15:04:05",
        p.timeLayout,
    }
    for _, format := range formats {
        if t, err := time.Parse(format, s); err == nil {
            return t
        }
    }
    return time.Now()
}
```

**なぜ重要か**:
openclaw は JSONL 形式のログを扱います。テンプレートの parseJSON() が
未実装だったため、JSON パーサーを完全に一から書く必要がありました。
多くのモダンなアプリケーションは JSON 形式でログを出力するため、
デフォルト実装は必須です。

---

#### A2-2: フォーマット自動検出モードの追加【P2 Medium（Step 4 に配置）】

**現状の問題**:
設定で指定したフォーマットでしかパースできません。
1つのアプリケーションが JSON とプレーンテキストの両方を出力する場合
（openclaw がまさにこのケース）、自動検出が必要です。

**現在のコード** (`parser.go.tmpl` 96-105行目):

```go
switch cfg.LogFormat {
case "common":
    p.parseFunc = p.parseCommon
case "json":
    p.parseFunc = p.parseJSON
case "custom":
    p.parseFunc = p.parseCustom
default: // "combined"
    p.parseFunc = p.parseCombined
}
```

**改善後のコード**:

```go
switch cfg.LogFormat {
case "auto":               // ← 追加
    p.parseFunc = p.parseAuto
case "common":
    p.parseFunc = p.parseCommon
case "json":
    p.parseFunc = p.parseJSON
case "custom":
    p.parseFunc = p.parseCustom
default: // "combined"
    p.parseFunc = p.parseCombined
}

// parseAuto は先頭文字で JSON/テキストを自動判定する
func (p *Parser) parseAuto(line string) (*LogEntry, error) {
    trimmed := strings.TrimSpace(line)
    if len(trimmed) > 0 && trimmed[0] == '{' {
        return p.parseJSON(trimmed)
    }
    return p.parseCombined(line)
}
```

また、`config.go.tmpl` の `Config.LogFormat` フィールドの説明を更新:

```go
type Config struct {
    LogFormat string // "auto", "combined", "common", "json", "custom"
    // ...
}
```

---

#### A2-3: LogEntry 構造体のドメイン非依存化【P1 High】

**現状の問題**:
A1-2 で PluginEvent のドメイン非依存化を提案していますが、parser.go.tmpl の
`LogEntry` 構造体（41-59行目）も同様に HTTP 固有フィールドをハードコーディングしています。
PluginEvent だけ非依存化しても、LogEntry が HTTP 固有のままでは意味がありません。

**現在のコード** (`parser.go.tmpl` 41-59行目):

```go
type LogEntry struct {
    RemoteAddr     string            // ← HTTP 固有
    RemoteUser     string            // ← HTTP 固有
    TimeLocal      time.Time         // ← HTTP 固有
    Method         string            // ← HTTP 固有
    Path           string            // ← HTTP 固有
    QueryString    string            // ← HTTP 固有
    HTTPVersion    string            // ← HTTP 固有
    Status         int               // ← HTTP 固有
    BodyBytes      int               // ← HTTP 固有
    Referer        string            // ← HTTP 固有
    UserAgent      string            // ← HTTP 固有
    Request        string
    Timestamp      time.Time         // 共通
    SecurityThreat SecurityThreatType // 共通
    Headers        map[string]string  // 共通
    Raw            string            // 共通
}
```

**改善の方向性**:

A1-2 と同様に、LogEntry を 2 層構造にします:

1. **共通フィールド**（全プラグイン共通）:
   - `Timestamp`, `Raw`, `Headers`, `SecurityThreat`

2. **ドメイン固有フィールド**（WF-Phase 0 で収集したフィールドから生成）:
   - A1-2 と同様にテンプレート展開で型付きフィールドを生成（推奨方式）
   - テンプレート変数でフィールド名と型を定義

**改善後のコード** (概念):

```go
type LogEntry struct {
    // 共通フィールド
    Timestamp      time.Time
    Raw            string
    Headers        map[string]string
    SecurityThreat SecurityThreatType

    // ドメイン固有フィールド（テンプレート展開で型付きフィールドを生成）
    ${DOMAIN_FIELDS_STRUCT}
}
```

**注**: A1-2 の PluginEvent と同様に、`${DOMAIN_FIELDS_STRUCT}` プレースホルダーを使用して
型付きフィールドを生成します。汎用マップ方式（`Fields map[string]interface{}`）は
A1-2 の設計選択肢テーブル（L283-289）で代替案として位置づけられており、
型安全性と IDE サポートの観点から推奨方式を採用します。

**なぜ重要か**:
PluginEvent と LogEntry は `parseLine()` でマッピングされます（A1-1）。
片方だけドメイン非依存化しても、もう片方が HTTP 固有のままではマッピングが破綻します。
A1-2 と A2-3 は必ずセットで実装する必要があります。

**依存関係**: A1-2（PluginEvent 非依存化）とセットで実装

**注: TimeLocal と Timestamp の統合**:
現在の LogEntry には `TimeLocal`（HTTP ログの time_local）と `Timestamp`（汎用タイムスタンプ）の
2 つの時刻フィールドが存在します。ドメイン非依存化の際、`TimeLocal` は HTTP 固有フィールドのため
`Timestamp` に統合し、単一のタイムスタンプフィールドとします。

**注: parser_test.go.tmpl の更新**:
既存の parser_test.go.tmpl (248 行、33 テストケース) には HTTP 固有のテスト
（TestParseCombined, TestParseCommon 等）が含まれています。
A2-3 の実装時に、ドメイン非依存化に対応したテストケースに更新する必要があります。

**注: フォーマット固有パーサー関数のテンプレート展開**:
LogEntry のドメイン非依存化に伴い、フォーマット固有のパーサー関数もテンプレート展開が必要です:
- **parseJSON()**: JSON キー → LogEntry フィールドのマッピングにドメイン固有の展開が必要。`${DOMAIN_FIELDS_PARSE_JSON}` プレースホルダーで対応（A2-1 参照）
- **parseCombined() / parseCommon()**: regex マッチグループ → LogEntry フィールドの代入が HTTP 固有。ドメイン非依存化後は、選択されたログフォーマットに応じて scaffold スキルがパーサー関数全体を生成する方式とする
- **parseCustom()**: ユーザー定義のため変更不要（TODO のまま維持）

---

### A3. regex_simple.go.tmpl の改善

---

#### A3-1: 入力サイズ超過時の挙動修正【P1 High】

**現状の問題**:
入力が 10KB を超えると検出をスキップします（`return "", false`）。
これは**セキュリティ上の問題**です。攻撃者が意図的に大きなペイロードを送ることで、
検出を完全に回避できます。

**現在のコード** (`regex_simple.go.tmpl` 27-29行目):

```go
func (d *SimpleSecurityDetector) DetectSecurityThreat(input string) (string, bool) {
    if len(input) > d.maxInputLength {
        return "", false   // ← 検出をスキップ（攻撃者が悪用可能）
    }
    // ...
}
```

**改善後のコード**:

```go
func (d *SimpleSecurityDetector) DetectSecurityThreat(input string) (string, bool) {
    // 10KB超の入力は切り詰めて検出を続行（スキップしない）
    if len(input) > d.maxInputLength {
        input = input[:d.maxInputLength]
    }
    // ...
}
```

**なぜ重要か**:
openclaw 開発中に発見された問題です。LLM のレスポンスは容易に 10KB を超えるため、
スキップ方式では多くの脅威を見逃します。切り詰め方式なら、先頭 10KB に含まれる
脅威パターンは検出できます。

**補足**: `Extract()` は切り詰めずに全文を Falco に返却します（P020、本要件 C2 で追加予定）。
これは意図的な設計です — 検出は 10KB 以内で行い、Falco ルール側では全文を
使ったマッチングが可能です。

---

#### A3-2: URL デコードの重複排除【P2 Medium】

**現状の問題**:
URL デコードが 2 箇所で実行されています:
- `parser.go.tmpl` の `detectSecurityPatterns()` 内（247-255行目）: デコード済み文字列を `DetectSecurityThreat()` に渡す
- `regex_simple.go.tmpl` の `DetectSecurityThreat()` 内（34-42行目）: 受け取った文字列を再度デコード

`detectSecurityPatterns()` でデコード済みの文字列を渡しているにも関わらず、
`DetectSecurityThreat()` で再度 3 段階デコードが実行されるため、最大 6 段階のデコードが発生します。

**改善方針**:
URL デコードは `detectSecurityPatterns()` 内で一度だけ実行し、
デコード済み文字列を `DetectSecurityThreat()` に渡す。
`regex_simple.go.tmpl` の `DetectSecurityThreat()` 内（34-42行目）の URL デコード処理を削除する。

---

### A4. Makefile.tmpl の改善

---

#### A4-1: OS/Arch 自動検出【P0 Critical】

**現状の問題**:
Makefile が Linux amd64 固定です。macOS で `make build` すると、
存在しないクロスコンパイラを使おうとしてビルドに失敗します。

**現在のコード** (`Makefile.tmpl` 1-5行目):

```makefile
PLUGIN_NAME := ${PLUGIN_NAME}
BINARY := lib$(PLUGIN_NAME)-plugin-linux-amd64.so   # ← Linux 固定
SRC_DIR := ./cmd/plugin-sdk
GO_BUILD_FLAGS := -buildmode=c-shared
GO_ENV := CGO_ENABLED=1 GOOS=linux GOARCH=amd64     # ← Linux 固定
```

**改善後のコード**:

```makefile
PLUGIN_NAME := ${PLUGIN_NAME}
SRC_DIR := ./cmd/plugin-sdk
GO_BUILD_FLAGS := -buildmode=c-shared
GO_RELEASE_FLAGS := -buildmode=c-shared -trimpath -ldflags="-s -w"

# OS/Arch 自動検出
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Darwin)
  ifeq ($(UNAME_M),arm64)
    BINARY := lib$(PLUGIN_NAME)-plugin-darwin-arm64.dylib
    GO_ENV := CGO_ENABLED=1 GOOS=darwin GOARCH=arm64
  else
    BINARY := lib$(PLUGIN_NAME)-plugin-darwin-amd64.dylib
    GO_ENV := CGO_ENABLED=1 GOOS=darwin GOARCH=amd64
  endif
else
  BINARY := lib$(PLUGIN_NAME)-plugin-linux-amd64.so
  GO_ENV := CGO_ENABLED=1 GOOS=linux GOARCH=amd64
endif
```

**なぜ重要か**:
macOS で開発する開発者は、`make build` を実行するたびにエラーになります。
openclaw の Makefile はこの自動検出を最初の改善として実装しました。

---

#### A4-2: build-release ターゲットの追加【P1 High】

**現在のコード**: `build-release` ターゲットなし。

**改善後のコード** (Makefile に追加):

```makefile
# リリース最適化ビルド（デバッグ情報除去、バイナリサイズ削減）
build-release:
	$(GO_ENV) go build $(GO_RELEASE_FLAGS) -o $(BINARY) $(SRC_DIR)/
```

`-trimpath` はソースパスの除去、`-ldflags="-s -w"` はシンボル情報の除去です。
リリースバイナリのサイズが 30-50% 削減されます。

---

#### A4-3: E2E テスト用 Makefile ターゲットの追加【P1 High】

**現在のコード**: `test` と `test-coverage` のみ。

**改善後のコード** (追加するターゲット):

```makefile
.PHONY: e2e-pattern e2e-pipeline e2e vet

# Level 1: パターンカバレッジテスト（Falco 不要）
e2e-pattern:
	go test ./test/e2e/ -v -race -run TestPattern -count=1

# Level 2: プラグインパイプラインテスト（Falco 不要、CGO_ENABLED=1 必要）
e2e-pipeline:
	go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -count=1 -timeout 120s

# Level 1 + 2（CI 高速パス）
e2e: e2e-pattern e2e-pipeline

# 静的解析（macOS 開発用）
vet:
	go vet ./...
```

---

### A5. CI/CD ワークフローの改善

---

#### A5-1: 3 ワークフローファイルへの分離【P1 High】

**現状の問題**:
単一の `ci.yml` に test + build + release の 3 ジョブが詰め込まれています。
E2E テスト、マルチプラットフォームリリースの仕組みがありません。

**現在の構成**:

```
.github/workflows/
└── ci.yml          ← test + build + release の 3 ジョブ（84行）
```

**注**: 現在の `ci.yml.tmpl` には既に release ジョブ（59-84行目）が含まれています。
このジョブを `release.yml.tmpl` に移行し、`ci.yml.tmpl` からは削除します。

**改善後の構成**:

```
.github/workflows/
├── ci.yml          ← テスト + lint + ルール検証（push/PR 時に自動実行）
├── e2e-test.yml    ← E2E テスト + Allure レポート（新規）
└── release.yml     ← マルチプラットフォームリリース（新規）
```

**ci.yml の改善点**:

| 項目 | 現在 | 改善後 |
|------|------|--------|
| ランナー | `ubuntu-latest` | `ubuntu-24.04`（バージョンピン留め） |
| テストフラグ | なし | `-race` 追加（データ競合検出） |
| golangci-lint | `@latest` | バージョンピン留め |
| ルール検証 | なし | YAML バリデーション追加 |

**e2e-test.yml（新規テンプレート）の構成** (R5-004 対応):

```yaml
name: E2E Tests
on:
  push:
    branches: [main]
    paths: ['cmd/**', 'pkg/**', 'test/**']
  pull_request:
    branches: [main]
  workflow_dispatch:

jobs:
  go-tests:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Level 1 - Pattern Coverage Tests
        run: make e2e-pattern
      - name: Level 2 - Pipeline Tests
        run: make e2e-pipeline
```

テンプレート変数: `${PLUGIN_NAME}`（テスト結果レポートのタイトルに使用）

**release.yml（新規テンプレート）の構成** (R5-004 対応):

```yaml
name: Release
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Release version (e.g., v0.1.0)'
        required: true

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: ubuntu-24.04
            goos: linux
            goarch: amd64
            ext: .so
          - os: macos-14
            goos: darwin
            goarch: arm64
            ext: .dylib
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build
        run: make build-release
      - name: Upload artifact
        uses: actions/upload-artifact@v4

  release:
    needs: build
    runs-on: ubuntu-24.04
    permissions:
      contents: write
    steps:
      - name: Create Release with SHA256 checksums
        uses: softprops/action-gh-release@v2
```

テンプレート変数: `${PLUGIN_NAME}`（バイナリ名、リリースタイトル）、`${VERSION}`（リリースバージョン）

参照: openclaw `.github/workflows/release.yml`

---

### A6. Falco 設定テンプレートの改善

---

#### A6-1: 3 環境 Falco 設定テンプレート【P1 High】

**現状の問題**:
`falco.yaml` が 1 ファイルしかなく、Linux 本番環境のパス固定です。
macOS でのローカルテスト、Docker コンテナでの実行に対応していません。

**現在の構成**: `falco.yaml.tmpl` 1 ファイルのみ

**改善後の構成**:

| ファイル | 環境 | バイナリパス | 特記事項 |
|---------|------|------------|---------|
| `falco.yaml.tmpl` | Linux 本番 | `/usr/share/falco/plugins/lib${PLUGIN_NAME}-plugin.so` | 変更なし |
| `falco-local.yaml.tmpl` (新規) | macOS ローカル | `./lib${PLUGIN_NAME}-plugin-darwin-arm64.dylib` | `outputs:` セクション除外（P017）、`-U` フラグ注記（P018） |
| `falco-docker.yaml.tmpl` (新規) | Docker | `/plugins/lib${PLUGIN_NAME}-plugin.so` | マウントパス、`json_output: true` |

**全設定ファイル共通の必須設定**:
- `load_plugins: [${PLUGIN_NAME}]` — P008: これがないとプラグインが無視される
- `rate: 0`, `max_burst: 0` — P009: デフォルトのレート制限がアラートを抑制する
- `rules_files` に個別パス — P007: ディレクトリ指定だと登録されない場合がある

---

### A7. テスト系テンプレートの新規作成

---

#### A7-1: Level 2 パイプラインテストテンプレート【P0 Critical】

**現状の問題**:
テンプレートには `parser_test.go.tmpl`（パーサー単体テスト）しかありません。
プラグイン全体のパイプライン（Init → Open → ログ書き込み → NextBatch → Extract）を
テストする仕組みがありません。

パーサーが正しく動いても、プラグインが正しくイベントを配信するとは限りません。
GOB エンコード/デコードのミス、チャネル処理のバグ、fsnotify の問題は
パーサー単体テストでは検出できません。

**新規作成ファイル**: `.claude/templates/plugin/plugin_test.go.tmpl`

**テンプレートに含めるテストケース**:

| カテゴリ | TC ID | テスト内容 | なぜ必要か |
|---------|-------|-----------|-----------|
| ライフサイクル | TC-1-01 | デフォルト設定での Init | 設定なしで起動できることの保証 |
| ライフサイクル | TC-1-02 | カスタム設定での Init | JSON 設定の正しいパース |
| ライフサイクル | TC-1-03 | バッファサイズ境界値 | 0, 1, 100001 等の異常値の処理 |
| ライフサイクル | TC-1-06 | Open() ファイル自動作成 | ログファイルが存在しない場合の動作 |
| ライフサイクル | TC-1-07 | Open() SeekEnd (P014) | 既存ログをスキップすることの保証 |
| ライフサイクル | TC-1-08 | Close() リソース解放 | ファイルハンドルやチャネルのリーク防止 |
| 取り込み | TC-2-01 | 基本ログ取り込み | ログ書き込み → イベント受信の基本動作 |
| 取り込み | TC-2-04 | 複数ファイル監視 | 複数ログファイルの同時監視 |
| 取り込み | TC-2-05 | GOB ラウンドトリップ | エンコード→デコードでデータが壊れないこと |
| 取り込み | TC-2-06 | Headers 非 nil (P004) | GOB エンコード時の nil map panic 防止 |
| 性能 | TC-5-01 | スループット | 100 events/sec 以上の処理能力 |
| 性能 | TC-5-02 | バッファオーバーフロー | チャネル溢れ時にハングしないこと |
| エラー耐性 | TC-6-01 | 不正 JSON 設定 | 不正な設定文字列での Init |
| エラー耐性 | TC-6-04 | ファイル削除時 | 監視中のファイルが消された場合 |

**テンプレートに含めるヘルパー関数**:

```go
// initPlugin — プラグイン初期化のボイラープレートを隠蔽
func initPlugin(t *testing.T, logPaths []string) *MyPlugin { ... }

// openAndCleanup — Open() + t.Cleanup(Close) を登録
func openAndCleanup(t *testing.T, plugin *MyPlugin) *MyInstance { ... }

// writeToLog — テスト用ログ行の書き込み + fsnotify 伝播待機
func writeToLog(t *testing.T, path, line string) { ... }

// waitForEvent — チャネルからイベントをタイムアウト付きで待機
func waitForEvent(t *testing.T, ch <-chan *PluginEvent, timeout time.Duration) *PluginEvent { ... }

// gobEncode / gobDecode — GOB エンコード/デコードヘルパー
func gobEncode(t *testing.T, event *PluginEvent) []byte { ... }
func gobDecode(t *testing.T, data []byte) *PluginEvent { ... }
```

参照: openclaw `cmd/plugin-sdk/plugin_test.go` (37 テスト関数、6 ヘルパー関数)

---

#### A7-2: Level 1 E2E パターンテストテンプレート【P1 High】

**新規作成ファイル**: `.claude/templates/plugin/e2e_pattern_test.go.tmpl`

E2E パターン JSON ファイルを動的に読み込み、各パターンをパーサーに通して
正しく検出されることを検証するテストフレームワークです。

**テストケース**:

| TC ID | テスト内容 |
|-------|-----------|
| TC-3-01 | 全攻撃カテゴリの True Positive（検出されるべきものが検出される） |
| TC-3-02 | True Negative テスト（正常なリクエストが誤検出されない） |
| TC-3-04 | 10KB 入力サイズ境界テスト |
| TC-3-05 | 大文字小文字非依存テスト |

参照: openclaw `test/e2e/e2e_pattern_test.go` (9 テストケース)

---

#### A7-3: E2E パターン JSON テンプレートの拡張【P1 High】

**現在のテンプレート** (`e2e_pattern.json.tmpl`):
5 カテゴリ x 4 パターン = 20 パターン（攻撃パターンのみ）

**改善後のテンプレート** (追加するパターン):

| ファイル | 実装方式 | 内容 | なぜ必要か |
|---------|---------|------|-----------|
| `benign` カテゴリ (e2e_pattern.json.tmpl に追加) | 既存テンプレートに追加セクションとして統合 | 正常リクエスト 5+ パターン | 偽陽性テスト。正常なリクエストが脅威として誤検出されないことの保証 |
| `edge_cases` カテゴリ (e2e_pattern.json.tmpl に追加) | 既存テンプレートに追加セクションとして統合 | 境界値パターン | 10KB 境界（10239/10240/10241 バイト）、空文字列、空白のみ |

**注**: `benign` と `edge_cases` は独立ファイルではなく、既存の `e2e_pattern.json.tmpl` 内に
カテゴリとして追加します。テスト側（`e2e_pattern_test.go.tmpl`）がディレクトリを走査して
全パターンを動的に読み込む設計のため、将来的にプラグイン固有のパターンを追加ファイルとして
配置することも可能です。

**生成フロー**: E2E パターンは `plugin-test` スキル（B4）のフローで生成されます。
scaffold スキルの生成対象には含まれません。

**パターン JSON スキーマの拡張** (追加フィールド):

| フィールド | 型 | 説明 |
|-----------|------|------|
| `format` | string | ログ形式（json / plaintext / combined）|
| `expected_threat` | string | パーサーレベルの期待脅威タイプ |
| `note` | string | テスト作成者向けの注記 |

---

### A8. ドキュメント系テンプレートの新規作成

---

#### A8-1: CLAUDE.md テンプレート【P2 Medium】

**新規作成ファイル**: `.claude/templates/plugin/CLAUDE.md.tmpl`

生成されたプラグインプロジェクトに、Claude Code 向けのプロジェクトガイドを自動生成します。

**テンプレート内容** (概要):

```markdown
# CLAUDE.md

## Project Overview
${PLUGIN_NAME} plugin for Falco. Monitors ${LOG_SOURCE} logs.

## Build & Development Commands
make build / make test / make e2e / make lint

## Architecture
- Plugin layer: cmd/plugin-sdk/plugin.go
- Parser layer: pkg/parser/

## Critical Constraints
- P002: -buildmode=c-shared required
- P004: Headers map must be initialized with make()
- P008: load_plugins required in falco.yaml
- P010: Fields() and Extract() must be consistent
```

参照: openclaw `CLAUDE.md`

---

#### A8-2: CHANGELOG.md テンプレート【P2 Medium】

**新規作成ファイル**: `.claude/templates/plugin/CHANGELOG.md.tmpl`

Keep a Changelog 準拠のスケルトンを生成します。

---

#### A8-3: README.md.tmpl の更新【P2 Medium】

**現状の問題**:
以下の改善によって README の記述も更新が必要ですが、既存テンプレートが未対応です:

| 改善ID | README への影響 |
|--------|---------------|
| A4-1 | OS 自動検出 → ビルドコマンドの説明変更（macOS / Linux 両対応の記載） |
| A4-3 | E2E ターゲット追加 → `make e2e`, `make e2e-pattern`, `make e2e-pipeline` の説明追加 |
| A6-1 | 3 環境設定 → `falco.yaml`, `falco-local.yaml`, `falco-docker.yaml` の説明追加 |
| A7-1/A7-2 | テストレベル → 3 層テストアーキテクチャの概要追加 |

**改善方針**: 既存 `README.md.tmpl` に上記セクションを追加する。

---

### A9. config.go.tmpl の改善

---

#### A9-1: Config 構造体のドメイン非依存化【P1 High】

**現状の問題**:
`config.go.tmpl`（9行）は HTTP パーサー固有の設定のみ:

```go
type Config struct {
    LogFormat              string // "combined", "common", "json", "custom"
    CustomFormat           string // Custom regex pattern (for "custom" format)
    SecurityPatterns       bool   // Enable security threat detection
    LargeResponseThreshold int    // Threshold for large response detection (bytes)
}
```

A1-2（PluginEvent ドメイン非依存化）および A2-3（LogEntry ドメイン非依存化）に伴い、
Config 構造体もドメイン非依存な形に改善が必要です。
特に `LargeResponseThreshold` は HTTP 固有の名称です。

**改善後のコード**:

```go
type Config struct {
    LogFormat          string // "auto", "combined", "common", "json", "custom"
    CustomFormat       string // Custom regex pattern (for "custom" format)
    SecurityPatterns   bool   // Enable security threat detection
    MaxFieldLength     int    // Threshold for large field truncation (bytes)
}
```

**変更点**:
- `LogFormat`: `"auto"` を選択肢に追加（A2-2 対応）
- `LargeResponseThreshold` → `MaxFieldLength`: ドメイン非依存な名称に変更

**注: MaxFieldLength の配線（R-302 対応）**:
現在の `LargeResponseThreshold` は `config.go.tmpl` に定義されているだけで、
`parser.go.tmpl` の `New()` 関数や `regex_simple.go.tmpl` の `SimpleSecurityDetector` から
参照されていません（`regex_simple.go.tmpl` は独自の `maxInputSize` 定数を使用）。
リネームだけでなく、以下の配線を実装する必要があります:

1. `parser.go.tmpl` の `New()` 内で `cfg.MaxFieldLength` を読み取る
2. `SimpleSecurityDetector` の初期化時に `maxInputLength` フィールドとして設定する
3. デフォルト値（0 の場合は 10KB）のフォールバック処理を追加する

**注: 2 つの Config 構造体の役割分担**:

本プラグインには 2 つの設定構造体があります:

```
Falco init_config (JSON)
  ↓ json.Unmarshal
PluginConfig (plugin.go.tmpl L48-51)
  ├── LogPaths []string      → Open() でファイル監視に使用
  └── EventBufferSize int    → eventCh のバッファサイズ
        ↓ Init() 内で parser.New() に渡す
parser.Config (config.go.tmpl)
  ├── LogFormat string       → パーサーの動作モード選択
  ├── SecurityPatterns bool  → セキュリティ検出の有効/無効
  └── MaxFieldLength int     → 入力サイズ制限
```

`LogPaths` は plugin 層の責務（ファイル監視）のため PluginConfig に属し、
parser.Config には含めません（openclaw の実装と同様の設計）。

---

### 変更なしのテンプレート

以下のテンプレートは本 v2 要件では変更対象外です:

| ファイル | 理由 |
|---------|------|
| `golangci.yml.tmpl` | 現在の linter 設定（gosec, govet, errcheck 等）で十分。openclaw と同等 |
| `LICENSE.tmpl` | ライセンステンプレートは変更不要 |
| `gitignore.tmpl` | .gitignore パターンは変更不要 |
| `go.mod.tmpl` | 依存関係は各改善項目の実装時に必要に応じて更新する（独立した改善項目としない） |
| `plugin_rules.yaml.tmpl` | 現在は HTTP/Web セキュリティルール（SQLi, XSS 等）がハードコーディングされている。ドメイン非依存化は B3（plugin-rules スキル更新、T5-9）のスコープに含め、スキル側でドメインに応じたルール生成ガイドを提供する。テンプレート自体は HTTP デフォルトを維持し、各プラグイン開発時にカスタマイズする運用とする |

---

## 4. スキル定義の改善（カテゴリ B）

スキル（`SKILL.md`）は、テンプレートを使ってコードを生成する際の手順書です。
テンプレートが変わればスキルも更新が必要です。

### B1: plugin-scaffold スキル

| 変更内容 | 現在 | 改善後 |
|---------|------|--------|
| 生成ファイル一覧 | 14 ファイル | 18 ファイル（Step 4 完了時点。scaffold 担当の新規 4 テンプレート追加分。残り 2 テンプレート（CLAUDE.md.tmpl, CHANGELOG.md.tmpl）は Step 5 完了後に追加し 20 に更新。テスト担当 2 テンプレートは test スキル担当） |
| フィールド収集 | HTTP 固定フィールド | WF-Phase 0 でドメインに応じたフィールドを対話的に収集 |
| ディレクトリ構造 | `e2e/scripts/` なし | E2E スクリプトディレクトリ追加（※Level 3 テンプレートは対象外、下記注記参照） |

**注**: scaffold が生成する 14 ファイルには `e2e_pattern.json` は含まれません。
E2E パターンは `plugin-test` スキル（B4）のフローで生成されます。
テンプレート総数（既存 15 + 新規 8 = 23）とは異なります。

**新規テンプレート一覧** (8 ファイル):

| # | ファイル | 対応する改善ID |
|---|---------|--------------|
| 1 | `plugin_test.go.tmpl` | A7-1 |
| 2 | `e2e_pattern_test.go.tmpl` | A7-2 |
| 3 | `falco-local.yaml.tmpl` | A6-1 |
| 4 | `falco-docker.yaml.tmpl` | A6-1 |
| 5 | `CLAUDE.md.tmpl` | A8-1 |
| 6 | `CHANGELOG.md.tmpl` | A8-2 |
| 7 | `e2e-test.yml.tmpl` | A5-1 |
| 8 | `release.yml.tmpl` | A5-1 |

**テンプレートとスキルの対応関係**:

```
テンプレート                    → 生成ファイル                           → 担当スキル
────────────────────────────────────────────────────────────────────────────────────
plugin.go.tmpl                 → cmd/plugin-sdk/plugin.go              → scaffold
parser.go.tmpl                 → pkg/parser/parser.go                  → scaffold
config.go.tmpl                 → pkg/parser/config.go                  → scaffold
regex_simple.go.tmpl           → pkg/parser/regex_simple.go            → scaffold
parser_test.go.tmpl            → pkg/parser/parser_test.go             → scaffold
plugin_rules.yaml.tmpl         → rules/${PLUGIN_NAME}_rules.yaml       → scaffold
falco.yaml.tmpl                → falco.yaml                            → scaffold
go.mod.tmpl                    → go.mod                                → scaffold
Makefile.tmpl                  → Makefile                              → scaffold
README.md.tmpl                 → README.md                             → scaffold
ci.yml.tmpl                    → .github/workflows/ci.yml              → scaffold
gitignore.tmpl                 → .gitignore                            → scaffold
LICENSE.tmpl                   → LICENSE                               → scaffold
golangci.yml.tmpl              → .golangci.yml                         → scaffold
e2e_pattern.json.tmpl          → test/e2e/patterns/categories/*.json   → test
plugin_test.go.tmpl (新規)     → cmd/plugin-sdk/plugin_test.go         → test
e2e_pattern_test.go.tmpl (新規)→ test/e2e/e2e_pattern_test.go          → test
falco-local.yaml.tmpl (新規)   → falco-local.yaml                      → scaffold
falco-docker.yaml.tmpl (新規)  → falco-docker.yaml                     → scaffold
CLAUDE.md.tmpl (新規)          → CLAUDE.md                             → scaffold
CHANGELOG.md.tmpl (新規)       → CHANGELOG.md                          → scaffold
e2e-test.yml.tmpl (新規)       → .github/workflows/e2e-test.yml        → scaffold
release.yml.tmpl (新規)        → .github/workflows/release.yml         → scaffold
```

**Level 3 (Falco 統合テスト) について**:
Level 3 テストスクリプト（`inject_patterns.sh`, `batch_analyzer.py` 等）はプラグイン固有の
内容が多いため、テンプレート化の対象外とします。`e2e/scripts/` ディレクトリの作成のみ行い、
スクリプトの実装は各プラグイン開発時に行います。

### B2: plugin-parser スキル

| 変更内容 | 現在 | 改善後 |
|---------|------|--------|
| JSON パーサー | 「未実装」と記載 | A2-1 のデフォルト実装の手順を追加 |
| フォーマット検出 | 固定フォーマットのみ | `"auto"` モードの説明を追加 |
| 入力サイズ超過 | 「スキップ」と記載 | 「切り詰めて続行」に変更 |

### B3: plugin-rules スキル

| 変更内容 | 現在 | 改善後 |
|---------|------|--------|
| ドメイン対応 | HTTP ルール例のみ | ドメイン非依存のルール構造ガイドを追加 |
| priority ガイド | なし | CRITICAL / WARNING / NOTICE の使い分け基準を追加 |

### B4: plugin-test スキル【P0 Critical】

| 変更内容 | 現在 | 改善後 |
|---------|------|--------|
| テストレベル | ユニットテストのみ記載 | 3 層 E2E テストアーキテクチャの説明を追加 |
| Level 2 テスト | なし | パイプラインテスト生成手順を追加 |
| Level 1 テスト | なし | E2E パターンテスト生成手順を追加 |
| 偽陽性テスト | なし | benign.json の生成手順を追加 |
| 成功基準 | ユニットテスト通過のみ | Level 2 テスト通過を追加 |

**3 層 E2E テストアーキテクチャの説明**:

```
Level 1: パターンカバレッジテスト
  場所:  test/e2e/e2e_pattern_test.go
  Falco: 不要
  内容:  パターン JSON → パーサーで直接検証
  実行:  make e2e-pattern

Level 2: プラグインパイプラインテスト
  場所:  cmd/plugin-sdk/plugin_test.go
  Falco: 不要（CGO_ENABLED=1 必要）
  内容:  Init → Open → ログ書き込み → NextBatch → Extract の全パイプライン検証
  実行:  make e2e-pipeline

Level 3: Falco 統合テスト
  場所:  e2e/scripts/
  Falco: 必要
  内容:  実際に Falco を起動してルール発火を検証
  実行:  make e2e-ci (Linux) / make e2e-native (macOS)
```

### B5: plugin-build スキル

| 変更内容 | 現在 | 改善後 |
|---------|------|--------|
| macOS ビルド | 「不可」と記載 | macOS ネイティブビルド（`.dylib`）が可能であることを明記 |
| build-release | なし | `build-release` ターゲットの説明を追加 |
| macOS 制約 | なし | P017 (outputs 除外)、P018 (-U フラグ) を追加 |

### B6: plugin-dev-workflow エージェント

| 変更内容 | 対象フェーズ |
|---------|------------|
| ドメイン固有フィールドの収集を追加 | WF-Phase 0 |
| parser 統合を自動実行 | WF-Phase 2 |
| Level 2 テスト生成を追加 | WF-Phase 4 |
| macOS ネイティブビルドを追加 | WF-Phase 5 |
| 品質ゲートに Level 2 テスト通過を追加 | WF-Phase 4→5 |
| 完了報告に E2E テスト結果サマリーを追加 | WF-Phase 6 |
| 新規テンプレートファイル一覧を更新 | WF-Phase 1 |

---

## 5. 新規追加項目（カテゴリ C）

### C1: dev-kit-feedback スキル（新規作成）【P2 Medium】

**目的**: 生成済みプラグインを分析し、テンプレートとの差分から改善提案を自動生成する。

**使い方**: `/dev-kit-feedback [plugin-path]`

**処理フロー**:
1. 指定パスのプラグインコードを読み込む
2. dev-kit テンプレートとの差分を検出する
3. 差分を分類する（テンプレート改善 / スキル改善 / 新規パターン）
4. 改善提案レポートを出力する
5. PROBLEM_PATTERNS.md への追加候補を提示する

**なぜ必要か**:
今回 openclaw → dev-kit のフィードバックを手動で行いました。
今後新しいプラグインを作るたびに、同様のフィードバックを自動化できます。

---

### C2: PROBLEM_PATTERNS.md への知見追加【P1 High】

openclaw 開発で発見された問題パターンを PROBLEM_PATTERNS.md に追加します。

#### 前提: P コード体系の新設

**現状**: PROBLEM_PATTERNS.md には A-code (A303-A334) のパターンのみが記録されています。
P コード (P001-P016) はスキル定義ファイル内でインライン知見として参照されていますが、
PROBLEM_PATTERNS.md には記録されていません。

**作業内容**:
1. PROBLEM_PATTERNS.md に「P コード: プラグイン共通パターン」セクションを新設する
2. スキル定義に散在する P001-P016 の知見を集約して記録する
3. その直後に P017-P021（openclaw 新知見）を追加する

#### P001-P016: スキル定義から集約する既存知見

| ID | パターン名 | 内容 | 参照元スキル |
|----|-----------|------|------------|
| P001 | macOS バイナリリリース | macOS でビルドしたバイナリを Linux 用として配布しない | scaffold, build |
| P002 | -buildmode=c-shared 忘れ | Makefile に `-buildmode=c-shared` を必ず含める | scaffold, build |
| P003 | source 指定必須 | 全ルールに `source: ${EVENT_SOURCE}` を含める | scaffold, rules |
| P004 | GOB nil map panic | `Headers map[string]string` は必ず `make()` で初期化 | scaffold, parser |
| P005 | evt.type 不使用 | プラグインルールで `evt.type` を使用しない | rules |
| P006 | URL エンコードパターン | ルール条件に URL エンコード版パターンも含める | parser, rules |
| P007 | rules_files 個別パス | ディレクトリ指定だと登録されない場合がある | scaffold |
| P008 | load_plugins 欠落 | falco.yaml に `load_plugins` ディレクティブを必ず含める | scaffold |
| P009 | レート制限でアラート抑制 | `rate: 0`, `max_burst: 0` を設定 | scaffold |
| P010 | Extract() フィールド処理漏れ | Fields() で定義した全フィールドを Extract() で処理 | scaffold |
| P011 | YAML コメント位置 | 複数行条件内にコメントを書かない | rules |
| P012 | headers 参照小文字 | headers フィールド参照は小文字で統一 | rules |
| P013 | ビルド環境不一致 | CI と開発環境の GLIBC 互換性・Go バージョンを合わせる | build |
| P014 | ファイルシーク位置 | Open() で `file.Seek(0, io.SeekEnd)` を実行 | scaffold |
| P015 | クロスルール干渉 | 複数ルール間の条件重複に注意 | rules |
| P016 | URL エンコード多段 | 多段エンコード（%25xx 等）への対応 | rules |

#### P017-P021: openclaw 開発で発見された新知見

| ID | パターン名 | 内容 | 発見経緯 |
|----|-----------|------|---------|
| P017 | macOS Falco outputs 拒否 | macOS の Falco 0.43.0 は `outputs:` セクションを拒否する。`falco-local.yaml` から除外が必要 | macOS でのローカルテスト中に発見 |
| P018 | macOS -U フラグ必須 | macOS で Falco 実行時、stdout バッファリング無効化のため `-U` フラグが必要 | macOS でアラートが表示されない問題の調査で発見 |
| P019 | Falco 1イベント1ルール制約 | Falco 0.43.0 は 1 イベントに対して最初にマッチしたルールのみ発火する。ルール順序が重要 | E2E テストで複数ルールが発火しないことを発見 |
| P020 | 検出 truncation vs 全文返却 | `DetectThreat()` は 10KB で切り詰めるが、`Extract()` は全文を返却する。意図的な設計 | 10KB 超のログで検出とフィールド抽出の不一致を発見 |
| P021 | fsnotify タイミング | Level 2 テストで `time.Sleep` による伝播待機が必要。テストにコメントで理由を明記 | テストが不安定に失敗する問題の調査で発見 |

---

## 6. 非機能要件

### E1: 後方互換性

テンプレート変更後に、新規プラグインを生成し `go vet ./...` と `go test ./...` が
パスすることを確認する。
既存プラグイン（nginx, openclaw）のコードには影響を与えない。

### E2: テンプレートのドメイン非依存性

テンプレートが HTTP/Web に依存しないこと。HTTP 固有の要素は WF-Phase 0 の選択肢として
提供し、デフォルトとしてハードコーディングしない。

検証: HTTP アクセスログ用、AI エージェントログ用、IoT センサーログ用の
3 種類のプラグインを生成して動作確認する。

### E3: ドキュメント整合性

全スキルの SKILL.md が最新のテンプレートと整合していること。
スキル内で参照されているファイルが全て存在すること。

### E4: 互換性要件

| 項目 | 要件 |
|------|------|
| Go バージョン | 1.22 以上 |
| Falco バージョン | 0.43.0 以上 |
| plugin-sdk-go | v0.7.4 以上 |
| OS | Linux (amd64), macOS (arm64, amd64) |

### E5: 性能要件

| 項目 | 基準 |
|------|------|
| スループット | 生成されたプラグインが 100 events/sec 以上を処理できること（A7-1 TC-5-01） |
| メモリ | NextBatch() 呼び出し間でイベントバッファが無制限に増加しないこと |
| 起動時間 | Init() + Open() が 5 秒以内に完了すること |

### E6: セキュリティ要件

テンプレートが生成するコードは以下のセキュリティ基準を満たすこと:

| 項目 | 基準 | 関連する改善ID |
|------|------|--------------|
| ReDoS 防止 | 正規表現を使用せず文字列マッチングで検出する | A3（既存設計の維持） |
| nil map panic 防止 | `Headers` マップは常に `make()` で初期化する | P004 |
| 入力サイズ制限 | 10KB 超の入力は切り詰めて処理する（スキップしない） | A3-1 |
| パストラバーサル防止 | ログファイルパスに `..` が含まれないことを検証する | A1-3（T5-1 でパス展開と同時に実装） |

### E7: テンプレート変数仕様

テンプレート (`.tmpl` ファイル) で使用される変数の一覧と仕様です。
「既存」は現在のテンプレートで使用中の変数、「v2 新規」は本要件で新たに追加する変数です。

#### 既存変数（現在のテンプレートで使用中）

| 変数名 | 用途 | 例 | 使用テンプレート | バリデーション |
|--------|------|------|----------------|-------------|
| `${PLUGIN_NAME}` | プラグイン名（小文字） | `openclaw` | plugin.go, parser.go, Makefile 等ほぼ全て | `^[a-z][a-z0-9-]*$` |
| `${PLUGIN_NAME_UPPER}` | プラグイン名（大文字） | `OPENCLAW` | plugin.go, plugin_rules.yaml, e2e_pattern.json, README | 自動生成 |
| `${PLUGIN_NAME_CAMEL}` | プラグイン名（CamelCase） | `Openclaw` | plugin.go（構造体名の変換指示コメント） | 自動生成 |
| `${PLUGIN_ID}` | プラグイン ID | `999` | plugin.go | 整数 |
| `${EVENT_SOURCE}` | Falco イベントソース名 | `openclaw` | plugin.go, plugin_rules.yaml, falco.yaml | `^[a-z][a-z0-9_]*$` |
| `${VERSION}` | プラグインバージョン | `0.1.0` | plugin.go | SemVer |
| `${AUTHOR}` | 作者名（GitHub ユーザー名） | `takaosgb3` | plugin.go, go.mod, LICENSE, README | GitHub ユーザー名 |
| `${LOG_PATH_DEFAULT}` | デフォルトログファイルパス | `/var/log/app/access.log` | plugin.go, falco.yaml, README | 有効なファイルパス |
| `${LOG_FORMAT}` | デフォルトログ形式 | `combined` | plugin.go（v2 A1-1 で追加予定） | `auto\|combined\|common\|json\|custom` |
| `${SDK_VERSION}` | plugin-sdk-go バージョン | `0.8.1` | go.mod | SemVer |
| `${YEAR}` | ライセンス年 | `2026` | LICENSE | 西暦 4 桁 |
| `${LICENSE}` | ライセンス種別 | `Apache-2.0` | README.md | SPDX 識別子 |
| `${TIME_FORMAT}` | タイムスタンプフォーマット | `02/Jan/2006:15:04:05 -0700` | plugin.go（コメント内参照） | Go time.Parse フォーマット |

#### v2 で新規追加する変数

| 変数名 | 用途 | 例 | 追加理由 | バリデーション |
|--------|------|------|---------|-------------|
| `${PLUGIN_DESCRIPTION}` | プラグインの説明文 | `Monitors AI assistant logs` | T4-1: Info() の Description を動的化（PluginEvent 非依存化と同時に実装） | 自由テキスト |
| `${LOG_SOURCE}` | 監視対象のログソース説明 | `OpenClaw AI assistant` | A8-1: CLAUDE.md 生成時に使用 | 自由テキスト |

**注**:
- 既存変数は `plugin-scaffold/SKILL.md` の Phase 1 で対話的に収集されます
- v2 新規変数は WF-Phase 0 の収集項目に追加されます
- `${LOG_FORMAT}` は v2 で選択肢に `"auto"` を追加します（A2-2 対応）
- `${LICENSE}` は scaffold SKILL.md で収集されます（LICENSE.tmpl は Apache-2.0 固定）
- Go モジュールパスは `github.com/${AUTHOR}/${PLUGIN_NAME}` として構成されます（独立した変数ではなく既存変数の組み合わせ）

### E8: 受け入れテスト

v2 改善の完了を判定するための受け入れテスト:

| TC ID | テストケース | 入力 | 期待結果 |
|-------|------------|------|---------|
| AT-1 | HTTP プラグイン生成 | ログ形式=combined, フィールド=HTTP 標準 | `go vet` + `go test` + `make build` 成功 |
| AT-2 | AI プラグイン生成 | ログ形式=json, フィールド=type,tool,args,session_id | 同上 |
| AT-3 | IoT プラグイン生成 | ログ形式=custom, フィールド=device_id,sensor_type,value(string) | 同上 |
| AT-4 | macOS ビルド | AT-1 を macOS arm64 で実行 | `make build` で `.dylib` 生成成功 |
| AT-5 | E2E テスト | AT-1 で `make e2e` を実行 | Level 1 + Level 2 テスト全通過 |

**カバレッジ注記**: AT-1〜AT-3 の `go test` には以下の検証が含まれます:
- A3-1（入力サイズ切り詰め）: Level 2 テスト TC-2-01 で 10KB 超入力の処理を検証
- A1-3（パス展開）: Level 2 テスト TC-6-01 でパストラバーサル防止を検証
- A7-1/A7-2/A7-3: Level 1 + Level 2 + E2E パターンテストとして AT-5 で包括検証

---

## 7. 実装ステップ計画

**用語の区別**:
- 「実装ステップ」= 本要件定義書の改善項目を実装する順序（Step 1〜5）
- 「ワークフロー Phase」= プラグイン開発時のワークフロー段階（WF-Phase 0〜6、B6 参照）

### Step 1: 基盤改善【P0 Critical】

| 改善ID | 内容 | 完了条件 |
|--------|------|---------|
| A1-1 | parseLine() と parser の接続 | 生成コードが `go vet` パス |
| A4-1 | Makefile OS 自動検出 | macOS/Linux 両方で `make build` 成功 |
| A7-1 | Level 2 パイプラインテストテンプレート | 生成コードに pipeline テストが含まれる |
| B4 | plugin-test スキル更新 | 3 層テストの手順が記載されている |

**検証手順**:
1. テスト用プラグイン「test-plugin」をテンプレートから生成
2. `cd test-plugin && go vet ./...` → パス
3. `go test ./... -v` → 全テスト通過
4. `make build` → macOS と Linux の両方で成功

### Step 2: テスト強化・セキュリティ修正【P1 High】

| 改善ID | 内容 | 完了条件 |
|--------|------|---------|
| A3-1 | 入力サイズ超過時の挙動修正 | 10KB 超の入力で検出が続行される |
| A2-1 | JSON パーサーのデフォルト実装 | JSON ログがパースできる |
| A7-2 | Level 1 E2E パターンテストテンプレート | パターンテストが生成される |
| A7-3 | benign/edge_case パターン追加 | 偽陽性テストと境界値テストが含まれる |
| C2 | PROBLEM_PATTERNS.md 追加 | P017-P021 が記載されている |

**検証手順**:
1. Step 1 のテスト用プラグインに Step 2 の変更を適用
2. `make e2e-pattern` → Level 1 テスト全通過
3. 10KB 超の入力でセキュリティ検出が動作することを確認

### Step 3: CI/CD・ビルド改善【P1 High】

| 改善ID | 内容 | 完了条件 |
|--------|------|---------|
| A4-2 | build-release ターゲット | `make build-release` 成功 |
| A4-3 | E2E テスト Makefile ターゲット | `make e2e` 成功 |
| A5-1 | CI/CD 3 ワークフロー分離 | 3 ファイルのテンプレートが存在 |
| A6-1 | 3 環境 Falco 設定 | 3 ファイルのテンプレートが存在 |
| A9-1 | config.go.tmpl ドメイン非依存化 | Config 構造体が汎用化されている |
| B5 | plugin-build スキル更新 | macOS ビルド手順が記載されている |

**検証手順**:
1. テンプレートからプラグインを生成
2. `make build-release` → 成功
3. `make e2e` → Level 1 + Level 2 テスト全通過
4. 3 つの Falco 設定ファイルが存在すること

### Step 4: ドメイン非依存化・拡張【P1 High】

| 改善ID | 内容 | 完了条件 |
|--------|------|---------|
| A1-2 | PluginEvent のドメイン非依存化 | HTTP 以外のフィールドでプラグインが生成できる |
| A2-2 | フォーマット自動検出（P2 Medium を Step 4 に前倒し） | `"auto"` モードで JSON/テキストが判定される |
| A2-3 | LogEntry のドメイン非依存化 | LogEntry が共通 + `${DOMAIN_FIELDS_STRUCT}` による型付きフィールドの 2 層構造になっている |
| B1 | plugin-scaffold スキル更新 | カスタムフィールド収集手順が記載されている |
| B2 | plugin-parser スキル更新 | JSON パーサーと auto モードの手順が記載されている |

**検証手順**:
1. 受け入れテスト AT-1（HTTP）、AT-2（AI）、AT-3（IoT）の 3 種類を生成
2. 各プラグインで `go vet` + `go test` + `make build` が成功すること

### Step 5: ドキュメント・仕組み化【P2 Medium】

| 改善ID | 内容 | 完了条件 |
|--------|------|---------|
| A1-3 | ~/パス展開 | `~/` パスが展開される |
| A1-4 | Extract() 冗長チェック削除 | 不要なコードが削除されている |
| A3-2 | URL デコード重複排除 | ~~デコードが 1 箇所のみ~~ ※ A1-1 (Step 1) に統合済み (R5-001) |
| A8-1 | CLAUDE.md テンプレート | テンプレートが生成される |
| A8-2 | CHANGELOG.md テンプレート | テンプレートが生成される |
| A8-3 | README.md.tmpl 更新 | テスト・ビルド・設定の説明が更新されている |
| C1 | dev-kit-feedback スキル | フィードバックスキルが動作する |
| B6 | ワークフローエージェント更新 | 全変更が反映されている |
| B3 | plugin-rules スキル更新 | ドメイン非依存のガイドが記載されている |

**検証手順**:
1. テンプレートから生成したプラグインに CLAUDE.md, CHANGELOG.md が含まれること
2. README.md に 3 層テストと macOS ビルドの説明があること

### 改善項目間の依存関係

```
Step 1 (P0 Critical)
  A1-1 (parser接続)  ←──────────────────────┐
  A4-1 (OS自動検出)                          │
  A7-1 (Level 2テスト) ── 依存 ── A1-1      │
  B4   (test スキル)                         │
                                             │
Step 2 (P1 High)                             │
  A2-1 (JSONパーサー) ←─────────────┐        │
  A3-1 (入力サイズ修正)              │        │
  A7-2 (Level 1テスト)               │        │
  A7-3 (パターン追加)                │        │
  C2   (PROBLEM_PATTERNS)           │        │
                                    │        │
Step 3 (P1 High)                    │        │
  A4-2 (build-release)              │        │
  A4-3 (E2Eターゲット) ←─ A7-1,A7-2│        │
  A5-1 (CI分離)                     │        │
  A6-1 (3環境設定)                   │        │
  A9-1 (config非依存化)              │        │
  B5   (build スキル)               │        │
                                    │        │
Step 4 (P1 High)                    │        │
  A1-2 (PluginEvent非依存化) ───────┤── A1-1の再改修
  A2-2 (auto検出) ── 依存 ── A2-1 ─┘        │
  A2-3 (LogEntry非依存化) ── セット ── A1-2  │
  B1   (scaffold スキル)                     │
  B2   (parser スキル)                       │
                                             │
Step 5 (P2 Medium)                           │
  A1-3, A1-4, A8-1, A8-2, A8-3             │
  (A3-2 は A1-1 に統合済み R5-001)          │
  C1   (feedback スキル)                     │
  B6   (エージェント) ── 全Stepの完了が前提  │
  B3   (rules スキル)                        │
```

**重要な依存関係**:
- A7-1（Level 2 テスト）は A1-1（parser 接続）完了が前提
- A2-2（自動検出）は A2-1（JSON パーサー）の前提が必要
- A1-2 と A2-3 はセットで実装（PluginEvent + LogEntry の同時非依存化）
- A4-3（E2E ターゲット）は A7-1, A7-2 のテストテンプレートが必要
- B1-B6（スキル更新）は対応する A カテゴリの完了が前提

---

## 8. 用語集

| 用語 | 説明 |
|------|------|
| dev-kit | falco-plugin-dev-kit。プラグイン生成ツールキット |
| テンプレート | `.claude/templates/plugin/` 内の `.tmpl` ファイル。変数展開でコードを生成する |
| スキル | `.claude/skills/` 内の `SKILL.md`。Claude Code の `/skill-name` コマンドで実行される手順書 |
| エージェント | `.claude/agents/` 内の定義。複数スキルを Phase 順に実行する自律エージェント |
| P コード | プラグイン共通の問題パターン ID（例: P004 = nil map panic）。現在はスキル定義内にインライン記載されており、C2 で PROBLEM_PATTERNS.md に集約予定 |
| Level 1/2/3 | E2E テストのレベル。1=パターン検証, 2=パイプライン検証, 3=Falco 統合テスト |
| GOB | Go 標準の `encoding/gob` バイナリシリアライゼーション。プラグイン←→Falco 間のデータ交換に使用 |
| fsnotify | ファイル変更監視ライブラリ。ログファイルへの書き込みを検知してプラグインに通知 |
| WF-Phase 0 | プラグイン開発ワークフローのフェーズ 0。スキャフォールディング前の対話的な情報収集 |
| 実装ステップ (Step) | 本要件定義書の改善項目を実装する順序。Step 1（P0 Critical）〜 Step 5（P2 Medium） |

---

## 9. 改訂履歴

| 日付 | バージョン | 変更内容 |
|------|-----------|---------|
| 2026-03-06 | 1.0 | 初版作成 |
| 2026-03-06 | 2.0 | 全面改訂。具体的な before/after コード例を追加。改善理由を明記 |
| 2026-03-07 | 3.0 | レビュー結果(22件)に基づく修正。R-001〜R-022 の全指摘を反映。非機能要件 E4-E8 追加、A2-3/A8-3/A9-1 追加、依存関係図追加、Phase 番号体系の明確化、テンプレート変数仕様追加 |
| 2026-03-07 | 4.0 | 再レビュー結果(16件)に基づく修正。R2-001〜R2-016 を反映。P001-P016 の PROBLEM_PATTERNS.md 集約を C2 に追加、E7 テンプレート変数仕様を既存/新規に分類して完全化、A1-1 import パス修正、PluginConfig/Config の関係明示、A9-1 LogPaths 削除、B1 ファイル数修正、テンプレート→スキル対応表追加、Fields()/Extract() 非依存化方針追記 |
| 2026-03-07 | 5.0 | レビュー R-301〜R-318 の全指摘を反映。scaffold ファイル数 22→20 修正(R-301)、MaxFieldLength 配線追加(R-302)、plugin_rules.yaml.tmpl 分類明記(R-303)、パストラバーサル防止を A1-3 に追加(R-304)、テンプレート展開機構の定義追加(R-306)、config.go.tmpl 行数修正(R-309)、parser_test.go.tmpl 行数修正(R-310)、A2-2 優先度注記(R-311)、A1-1 After コードを Step 1 実装に限定(R-313) |
| 2026-03-10 | 5.1 | 再レビュー R2-001〜R2-010 の指摘を反映。openclaw テスト関数数 36→37 修正(R2-002)、E7 LOG_FORMAT バリデーションに auto 追加(R2-005)、E6 パストラバーサル記述をコードと統一(R2-006) |
| 2026-03-10 | 5.2 | 修正レビュー R3-001〜R3-018 反映。A5-1 ジョブ数矛盾修正(R3-001)、A2-3 を A1-2 と同じテンプレート展開方式に統一(R3-002)、PLUGIN_DESCRIPTION 参照先修正(R3-003)、ci.yml.tmpl 行数 85→84 修正(R3-012/R3-013)、ターゲート→ターゲット誤字修正(R3-014)、E カテゴリ注記追加(R3-015)、E7 使用テンプレート列整理(R3-009/R3-016)、E8 カバレッジ注記追加(R3-008)、A7-3 ファイル名表記修正(R3-010) |
| 2026-03-10 | 5.3 | R3-019 反映。parseJSON() ドメイン固有フィールド設定の `${DOMAIN_FIELDS_PARSE_JSON}` プレースホルダー追加。A2-3 にフォーマット固有パーサー関数のテンプレート展開注記追加 |
| 2026-03-10 | 5.4 | 第3回レビュー R4-001〜R4-006 反映。Step 4 サマリー A2-3 完了条件を型付きフィールド方式に修正(R4-002) |
| 2026-03-15 | 5.5 | 第4回レビュー R5-001〜R5-013 反映。A5-1 に e2e-test.yml/release.yml の詳細構成追加(R5-004)、AT-3 の value フィールドを float64→string に変更(R5-012) |
| 2026-03-15 | 5.6 | 実装リハーサルレビュー RH-009 反映。Step 5 テーブルの A3-2 に T1-1 統合注記追加 |

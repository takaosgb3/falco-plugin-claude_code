---
name: plugin-scaffold
description: Falcoプラグインの初期構造（ディレクトリ、設定、スケルトンコード）を対話的に生成する。新しいプラグインの作成、scaffold、プロジェクト初期化、ディレクトリ構造の生成に使用する。ユーザーが「新しいプラグインを作りたい」「プラグインを生成して」「scaffold」「初期構造」「プロジェクト作成」と言った場合にトリガーする。20テンプレートからコード生成し（scaffold担当18 + test担当2）、5つのドメイン固有プレースホルダー（STRUCT/DEFS/EXTRACT/MAPPING/PARSE_JSON）を展開する。パーサー実装の詳細にはplugin-parser、ルール作成にはplugin-rules、テストにはplugin-test、ビルドにはplugin-buildを使用すること。
argument-hint: "[plugin-name] [log-source]"
---

# プラグインスキャフォールディング

$ARGUMENTS についてFalcoプラグインの初期構造を生成します。

## 引数

- **plugin-name**: プラグイン名（英小文字、ハイフン、アンダースコアのみ）
  - 形式: `^[a-z][a-z0-9_-]*$`
  - 例: `apache`, `envoy`, `haproxy`, `my-custom-app`
- **log-source** (任意): ログソースの種類
  - `file`: ファイル監視（デフォルト）
  - `pipe`: 標準入力
  - `socket`: ソケット

## 実行手順

### 1. Phase 1: 情報収集（対話型）

ユーザーから以下の情報を対話的に収集する。

#### 1.1 プラグイン名の確認・バリデーション

```bash
# 名前の形式検証
echo "${PLUGIN_NAME}" | grep -qE '^[a-z][a-z0-9_-]*$' || echo "ERROR: プラグイン名は英小文字、数字、ハイフン、アンダースコアのみ使用可能"
```

確認項目:

- 名前が `^[a-z][a-z0-9_-]*$` に適合するか
- 既存のFalcoプラグイン名と重複しないか（Falco Registry確認推奨）
- Go命名規則に変換可能か

#### 1.2 ログソース情報の収集

ユーザーに以下を質問する:

1. **ログフォーマット**: auto / combined / common / json / custom
2. **ログファイルパス**: デフォルトパスを提案（例: `/var/log/<service>/access.log`）
3. **セキュリティ検出要件**: 検出したい攻撃の種類（SQLi, XSS, Path Traversal, Command Injection, Suspicious Agent等）
4. **プラグイン説明文**: Falco Info() の Description に使用（例: "Apache access log monitoring plugin for Falco"）
5. **ドメイン固有フィールド定義**: 抽出したいフィールドの一覧。各フィールドについて以下を収集:
   - **Go フィールド名**: 構造体フィールド名（例: `RemoteAddr`）
   - **Falco フィールド名**: `${EVENT_SOURCE}.xxx` 形式（例: `apache.remote_addr`）
   - **Go 型**: `string` または `uint64`（Falco SDK 制約: Extract() で使用可能なのはこの 2 型のみ）
   - **JSON キー名**: JSON パース時のキー名（例: `remote_addr`）。parseJSON() 生成に使用
   - **説明**: フィールドの説明文
   - **LogEntry 型**: parser 内部の型（例: `int` → Extract() では `uint64` に変換）

#### 1.3 プラグインメタデータの収集

1. **Plugin ID**: 開発用は `999`。本番用はFalco Registryで正式IDを取得
2. **バージョン**: 初期は `0.1.0`
3. **作者情報**: GitHub ユーザー名
4. **ライセンス**: デフォルト `Apache-2.0`

### 2. Phase 2: コード生成

テンプレートファイル（`.claude/templates/plugin/`）を読み込み、変数を展開してコードを生成する。

#### 2.1 CamelCase変換ルール

プラグイン名からGo構造体名への変換:

```
plugin-name → 構造体名
apache      → ApachePlugin, ApacheInstance, ApacheEvent, ApacheConfig
my-app      → MyAppPlugin, MyAppInstance, MyAppEvent, MyAppConfig
haproxy     → HaproxyPlugin, HaproxyInstance, HaproxyEvent, HaproxyConfig
```

変換ルール:

- ハイフン/アンダースコアで分割
- 各パートの先頭を大文字化
- "My" プレフィックス付き構造体: `MyPlugin` → `${CamelCase}Plugin`
- "Plugin" プレフィックス付き構造体: `PluginConfig` → `${CamelCase}Config`, `PluginEvent` → `${CamelCase}Event`

#### 2.2 ディレクトリ構造の作成

```bash
mkdir -p ${PLUGIN_NAME}/{cmd/plugin-sdk,pkg/parser,rules,test/e2e/patterns/categories,test/fixtures/sample_logs,.github/workflows,docs}
```

生成されるディレクトリ構造:

```
${PLUGIN_NAME}/
├── cmd/plugin-sdk/
│   └── plugin.go          ← メインプラグインコード
├── pkg/parser/
│   ├── parser.go          ← ログパーサー
│   ├── config.go          ← パーサー設定
│   └── regex_simple.go    ← セキュリティ検出
├── rules/
│   └── ${PLUGIN_NAME}_rules.yaml  ← Falcoルール
├── test/
│   ├── e2e/patterns/categories/   ← E2Eテストパターン
│   └── fixtures/sample_logs/      ← サンプルログ
├── .github/workflows/
│   └── ci.yml             ← CI/CDワークフロー
├── docs/
├── go.mod
├── Makefile
├── README.md
├── LICENSE
├── .gitignore
├── .golangci.yml
└── falco.yaml             ← Falco設定ファイル
```

#### 2.3 コードファイルの生成

テンプレートから以下のファイルを生成する:

1. `cmd/plugin-sdk/plugin.go` ← `.claude/templates/plugin/plugin.go.tmpl`
2. `pkg/parser/parser.go` ← `.claude/templates/plugin/parser.go.tmpl`
3. `pkg/parser/config.go` ← `.claude/templates/plugin/config.go.tmpl`
4. `pkg/parser/regex_simple.go` ← `.claude/templates/plugin/regex_simple.go.tmpl`
5. `pkg/parser/parser_test.go` ← `.claude/templates/plugin/parser_test.go.tmpl`
6. `rules/${PLUGIN_NAME}_rules.yaml` ← `.claude/templates/plugin/plugin_rules.yaml.tmpl`
7. `falco.yaml` ← `.claude/templates/plugin/falco.yaml.tmpl`
8. `falco-local.yaml` ← `.claude/templates/plugin/falco-local.yaml.tmpl`
9. `falco-docker.yaml` ← `.claude/templates/plugin/falco-docker.yaml.tmpl`
10. `go.mod` ← `.claude/templates/plugin/go.mod.tmpl`
11. `Makefile` ← `.claude/templates/plugin/Makefile.tmpl`
12. `README.md` ← `.claude/templates/plugin/README.md.tmpl`
13. `.github/workflows/ci.yml` ← `.claude/templates/plugin/ci.yml.tmpl`
14. `.github/workflows/e2e-test.yml` ← `.claude/templates/plugin/e2e-test.yml.tmpl`
15. `.github/workflows/release.yml` ← `.claude/templates/plugin/release.yml.tmpl`
16. `.gitignore` ← `.claude/templates/plugin/gitignore.tmpl`
17. `LICENSE` ← `.claude/templates/plugin/LICENSE.tmpl`
18. `.golangci.yml` ← `.claude/templates/plugin/golangci.yml.tmpl`
19. `CLAUDE.md` ← `.claude/templates/plugin/CLAUDE.md.tmpl`
20. `CHANGELOG.md` ← `.claude/templates/plugin/CHANGELOG.md.tmpl`

**E2E パターン JSON の分割生成**:
`e2e_pattern.json.tmpl` は全カテゴリを1ファイルに含むマスターテンプレート。
scaffold 実行時に、各カテゴリを個別の JSON ファイルとして `test/e2e/patterns/categories/` に分割生成する:
- `sqli.json`, `xss.json`, `path_traversal.json`, `cmd_injection.json`, `suspicious_agent.json`, `benign.json`, `edge_cases.json`
- 各ファイルは `{"category": "...", "patterns": [...]}` 形式（トップレベルの `categories` ラッパーは除去）

テンプレート変数の展開:

```
${PLUGIN_NAME}           → プラグイン名（小文字: apache）
${PLUGIN_NAME_UPPER}     → プラグイン名（大文字: APACHE）
${PLUGIN_NAME_CAMEL}     → プラグイン名（CamelCase: Apache）
${PLUGIN_ID}             → Plugin ID（999）
${PLUGIN_DESCRIPTION}    → プラグイン説明文
${EVENT_SOURCE}          → イベントソース名（= ${PLUGIN_NAME}）
${LOG_PATH_DEFAULT}      → デフォルトログパス
${VERSION}               → プラグインバージョン（0.1.0）
${AUTHOR}                → 作者名
${LICENSE}               → ライセンス（Apache-2.0）
${SDK_VERSION}           → Plugin SDK バージョン（0.8.1）
${LOG_FORMAT}            → ログフォーマット定義（auto/combined/common/json/custom）
${TIME_FORMAT}           → タイムスタンプフォーマット文字列

#### 2.4 ドメイン固有フィールドのコード生成（5プレースホルダー展開）

Phase 1 で収集したフィールド定義に基づいて、以下の 5 つのプレースホルダー位置にコードブロックを生成・挿入する。
**Claude Code が直接コードを生成**する方式（sed/envsubst による機械的置換ではない）。

| プレースホルダー | 展開先 | 生成内容 |
|----------------|--------|---------|
| `${DOMAIN_FIELDS_STRUCT}` | `plugin.go.tmpl`（PluginEvent）+ `parser.go.tmpl`（LogEntry） | Go 構造体フィールド定義 |
| `${DOMAIN_FIELDS_DEFS}` | `plugin.go.tmpl`（Fields()） | `sdk.FieldEntry` 配列要素 |
| `${DOMAIN_FIELDS_EXTRACT}` | `plugin.go.tmpl`（Extract()） | switch/case 文 |
| `${DOMAIN_FIELDS_MAPPING}` | `plugin.go.tmpl`（parseLine()） | LogEntry → PluginEvent マッピング |
| `${DOMAIN_FIELDS_PARSE_JSON}` | `parser.go.tmpl`（parseJSON()） | JSON → LogEntry フィールド設定 |

**生成例（HTTP ドメイン、RemoteAddr フィールド）**:

```go
// ${DOMAIN_FIELDS_STRUCT}
RemoteAddr  string  `json:"remote_addr"`

// ${DOMAIN_FIELDS_DEFS}
{Type: "string", Name: "${EVENT_SOURCE}.remote_addr", Desc: "Client IP address"},

// ${DOMAIN_FIELDS_EXTRACT}
case "${EVENT_SOURCE}.remote_addr":
    req.SetValue(event.RemoteAddr)

// ${DOMAIN_FIELDS_MAPPING}
RemoteAddr: entry.RemoteAddr,

// ${DOMAIN_FIELDS_PARSE_JSON}
if v, ok := raw["remote_addr"].(string); ok {
    entry.RemoteAddr = v
}
```

**型変換ルール**（Falco SDK 制約: Extract() は string と uint64 のみ）:
- `string` → そのまま `req.SetValue(event.Field)`
- `uint64` → そのまま `req.SetValue(event.Field)`
- `int`（LogEntry 型）→ PluginEvent では `uint64` に変換: `Field: uint64(entry.Field)`
```

### 3. Phase 3: 検証

#### 3.1 環境判定ビルド

```bash
# OS判定
OS=$(uname -s)

if [ "$OS" = "Linux" ]; then
  # Linux: フルビルド
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -o lib${PLUGIN_NAME}-plugin.so ./cmd/plugin-sdk/
  file lib${PLUGIN_NAME}-plugin.so
  # 期待: "ELF 64-bit LSB shared object"
else
  # macOS: 静的解析のみ（CGO+GOOS=linux -buildmode=c-shared はクロスコンパイル不可）
  echo "INFO: macOS環境のため go vet のみ実行"
  go vet ./...
fi
```

#### 3.2 テスト実行

```bash
go test ./...
```

#### 3.3 ルール検証

```bash
# falcoがインストールされている場合のみ
if command -v falco &> /dev/null; then
  falco -V rules/${PLUGIN_NAME}_rules.yaml
else
  echo "INFO: falcoが未インストール。YAML構文チェックで代替。"
  yamllint rules/${PLUGIN_NAME}_rules.yaml 2>/dev/null || echo "yamllintも未インストール。CI/CDで検証推奨。"
fi
```

## コンテキスト補完情報

### 参照すべきドキュメント

- `CLAUDE.md`: プロジェクト全体のガイドライン
- `PROBLEM_PATTERNS.md`: 過去の失敗パターンと対策
- `cmd/plugin-sdk/nginx.go`: 参照実装（プラグイン本体）
- `pkg/parser/nginx.go`: 参照実装（パーサー）
- `pkg/parser/regex_simple.go`: 参照実装（セキュリティ検出）
- `.claude/templates/plugin/`: テンプレートファイル群

### 過去の失敗パターン（要注意）

1. **P001: macOSバイナリリリース** — macOSでビルドしたバイナリをLinux用として配布しない。Phase 3で環境判定ビルドを実行
2. **P002: -buildmode=c-shared忘れ** — Makefileテンプレートに`-buildmode=c-shared`を必ず含める
3. **P004: GOB nil map panic** — `Headers map[string]string`は必ず`make(map[string]string)`で初期化
4. **P008: load_plugins欠落** — `falco.yaml`に`load_plugins`ディレクティブを必ず含める
5. **P009: レート制限でアラート抑制** — `falco.yaml`に`rate: 0`, `max_burst: 0`を設定
6. **P010: Extract()フィールド処理漏れ** — `Fields()`で定義した全フィールドを`Extract()`で処理
7. **P014: ファイルシーク位置初期化** — `Open()`で`file.Seek(0, io.SeekEnd)`を実行

### Falco Plugin SDK 必須メソッド

```
1. Info()        → プラグインメタデータ（ID, Name, EventSource等）
2. InitSchema()  → 設定スキーマ（JSON Schema）
3. Init()        → プラグイン初期化（デバッグモード、設定パース）
4. Open()        → インスタンス生成・ファイル監視開始
5. Fields()      → 抽出可能フィールド定義
6. Extract()     → フィールド値の抽出（GOBデコード + フィールド分岐）
7. NextBatch()   → イベントバッチの配信（GOBエンコード）
+ Close()        → リソースクリーンアップ（推奨: 未実装だとリソースリーク）
```

## 成功基準

| ID | 基準 | 検証方法 |
|----|------|----------|
| SC-001 | 生成されたコードがコンパイル可能 | Linux: `go build -buildmode=c-shared` / macOS: `go vet ./...` |
| SC-002 | 基本テストがパス | `go test ./...` |
| SC-003 | 必須7メソッドがすべて実装されている | ソースコード検査 |
| SC-004 | `source: <plugin_name>` が初期ルールに含まれている | ルールファイル検査 |
| SC-005 | Plugin IDが設定されている（開発用999） | Info()メソッド確認 |
| SC-006 | Close()メソッドが実装されリソース解放を行う | ソースコード検査 |
| SC-007 | デバッグログシステムが環境変数で制御可能 | `FALCO_<PLUGIN>_DEBUG=true`で確認 |
| SC-008 | イベントチャネルオーバーフロー処理が実装されている | 非ブロッキング送信のコード確認 |

## 重要な注意事項

- テンプレートファイルは `.claude/templates/plugin/` から読み込む
- macOSでは `CGO_ENABLED=1 GOOS=linux -buildmode=c-shared` は実行不可（LC-001）
- 新規リポジトリにはセルフホストランナーがない可能性がある（LC-002）。CI/CDテンプレートは両オプション（self-hosted / GitHub-hosted）を提供
- mapフィールド（Headers等）は必ず `make(map[string]string)` で初期化（P004）
- `load_plugins` ディレクティブを falco.yaml に必ず含める（P008）
- `source: <plugin_name>` を全ルールに必ず指定（P003）
- Plugin IDは開発用に `999` を使用し、本番登録時にFalco Registryから正式IDを取得
- `Fields()` で定義した全フィールドは `Extract()` でも処理すること（P010）

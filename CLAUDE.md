# CLAUDE.md — falco-plugin-claude-code

## プロジェクト概要

Claude Code 用 脅威検知 Falco プラグイン (`falco-plugin-claude-code`) の開発リポジトリ。
falco-plugin-dev-kit のスキル群（7 スキル + 1 ワークフローエージェント + 23 テンプレート）を
インストール済み。Phase 0〜6 のワークフローに沿って Claude Code Hook イベントを Falco の
event source `claude_code` として取り込むプラグインを生成・実装・テスト・ビルドする。

### 基本設計（要件 v3 ドキュメント準拠）

| 項目 | 値 |
|------|-----|
| plugin name | `claude-code` |
| event source | `claude_code` |
| field prefix | `claude_code.*` |
| 初期バージョン | `v0.1.0` |
| 一次入力 | `~/.claude/security/events.jsonl`（`claude-code-security-logger` が出力する正規化 JSONL） |
| 検知方針 | detect-first（v0.1）。block は別途 Hook policy で実装 |
| 主対象環境 | macOS local runtime / Linux runtime / enterprise managed |
| 低遅延目標 | Hook 発火 → Falco alert で p95 1 秒以内（最低条件 p95 5 秒以内） |

詳細は `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md` を参照。

## 主要コマンド

```bash
# プラグインディレクトリ生成後（claude-code/ 配下）で実行
make build           # ビルド（OS自動検出: .dylib / .so）
make build-release   # リリース用最適化ビルド
make test            # ユニットテスト
make vet             # 静的解析
make e2e             # Level 1 + Level 2 E2E テスト
make e2e-pattern     # Level 1 パターンテスト
make e2e-pipeline    # Level 2 パイプラインテスト
make verify          # ELF バイナリ検証（Linux）
make package         # リリースパッケージ（ビルド + 検証 + チェックサム）
make clean           # ビルドアーティファクト削除
```

## 開発ワークフロー（dev-kit）

Claude Code 内で以下のスラッシュコマンドが利用可能（`.claude/skills/` から自動検出）:

| スキル | コマンド | 用途 |
|--------|---------|------|
| plugin-scaffold | `/plugin-scaffold claude-code json` | プラグイン初期構造の対話生成 |
| plugin-parser | `/plugin-parser json [sample-log]` | パーサー実装支援 |
| plugin-rules | `/plugin-rules [action] [type]` | Falco ルール作成・検証 |
| plugin-test | `/plugin-test [action] [type]` | テスト作成・実行・レポート |
| plugin-build | `/plugin-build [action] [target]` | ビルド・検証・パッケージング |
| plugin-debug | `/plugin-debug [error-or-symptom]` | エラー・障害のトラブルシューティング |
| dev-kit-feedback | `/dev-kit-feedback [plugin-path]` | テンプレート改善フィードバック |

`plugin-dev-workflow` エージェントは Phase 0〜6 を順次実行し、要件確認からビルド検証まで自律実行する。
「新しい Falco プラグインを作成してください」のような依頼で起動する。

## リポジトリ構成（インストール済み）

```
falco-plugin-claude_code/
├── .claude/
│   ├── templates/plugin/       # 23 テンプレート (.tmpl)
│   ├── skills/                 # 7 スキル
│   │   ├── plugin-scaffold/    # 初期構造生成
│   │   ├── plugin-parser/      # パーサー実装
│   │   ├── plugin-rules/       # ルール作成
│   │   ├── plugin-test/        # テスト作成・実行
│   │   ├── plugin-build/       # ビルド・検証
│   │   ├── plugin-debug/       # トラブルシューティング
│   │   └── dev-kit-feedback/   # フィードバック収集
│   └── agents/
│       └── plugin-dev-workflow.md  # 統合ワークフロー
├── docs/
│   ├── claude_code_falco_plugin_requirements_2026-04-26_v3.md  # 要件定義 v3
│   └── prompt.txt
├── PROBLEM_PATTERNS.md         # P001-P021 + A コード失敗パターン集
└── CLAUDE.md                   # 本ファイル
```

## テンプレート変数

| 変数 | 説明 | claude-code での値（推奨） |
|------|------|---------------------------|
| `${PLUGIN_NAME}` | プラグイン名（小文字） | `claude-code` |
| `${PLUGIN_NAME_UPPER}` | プラグイン名（大文字） | `CLAUDE_CODE` |
| `${PLUGIN_NAME_CAMEL}` | プラグイン名（CamelCase） | `ClaudeCode` |
| `${PLUGIN_ID}` | Plugin ID | `999`（開発用 / 本番は Falco Registry で取得） |
| `${PLUGIN_DESCRIPTION}` | プラグイン説明文 | `"Claude Code Hook event monitoring plugin for Falco"` |
| `${EVENT_SOURCE}` | イベントソース名 | `claude_code` |
| `${LOG_PATH_DEFAULT}` | デフォルトログパス | `~/.claude/security/events.jsonl` |
| `${LOG_FORMAT}` | ログフォーマット | `json` |
| `${TIME_FORMAT}` | タイムスタンプフォーマット | `2006-01-02T15:04:05Z07:00`（RFC3339） |
| `${VERSION}` | バージョン | `0.1.0` |
| `${AUTHOR}` | 作者名 | `takaosgb3` / `FALCOYA` |
| `${SDK_VERSION}` | Plugin SDK バージョン | `0.8.1` |

## ドメイン固有プレースホルダー

テンプレート内の以下のプレースホルダーは、scaffold スキルがフィールド定義に基づいてコードブロックを生成する:

| プレースホルダー | 展開先 | 生成内容 |
|----------------|--------|---------|
| `${DOMAIN_FIELDS_STRUCT}` | plugin.go + parser.go | Go 構造体フィールド |
| `${DOMAIN_FIELDS_DEFS}` | plugin.go Fields() | sdk.FieldEntry 配列 |
| `${DOMAIN_FIELDS_EXTRACT}` | plugin.go Extract() | switch/case 文 |
| `${DOMAIN_FIELDS_MAPPING}` | plugin.go parseLine() | LogEntry → PluginEvent |
| `${DOMAIN_FIELDS_PARSE_JSON}` | parser.go parseJSON() | JSON → LogEntry |

## 重要な制約（Critical Constraints）

### P002: -buildmode=c-shared 必須
Makefile に `-buildmode=c-shared` を必ず含める。このフラグがないと Falco がプラグインをロードできない。

### P004: GOB nil map panic
`Headers map[string]string` 等の map 型は必ず `make()` で初期化。parser.go の `Parse()` 内で初期化を保証すること。Claude Code の hook context は metadata/headers を持つため特に重要。

### P008: load_plugins 必須
`falco.yaml` に `load_plugins: [claude-code]` を必ず含める。欠落するとプラグインが無視される。

### P010: Fields/Extract 一致
`Fields()` で定義した全フィールドを `Extract()` で処理すること。不一致はランタイムエラーになる。

### P003: source 指定必須
全ルールに `source: claude_code` を含める。欠落するとルールが syscall ソースに適用される。

### P014: ファイルシーク位置
`Open()` で `file.Seek(0, io.SeekEnd)` を実行。これを忘れるとプラグイン起動時に既存ログを全行再処理し、大量重複アラートが発生する。Claude Code の events.jsonl は累積するため必須。

### Falco SDK 型制約
`Extract()` で使用可能な型は `string` と `uint64` のみ。`int` や `float64` は `uint64` に変換する。
bool は文字列化（`"true"`/`"false"`）または `uint64`（`0`/`1`）に変換する。

### Claude Code 固有の追加制約

- **JSONL ローテーション対応**: hook logger は events.jsonl をローテートする可能性がある。
  fsnotify + polling fallback + rotation reopen を実装する（要件 v3）。
- **redaction 前提**: hook logger 側で秘密情報の redaction を行う設計。プラグインは生のシークレット値を
  そのまま Falco alert に流さない（テストデータでも実シークレットを使わない）。
- **OTel と非結合**: v0.1 では OTLP receiver を持たない。OpenTelemetry 経路は相関用。

## 参照実装

```
/Users/takaos/lab/falco-plugin-openclaw/
```

OpenClaw AI アシスタントログ監視プラグイン。JSONL 入力、複数セキュリティカテゴリ、parseJSON 実装、
fsnotify によるファイル tail の参照点。Claude Code とは入力スキーマが異なるため、
構造はそのまま流用せず参考としてのみ使用する。

## 失敗パターン集

`PROBLEM_PATTERNS.md` に P001-P021 + A コード（nginx-proxy 由来パターン）を収録。
新規実装・デバッグ時は必ず参照すること。

## 参照すべきドキュメント

| ドキュメント | 内容 |
|-------------|------|
| `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md` | 要件定義 v3（最新版） |
| `PROBLEM_PATTERNS.md` | P001-P021 + A コード失敗パターン集 |
| `.claude/agents/plugin-dev-workflow.md` | Phase 0〜6 統合ワークフロー |
| `.claude/skills/plugin-scaffold/SKILL.md` | スキャフォールディング手順 |
| `.claude/skills/plugin-parser/SKILL.md` | パーサー実装手順 |
| `.claude/skills/plugin-rules/SKILL.md` | ルール作成・検証手順 |
| `.claude/skills/plugin-test/SKILL.md` | テスト作成・実行手順 |
| `.claude/skills/plugin-build/SKILL.md` | ビルド・パッケージング手順 |
| `.claude/skills/plugin-debug/SKILL.md` | トラブルシューティング |
| `.claude/templates/plugin/*.tmpl` | コード生成テンプレート（23 個） |

## 開発開始手順

1. 要件 v3 ドキュメントを再確認: `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`
2. Phase 0（要件確認）: scaffold 用パラメータ（プラグイン名・event source・フィールド一覧・検出カテゴリ）を確定
3. Phase 1（scaffold）: `/plugin-scaffold claude-code json` または「Claude Code 用の Falco プラグインを作成してください」と依頼してワークフローエージェント起動
4. Phase 2 以降は plugin-dev-workflow が順次実行

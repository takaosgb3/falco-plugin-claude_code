# Rehearsal Round 1: Phase 0-1（要件確認・scaffold）

## リハーサル想定シナリオ
新規実装者が `docs/tasks/detailed_task_definition.md` と要件 v3 を頼りに、
今すぐ `/plugin-scaffold claude-code json` を実行して claude-code プラグインの初期構造を生成する。

## 抽出論点（実装者として手が止まった箇所）

### RH1-01: scaffold 入力ワークシートが存在しない【情報十分性】
**詰まる場面**: scaffold スキルが対話的に質問するフィールド一覧（要件 §10.2 の `claude_code.*` 約 28 フィールド）と各テンプレート変数（PLUGIN_NAME, AUTHOR, VERSION, SDK_VERSION 等）を一覧で投入したいが、要件 §27.1 の表は「テンプレート変数」、要件 §10.2 は「Falco field」で、scaffold が聞く順序や形式が揃っていない。実装者は両方の表を手動でマージする必要がある。

**修正対象**: 要件 §27 末尾に **scaffold 入力ワークシート（27.3）** を追加し、対話で投入する全項目を 1 ヶ所にまとめる。

---

### RH1-02: 「`claude-code/` ディレクトリ」のレイアウトが曖昧【手順実行可能性】
**詰まる場面**: 詳細タスク定義書 §8.3.2 で「`claude-code/` ディレクトリ生成」、要件 §27.1 で「Repository: `falco-plugin-claude-code`」とあるが、本リポ `falco-plugin-claude_code`（アンダースコア） 内で scaffold するときに:
- 本リポ root 直下にプラグイン構造を展開するのか
- `claude-code/` サブディレクトリを作って中に展開するのか
- 別リポジトリ `falco-plugin-claude-code` を新規作成するのか

3 通りの解釈が成り立ってしまう。

**修正対象**: 詳細タスク定義書 §8.3.2 のロードマップに「**本リポ root 直下に展開**（既存の `docs/` `.claude/` と並列）」を明示。

---

### RH1-03: hook logger は新規開発が必要だが明示されていない【情報十分性】
**詰まる場面**: 要件 §22.1 個人 macOS 導入の手順 1 で `install -m 0755 claude-code-security-logger-darwin-arm64 ...` とあるが、このバイナリは scaffold で生成されない（dev-kit テンプレートに含まれず、要件 §6.1 で要件のみ）。実装者は「どこからこのバイナリを取ってくるか」で詰まる。

**修正対象**: 要件 §22.1 冒頭に「**前提**: `claude-code-security-logger` は本プロジェクトで新規実装する（要件 §6.1）。このバイナリは dev-kit scaffold で生成されないため、`cmd/claude-code-security-logger/` を Phase 1 で手動で追加し、Phase 2 で実装する」を追記。

---

### RH1-04: AUTHOR の値が二択で確定していない【設定値整合性】
**詰まる場面**: 要件 §27.1 で `AUTHOR | takaosgb3 / FALCOYA` と二択。Go module path は `github.com/${AUTHOR}/${PLUGIN_NAME}` （詳細タスク §7.5）なので **どちらかに確定**しないと scaffold 後の `go.mod` がエラーになる。

**修正対象**: 要件 §27.1 で「`takaosgb3`（GitHub user / 開発者）を主、`FALCOYA` は OSS organization 表記（README 等のクレジット用）」と区別。

---

### RH1-05: 詳細タスク §8.3.2 の WF-Phase 0 入力例が不完全【情報十分性】
**詰まる場面**: 詳細タスク §8.3.2 で「`LOG_FORMAT="json", LOG_PATH_DEFAULT="~/.claude/security/events.jsonl"` を入力」とあるが、scaffold が実際に聞くテンプレート変数は他にも多数（PLUGIN_NAME, EVENT_SOURCE, PLUGIN_ID, VERSION, SDK_VERSION, AUTHOR, TIME_FORMAT, LICENSE, YEAR, PLUGIN_DESCRIPTION, LOG_SOURCE 等）。

**修正対象**: 詳細タスク §8.3.2 で「**完全な入力ワークシートは要件 §27.3（RH1-01 で新設）を参照**」と注記。

---

## 観点別件数

| 観点 | 件数 |
|---|---:|
| 1. 手順実行可能性 | 1（RH1-02） |
| 2. 情報十分性 | 3（RH1-01, RH1-03, RH1-05） |
| 3. 設定値整合性 | 1（RH1-04） |
| 4. テスト実行可能性 | Round 3 で扱う |

合計 5 件抽出。Round 1 で適用予定。

## Round 1 適用済み

| # | 章 | 修正 | 状態 |
|---|----|------|------|
| RH1-01 | 要件 §27.3 新設 | scaffold 入力ワークシート（19 項目）を新設、対話入力の正解セットを明示 | 適用 |
| RH1-02 | 詳細タスク §8.3.2 | 「root 直下に展開、`claude-code/` サブディレクトリは作らない」を明示。リポ名の差異も注記 | 適用 |
| RH1-03 | 要件 §22.1 冒頭 | 「hook logger は本プロジェクトで新規実装、scaffold には含まれない」を前提注記 | 適用 |
| RH1-04 | 要件 §27.1 | Author を `takaosgb3`（GitHub user）と `FALCOYA`（organization 表記）に役割分離 | 適用 |
| RH1-05 | 詳細タスク §8.3.2 | 要件 §27.3 への参照（RH1-01 で新設） | 適用（RH1-02 の修正に同梱） |

合計 5 件適用、Round 1 抽出 5 件全件解消。


# Rehearsal Round 2: Phase 2-3（parser 実装・rules 作成）

## リハーサル想定シナリオ
Phase 1 scaffold 完了後、Phase 2 で parser / detector / hook logger を実装し、Phase 3 で rules pack（`rules/claude_code_rules.yaml`）を書く。

## 抽出論点

### RH2-01: T-002〜T-018 の Falco condition 例が示されていない【情報十分性】
**詰まる場面**: 要件 §12.3 で初期ルール例として T-001（Dangerous Bash Command）の condition 1 件のみ。残り T-002〜T-018（17 件）の condition 例は要件にも詳細タスクにもない。実装者は「T-007 MCP Config Changed の Falco condition をどう書くか？」を独自判断で書く必要がある。

特に複雑な T:
- T-014 Agent Runaway / Tool Storm: `claude_code.tool_count > X` の閾値が不明
- T-008 Suspicious MCP Tool Use: `mcp__*` 接頭辞の icontains 例がない
- T-016 Config Policy Downgrade: 「`disableBypassPermissionsMode` 解除」の condition 表現

**修正対象**: 要件 §29 付録 B の各ルール名の下に **概念的な condition 雛形**を追記、または §12.3 を「初期ルール例（代表 3 件）」に拡張。

---

### RH2-02: parseLine() の boolean → string 変換例コードがない【情報十分性】
**詰まる場面**: 要件 §10.2 で「dropped: false → claude_code.dropped: "false" に変換」と注釈があるが、Go 実装で具体的にどう書くかの例がない。`strconv.FormatBool()` を使うのか、`if dropped { return "true" } else { return "false" }` か。Falco SDK の `req.SetValue()` の引数型（string / uint64）と整合する書き方が要件に未明示。

**修正対象**: 要件 §10.2 の型変換注釈の下に Go コード例（`strconv.FormatBool(event.Dropped)`）を追加。

---

### RH2-03: risk_type の値リスト（命名規約）が未定義【情報十分性】
**詰まる場面**: 要件 §10.1 サンプルで `"risk_type": "dangerous_bash"` と書かれているが、T-001 = Dangerous Bash Command。命名規約は:
- `dangerous_bash` (snake_case)
- `T-001` (ID 直記)
- `DangerousBashCommand` (CamelCase)

3 通り考えられる。detector 側で実装する `risk_type` の値が決まっていないと、ルール側で `claude_code.risk_type = "..."` の比較ができない。

**修正対象**: 要件 §10.2 `claude_code.risk_type` の説明欄に「命名規約: snake_case で T-* と 1:1 対応（例: T-001 → `dangerous_bash`）」を明示。または §12.1 の表に「risk_type 値」列を追加。

---

### RH2-04: 「初期ルールは代表的なパターンのみ」の明示【情報十分性】
**詰まる場面**: 要件 §12.3 の T-001 condition 例は `rm -rf /` 等 5 パターンだが、「rm -rf .」「rm -rf *」「rm -fr /」等の variant はカバーされていない。実装者は「これで網羅的か？」と悩む。

**修正対象**: 要件 §12.3 末尾に「**本例は代表パターンのみ**。網羅性は Phase 4 のテストで検証し、benign / edge_cases パターンと併せてカバレッジ調整する」を明示。

---

### RH2-05: 条件付き昇格 priority の Falco rule 書き分け方が未明示【情報十分性】
**詰まる場面**: 要件 §12.1 priority 凡例で「ベース priority は NOTICE だが `claude_code.risk_score >= 70` 等で WARNING に昇格」とあるが、Falco rule の **書き分け方** が不明。Falco の制約上 1 ルールで動的 priority は出せないので、実装者は次のいずれかの方法を選ぶ:
- 2 ルールに分ける（NOTICE 用と WARNING 用、condition で `risk_score < 70` / `risk_score >= 70` を分岐）
- detector 側で priority を決定し、ルールはそれを参照する macro として書く
- detector の出力 `severity` field をルールが `claude_code.severity in (...)` で評価

実装者は判断に時間を使う。

**修正対象**: 要件 §12.1 priority 凡例を拡張し、「Falco 仕様上 1 ルール 1 priority。条件付き昇格は **2 ルール（NOTICE 用 / WARNING 用）に分割** して書く方針を v0.1 で採用」を明示。

---

## 観点別件数

| 観点 | 件数 |
|---|---:|
| 1. 手順実行可能性 | 0 |
| 2. 情報十分性 | 5（RH2-01 〜 RH2-05） |
| 3. 設定値整合性 | 0 |
| 4. テスト実行可能性 | Round 3 で扱う |

合計 5 件。Round 2 で適用予定。

## Round 2 適用済み

| # | 章 | 修正 | 状態 |
|---|----|------|------|
| RH2-01 | 要件 §12.4 新設 | T-002〜T-018 の condition 雛形を 17 行の表で追加 | 適用 |
| RH2-02 | 要件 §10.2 型変換注釈 | Go 実装例（strconv.FormatBool, uint64 cast, negative clamp）を追加 | 適用 |
| RH2-03 | 要件 §10.2 risk_type 行 | 命名規約（T-001 → dangerous_bash 等）を 18 件分明示 | 適用 |
| RH2-04 | 要件 §12.3 末尾 | 「本例は代表パターンのみ。網羅性は Phase 4 で検証」を注記（§12.4 と一体で対応） | 適用 |
| RH2-05 | 要件 §12.1 priority 凡例 | 「Falco 1 ルール 1 priority。条件付き昇格は 2 ルール分割」の v0.1 方針を明示 | 適用 |

合計 5 件適用、Round 2 抽出 5 件全件解消。


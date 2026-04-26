# Rehearsal Round 3: Phase 4（3 層 E2E テスト）

## リハーサル想定シナリオ
実装者として「**各テストを今すぐ実行したい**」とする。fixture を作り、テストコマンドを叩き、判定する流れをすべて再現できるか。

## 抽出論点

### RH3-01: fixture の配置場所と命名規則が未明示【テスト実行可能性】
**詰まる場面**: 要件 §20.2 で 15 件の fixture 名が列挙されているが、ファイルの **配置先ディレクトリ**が未明示。
- `test/e2e/patterns/categories/` の下なのか
- `test/fixtures/` の下なのか
- カテゴリ別サブディレクトリを切るのか

詳細タスク T2-3（Level 1 E2E パターンテスト）に「`test/e2e/patterns/categories/*.json` を走査」とあるが、要件 §20.2 の fixture 名がそのまま `categories/*.json` の中なのか、別ディレクトリなのか整合不明。

**修正対象**: 要件 §20.2 に「**配置先**: `test/fixtures/hook_events/<event_name>/<scenario>.json`」として標準パスを明示（または T2-3 と整合）。

---

### RH3-02: TEST-001〜TEST-008 の実行コマンドが未明示【テスト実行可能性】
**詰まる場面**: 受入基準として TEST-001（Unit）, TEST-002（Level 1）, TEST-003（benign）, TEST-004（redaction）, TEST-005（Level 2）, TEST-006（Level 3）, TEST-007（macOS native）, TEST-008（latency p95 ≤ 5s）が並ぶが、それぞれを実行する **具体的コマンド** が要件にない。実装者は「TEST-004 redaction はどう実行するか？」で詰まる。

**修正対象**: 要件 §20.3 の各 TEST-* に「実行コマンド」列を追加。

---

### RH3-03: latency test の実行方法が未明示【テスト実行可能性】
**詰まる場面**: TEST-008「latency test が p95 ≤ 5s」とあるが、具体的に:
- どの fixture を投入するか
- 何回繰り返すか（N=100, 1000？）
- 計測区間はどこ（`logger 入力時刻 → Falco alert 出力時刻`?）
- p95 をどう算出するか（Go の test 内、外部スクリプト?）

すべて不明。実装者は「自分で latency 計測ハーネスを書く」必要がある。

**修正対象**: 要件 §20.3 TEST-008 に latency 計測手順（fixture / 反復数 / 計測区間 / p95 算出ロジック）を追加。

---

### RH3-04: TC-5-01「100 events/sec」の負荷生成方法が未明示【テスト実行可能性】
**詰まる場面**: 詳細タスク §2.5 T1-3 で「TC-5-01 100 events/sec のテスト」「TC-5-02 バッファオーバーフロー」とある。100 events/sec をどう発生させるか、Go の test 内で `for i := 0; i < 100; i++ { writeToLog(...); time.Sleep(10ms) }` を回すのか、`fortio` 等を使うのか不明。バッファオーバーフローのテストもバッファサイズと書き込み数の関係が必要だが、未明示。

**修正対象**: 詳細タスク §2.5 T1-3 の検証コマンドの下に「**負荷生成サンプル**」を追加、または要件 §8.3 末尾に「Level 2 テストの負荷生成パターン」を追加。

---

### RH3-05: `rotation_scenario` の具体手順が未明示【テスト実行可能性】
**詰まる場面**: 要件 §20.2 fixture リストの最後に `rotation_scenario | log rotate` のみ。具体的な手順:
1. 既存 events.jsonl に N 行書き込む
2. `mv events.jsonl events.jsonl.1` （rotate 模擬）
3. 同名で新規 events.jsonl を作成
4. 新規 events.jsonl に M 行書き込む
5. plugin が新規ファイルから読めるか確認

これがどこにも書かれていない。実装者はファイル rename / truncate / inode 変更のどのパターンを再現するか自分で決める必要がある。

**修正対象**: 要件 §20.2 末尾に rotation_scenario の手順詳細を追加。

---

## 観点別件数

| 観点 | 件数 |
|---|---:|
| 1. 手順実行可能性 | 0 |
| 2. 情報十分性 | 0 |
| 3. 設定値整合性 | 0 |
| 4. **テスト実行可能性** | **5（RH3-01 〜 RH3-05）** |

合計 5 件。Round 3 で適用予定。

## Round 3 適用済み

| # | 章 | 修正 | 状態 |
|---|----|------|------|
| RH3-01 | 要件 §20.2 冒頭 | fixture 配置先「`test/fixtures/hook_events/<event_name>/<scenario>.json`」を標準化、Level 1 patterns との関係を明示 | 適用 |
| RH3-02 | 要件 §20.3 | TEST-001〜TEST-008 に「実行コマンド」「判定方法」の 2 列を追加 | 適用 |
| RH3-03 | 要件 §20.3.1 新設 | TEST-008 latency 計測手順（fixture / N=1000 / 計測区間 / p95 算出ロジック）を擬似コードで追加 | 適用 |
| RH3-04 | 詳細タスク §2.5 T1-3 | TC-5-01 / TC-5-02 の負荷生成サンプルコードを追加 | 適用 |
| RH3-05 | 要件 §20.2.1 新設 | rotation_scenario の具体手順（rename rotation 5 ステップ）を擬似コードで追加 | 適用 |

合計 5 件適用、Round 3 抽出 5 件全件解消。


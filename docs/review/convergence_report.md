# 要件定義 v3 レビュー 収束レポート（5 ラウンド完了）

対象: `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`
実施日: 2026-04-26
レビュー実施者: Claude Code（Opus 4.7 1M context）

## 1. 件数推移

| Round | 抽出論点 | 適用修正 | 残（次ラウンド繰越） | 主なテーマ |
|---:|---:|---:|---:|---|
| 1 | 8 | 7 | 1 | 命名・型・SDK/Go/Falco 最低版・PROBLEM_PATTERNS マッピング |
| 2 | 8 | 8 | 0 | TOC・matcher・redaction patterns・Health check・Container・SBOM |
| 3 | 6 | 6 | 0 | TR-* / RK-* ID 体系・priority 凡例・条件式の括弧明示・OpenClaw 流用範囲 |
| 4 | 7 | 7 | 0 | §18.4 セクションレベル・doctor CLI exit code・tool 名統一・付録 D 拡張・TOC 反映 |
| 5 | 2 | 2 | 0 | §25 ロードマップに supply-chain / container を反映・README 注記 |
| **計** | **31** | **30** | **0** | （R1-08「§0/§32 重複の整理」は §32 冒頭注釈で代替対応） |

> 「抽出 31 / 適用 30」の差は Round 1 の R1-08（§0/§32 重複の構造的整理）。これは §32 の役割明示（Round 4 R4-06）で論理的に解消されており、内容的な追加修正は不要と判断した。

### 収束カーブ

```
抽出: 8 → 8 → 6 → 7 → 2
適用: 7 → 8 → 6 → 7 → 2
```

抽出件数は Round 4 で再増加しているが、これは「Round 4 で構造リファクタを集中投入」した結果であり、論点の質は Round 1〜3 の高優先度から細部詰め・新規節の TOC 反映・チェックリスト同期に推移している。

## 2. 物量変化

| 指標 | レビュー前 | レビュー後 | 増分 |
|---|---:|---:|---:|
| 行数 | 1378 | 1597 | +219 行（+15.9%） |
| 章 | 33 (§0〜§32) | 33 + sub 8 新設 | §3.3.1 / §17.1 / §18.4 / §21.3 / §22.4 / §22.5 / §27.1 / §27.2 |
| ID 体系 | 約 16 種 | 17 種（TR-*, OPS-*, SC-* 新設、R-011 拡張） | +3 種 / +個別 |
| 未完了マーカー（TODO/FIXME 等） | 0 | 0 | 変化なし |

## 3. 解消された主要論点

### 3.1 論理・型整合
- §10.1 schema (boolean) ↔ §10.2 Falco field (string) の暗黙変換に注釈
- §13 R-011 で boolean 由来 string field の比較リテラル必須化
- §12.3 condition の括弧明示

### 3.2 用語・ID 整合
- §10.2 `permission_mode` 値リスト（`auto`/`dontAsk` の誤記訂正）
- §11.4 リスク表に TR-001〜TR-010 ID 付与、§23.1 RK-* との相互参照
- §6.1.3 tool 名 `Agent` / `Task` 統一注釈
- §18.4 セクションレベル（h2 → h3）
- TOC への新規節 8 個と ID 体系の反映

### 3.3 抜け漏れ
- §27.1 ライセンス（Apache-2.0）追加
- §27.2 Plugin SDK / Go / Falco / macOS / Linux / GLIBC / アーキテクチャ最低版を明示
- §17.1 redaction 最小 regex セット
- §18.4 PROBLEM_PATTERNS（P001〜P021）と要件項目の対応マッピング
- §21.3 SBOM / cosign / SLSA / 脆弱性スキャン
- §22.4 Health check / 監視戦略（doctor CLI exit code 含む）
- §22.5 Container / Kubernetes 環境の v0.1 非対象方針
- §3.3.1 OpenClaw からの流用不可領域 / 安全再利用領域
- §29 付録 B の T-013/T-015/T-016/T-017/T-018 ルール網羅
- §28 hook event 名の公式仕様確認注釈
- §32 §0/§1/§32 役割分担の明示
- §31 付録 D チェックリスト 10 → 15 項目に拡張
- §25 ロードマップに supply-chain / container / k8s / SLSA を反映

### 3.4 数値・SLO
- §8.2 SLO の sub-budget 合計検算は整合（目標 770ms < 1s、最低 3050ms < 5s）
- §27.1 buffer / poll_interval と §8.2 throughput の整合確認

## 4. 残課題（5 ラウンド外）

| 種別 | 内容 | 取り扱い |
|---|---|---|
| 実装時の確認事項 | PROBLEM_PATTERNS.md（19770 行）のうち §18.4 マッピング表に書いた P001〜P021 の正式タイトルとの一致確認 | Phase 1 scaffold 後の実装前 spot-check で吸収 |
| 仕様確認事項 | Claude Code 公式 hook event 名（`UserPromptExpansion` / `Elicitation` / `Result` / tool 名 `Agent` vs `Task`） | リリース前テストで fixtures と差分検証（§28 冒頭注釈で明文化済） |
| 構造的軽量化 | §0/§1/§32 の散文重複は意図的（異なる切り口）として §32 冒頭注釈で運用 | §32 冒頭で役割分担を明示済（Round 4 R4-06） |
| Plugin Registry | Plugin ID 999 の正式取得 | リリース前タスク（§27.1 / FP-003 で明示） |

これらは「要件書としての精度」ではなく「実装/リリース時に消化すべきもの」であり、レビュー対象としては収束済み。

## 5. 収束評価

- **収束した**と判定する。
- 5 ラウンドを通じて 30 件の修正を適用。Round 5 で抽出件数が 2 件に低下し、いずれもロードマップ表の細部反映で構造的論点ではない。
- ID 体系・章レベル・参照リンクの整合は Round 4 で確立、Round 5 で確認。
- 6 ラウンド目を実施しても、新たに発見されるのはマイクロ表現の調整（句読点・テーブル整列）程度と見込まれる。
- **要件書 v3 は実装フェーズ（Phase 1 scaffold）に進んでよい状態**。

## 6. レビュー成果物

| ファイル | 内容 |
|---|---|
| `docs/review/PLAN.md` | 5 ラウンドの計画書 |
| `docs/review/round1.md` | Round 1 所見・適用済み 7 件 |
| `docs/review/round2.md` | Round 2 所見・適用済み 8 件 |
| `docs/review/round3.md` | Round 3 所見・適用済み 6 件 |
| `docs/review/round4.md` | Round 4 所見・適用済み 7 件 |
| `docs/review/round5.md` | Round 5 所見・適用済み 2 件 |
| `docs/review/convergence_report.md` | 本ファイル：収束評価・件数推移・残課題 |
| GitHub Issue #1 | 各ラウンドの進捗コメント（5 + 1 = 6 コメント想定） |

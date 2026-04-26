# 実装リハーサル 収束レポート（5 ラウンド完了）

対象:
- `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`
- `docs/tasks/detailed_task_definition.md`

実施日: 2026-04-26
リハーサル実施者: Claude Code（Opus 4.7 1M context）

## 1. リハーサルとレビューの違い

| 観点 | レビュー（前回まで） | リハーサル（今回） |
|---|---|---|
| 視点 | 観察者・校正者 | 実装者・デプロイ作業者 |
| 問い | 「整合は取れているか」 | 「これだけ読んで作業できるか」 |
| 抽出対象 | 整合崩れ・抜け漏れ | **詰まる箇所・判断に迷う箇所** |
| ID | RX-NN（Review） | **RHN-NN（Rehearsal）** |

## 2. 件数推移

| Round | 抽出 | 適用 | スコープ | 主な詰まり |
|---:|---:|---:|---|---|
| 1 | 5 | 5 | Phase 0-1（要件確認・scaffold） | scaffold 入力ワークシート、ディレクトリレイアウト、hook logger 前提、AUTHOR 確定 |
| 2 | 5 | 5 | Phase 2-3（parser・rules） | T-002〜T-018 condition 雛形、Go 型変換例、risk_type 命名規約、priority 昇格分割 |
| 3 | 5 | 5 | Phase 4（3 層 E2E テスト） | fixture 配置、TEST 実行コマンド/判定、latency 計測、負荷生成、rotation 手順 |
| 4 | 5 | 5 | Phase 5-6（build/release/運用） | SBOM/cosign 実装例、duration 形式、self-check 別ファイル、checksum コマンド、Linux/macOS 取得手順 |
| 5 | 3 | 3 | 最終通読 | Linux 導入手順、scaffold ワークシートリンク、Pコード食い違い Issue テンプレート |
| **計** | **23** | **23** | — | — |

## 3. 観点別累計

| 観点 | 件数 | 割合 |
|---|---:|---:|
| 1. 手順実行可能性 | 6 | 26% |
| 2. **情報十分性** | **11** | **48%** |
| 3. 設定値整合性 | 1 | 4% |
| 4. テスト実行可能性 | 5 | 22% |

→ **情報十分性**が最多。実装者として「判断に迷う」場面が多く、これらを具体化する補強が中心となった。

## 4. 件数推移（収束カーブ）

```
抽出: 5 → 5 → 5 → 5 → 3   ← Round 5 で減少
適用: 5 → 5 → 5 → 5 → 3
```

Round 1-4 で各 Phase を集中検証して 5 件ずつ安定して抽出。Round 5（最終通読）で 3 件に減少して収束。

## 5. 解消された主要詰まり

### 5.1 Phase 0-1（Round 1）
- 要件 §27.3 を新設し scaffold 入力ワークシート 19 項目を提供
- 詳細タスク §8.3.2 に「root 直下に展開、`claude-code/` 作らない」を明示
- 要件 §22.1 冒頭に hook logger 新規実装の前提を明記
- AUTHOR を `takaosgb3`（GitHub user）と `FALCOYA`（organization）に役割分離

### 5.2 Phase 2-3（Round 2）
- 要件 §12.4 を新設し T-002〜T-018 condition 雛形を 17 行で提供
- 要件 §10.2 に Go 型変換例（strconv.FormatBool, uint64 cast, negative clamp）追加
- risk_type の命名規約（T-001 → dangerous_bash 等 18 件）を明示
- priority 凡例に「Falco 1 ルール 1 priority。条件付き昇格は 2 ルール分割」の v0.1 方針

### 5.3 Phase 4（Round 3）
- fixture 配置先を `test/fixtures/hook_events/<event_name>/<scenario>.json` に標準化
- TEST-001〜TEST-008 に「実行コマンド」「判定方法」列を追加
- 要件 §20.3.1 を新設し latency 計測手順を擬似コードで提供
- 詳細タスク §2.5 T1-3 に TC-5-01/02 の負荷生成サンプル追加
- 要件 §20.2.1 を新設し rotation_scenario の rename rotation 5 ステップ擬似コード

### 5.4 Phase 5-6（Round 4）
- 要件 §21.3.1 を新設し SBOM (anchore/sbom-action) と cosign keyless 署名の最小実装例
- duration 形式を Go time.ParseDuration 互換と明示
- self-check rule を別ファイル `rules/claude_code_health.yaml` に切り出す方針
- チェックサム生成・検証コマンドを Linux/macOS 別に明記
- §22.1 個人 macOS 導入に手順 0（release artifacts ダウンロード + shasum 検証）を追加

### 5.5 最終通読（Round 5）
- 要件 §22.1.1 を新設し Linux production 導入手順を 0〜7 の番号付きで追加
- 詳細タスク §5.7 T4-4 に要件 §27.3 へのリンク追加
- 詳細タスク §3.7 T2-5 にPコード食い違い別 Issue の title / body 雛形を追加

## 6. 残課題（リハーサル外）

| 種別 | 内容 | 取り扱い |
|---|---|---|
| 実装時 | テンプレート行番号は実装着手時に grep で再確認 | §1.5.2 で既に明記 |
| 別 Issue | Pコード食い違い 8 件の整合プラン | T2-5 完了基準に Issue 雛形と推奨案 (a) を組込済 |
| 既存プラグイン回帰 | nginx-plugin / openclaw の go vet/test | ET-7 として各 Step で実施 |
| Falco 多バージョン | 0.43+ 以外での動作確認 | v0.5 で多バージョン CI を検討 |

## 7. 収束評価

- **収束した**と判定する。
- 5 ラウンドで 23 件のリハーサル詰まりを解消。
- Round 5 抽出 3 件はすべて既存ラウンドの延長で構造的論点なし。
- 「**ドキュメントだけを頼りに実装着手・デプロイ作業を完走できる**」状態に到達。

### 物量変化（リハーサル前後）

| ファイル | 前 | 後 | 増分 |
|---|---:|---:|---:|
| 要件 v3 | 1614 行（前回レビュー後） | 1759 行 | +145 行（+9.0%） |
| 詳細タスク定義書 | 2017 行（前回レビュー後） | 2106 行 | +89 行（+4.4%） |

**主な増分**: §12.4（T-002〜T-018 condition 雛形 17 行+）、§20.3.1（latency 手順）、§20.2.1（rotation 手順）、§21.3.1（SBOM/cosign）、§22.1.1（Linux 導入）、§27.3（scaffold ワークシート）。

## 8. レビュー履歴の総括（Round 1〜3 の旧レビュー + リハーサル）

| 段階 | 累計 | 主な成果 |
|---|---:|---|
| 要件 v3 旧レビュー（5 ラウンド） | 30 件 | 構造整合・ID 体系・抜け漏れ |
| 詳細タスク定義書 旧レビュー（5 ラウンド） | 36 件 | Pコードマッピング、依存関係、SC コード |
| **実装リハーサル（5 ラウンド）** | **23 件** | **実装者として詰まる箇所の具体化** |
| 総計 | **89 件** | — |

3 段階のレビュー/リハーサルを経て、要件と詳細タスクは**実装着手時の判断材料がほぼ揃った状態**に到達。

## 9. レビュー成果物

| ファイル | 内容 |
|---|---|
| `docs/review/rehearsal/PLAN.md` | 5 ラウンド計画書 |
| `docs/review/rehearsal/round{1..5}.md` | 各ラウンドの詰まり所見と修正 |
| `docs/review/rehearsal/convergence_report.md` | 本ファイル：収束評価 |
| GitHub Issue #1 | 各ラウンドの進捗コメント（5 + 1 = 6 コメント） |

## 10. 提言

実装着手時に推奨する読む順序:
1. 詳細タスク §1（全体概要） → 該当 Step（§2-§6）
2. 各タスクの「コンテキスト復元」表
3. 要件 §27.3（scaffold 入力ワークシート、Round 1 で新設）
4. 該当 Phase の参照節（§10.2 / §12.4 / §20.3 等）
5. リスク（詳細タスク §8.1）と受入条件（§8.6）

リハーサルで詰まり所として抽出された 23 件はすべて修正済のため、初見の実装者は本ガイドを読みながら作業を進められる。

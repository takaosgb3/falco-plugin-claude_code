# Round 1 レビュー所見

対象: `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`（1378 行）
観点: 章ごとに通読し、論理一貫性・章間整合・抜け漏れ・誤りを抽出。

## A. 命名・用語の整合性
- A1: 表紙、§0、§5.1、§18、§27 すべてで `plugin name = claude-code` / `event source = claude_code` / `field prefix = claude_code.*` が一致。OK。
- A2: `claude-code-security-logger` の名称は §6.1〜§22 を通じて一貫。OK。
- A3: §18.1 / §18.2 / §21.1 の artifact 名と library_path 表記は整合（`libclaude-code-plugin-{os}-{arch}.{so|dylib}`）。ただし「ファイル命名規約」を独立節で明記していないため、新規開発者が紛れる懸念。Round 2 候補。

## B. 論理一貫性・章間整合
- B1: §0 の v0.1 detect-first / NG-004 / §15 prevention 分離 / §25 v0.2 prevention layer は整合。OK。
- B2: §8.2 SLO `JSONL write → plugin ingest p95 ≤ 250ms` と §27 `poll_interval 250ms` は整合（fsnotify 主、polling fallback）。OK。
- B3: §10.1 サンプル schema で `dropped: false`（JSON boolean）、§10.2 Falco field `claude_code.dropped` は `string`（"true"/"false"）。`bool` を Falco SDK の制約で string 化することは整合だが、両者の差を本文中で明示していない。**修正対象**。
- B4: §10.2 `claude_code.permission_mode` の値リスト `default/acceptEdits/plan/auto/dontAsk/bypassPermissions` は Claude Code 公式値と乖離している可能性あり。`auto` は Claude Code 仕様上正規の permission_mode 値でない（`acceptEdits` と混同されがち）。`dontAsk` は permission_mode の値ではなく `permissions.dontAsk` 設定の話。**修正対象**。
- B5: §29 付録B の rule pack に T-013/T-015/T-016/T-017/T-018 に対応する rule が抜けている。§12.1 で 18 カテゴリ宣言→§24 で「10+ rules」必須→付録 B は 13 件。MVP として最低 10 件は満たすが、付録 B が「想定するすべて」を書いているのか「v0.1 MVP」だけ書いているのか不明確。**修正対象**。
- B6: §28 の hook event matrix で「`UserPromptExpansion` 推奨」だが、§6.1.3 の hook 設定例には登場せず、必須/推奨/任意の整合を取るための注釈が必要。Round 2 候補。

## C. 抜け漏れ
- C1: **Plugin SDK バージョン**（`0.8.1`）が §27 初期値テーブルに不在。CLAUDE.md には記載があるため要件書と整合させるべき。**修正対象**。
- C2: **Go の最小バージョン**、**Falco の最小バージョン**、**macOS / Linux の最低サポートバージョン**、**GLIBC バージョン要件**が一切記載なし。B-006 で「GLIBC互換性を確認する」とあるが具体値なし。**修正対象**。
- C3: **ライセンス**（Apache-2.0 を想定。リポジトリに LICENSE あり）が要件書本文に明記なし。**修正対象**。
- C4: **PROBLEM_PATTERNS.md（P001〜P021）への明示参照**が本文中に乏しい。§14（Parser）§18.3（Falco設定）§7（macOS）に散在する制約はあるが、対応マッピングが未整理。**修正対象**。
- C5: **Plugin ID 999 の衝突リスク**に注釈なし。§6.3 FP-003 / §27 で `999 開発用` とあるのみで「Falco Plugin Registry の予約番号と衝突しないか確認」の注意書きなし。**修正対象**。
- C6: **redaction の具体パターン**（`AKIA[0-9A-Z]{16}`、`xoxb-`、`-----BEGIN .* PRIVATE KEY-----` など）が §17 SEC-006 に列挙されているがマッチング戦略の最小例が無い。Round 2 候補。
- C7: **stakeholder（個人開発者 / セキュリティ担当 / 企業 IT 管理者）**の明示記述が弱い。§22 で導入別に分けているが、ペルソナと requirements を結ぶ説明がない。Round 3 候補。
- C8: **timezone ポリシー**（§10.1 サンプルが `+09:00` ローカル時刻、組織横断では UTC 推奨）の判断が未記載。**修正対象**。
- C9: **Health check / 死活監視**が v0.1 で未定義。§25 で v0.3 候補だが「v0.1 では何をもって正常稼働とするか」の最小要件が必要（counters のみで十分か）。Round 2 候補。
- C10: **container / Kubernetes での deploy**を §22.3 で扱っていない。Falco の主たる本番環境であるためトーンとして不自然。Round 2 候補。
- C11: **SBOM / artifact signing**は §21 に存在せず。supply-chain 対策として弱い。Round 2 候補。
- C12: **Backwards compatibility ポリシー**（plugin field の削除可否、major/minor の境界）が §10.3 schema versioning に書かれているが、plugin と rules の両方を含む統一ポリシーが必要。Round 3 候補。
- C13: §6.1.3 の matcher 例で `Bash|Read|Write|Edit|WebFetch|WebSearch|Agent|.*` は `.*` がすべてにマッチするため列挙が冗長。注意書きはあるが「初期推奨」としての具体マッチャーが提示されていない。Round 2 候補。

## D. 誤り疑い
- D1: §10.2 `permission_mode` 値リスト（B4 と同根）。**修正対象**。
- D2: §6.1.3 hook 設定の `Agent` tool 名は Claude Code 公式 tool 名としての存在が要確認。仮に存在しなくても `.*` で拾えるが、表現として誤解を招くので注釈必要。Round 2 候補。
- D3: §28 `UserPromptExpansion` が公式 hook event 名と完全一致するか要確認。Round 2 候補。
- D4: §10.2 `claude_code.dropped` の表記（B3 と同根）。**修正対象**。

## E. 構造・冗長性
- E1: §0 / §1 / §32 の三箇所に最終判断/要約が重複。情報冗長。残してよいが、Round 4/5 で簡略化検討。
- E2: 章番号と付録番号の整合（§28 が付録A）は OK だが、目次 (TOC) が無いため 1378 行を navigate しにくい。Round 2 候補。

## Round 1 で適用する修正リスト

| # | 章 | 修正内容 |
|---|----|---------|
| R1-01 | §10.1 / §10.2 | `claude_code.dropped` が string 型化されることを注釈。schema 内 boolean → Falco field string の変換を §10.2 直前に明記 |
| R1-02 | §10.2 | `permission_mode` の値リストを Claude Code 公式仕様に合わせて修正（`default/acceptEdits/plan/bypassPermissions` を主、`auto` `dontAsk` の扱いを訂正） |
| R1-03 | §29 付録B | 不足ルール（T-013/T-015/T-016/T-017/T-018）を追記。スコープを「v0.1 MVP rule pack」と明記 |
| R1-04 | §27 | Plugin SDK バージョン `0.8.1`、Go 最小版、Falco 最小版、macOS / Linux 最低サポート、GLIBC 要件、ライセンスを追加 |
| R1-05 | 新節（§19 直前または §27 後） | PROBLEM_PATTERNS.md（P001〜P021）と要件項目の対応マッピング表を追加 |
| R1-06 | §6.3 FP-003 / §27 | Plugin ID 999 衝突確認の注意書きを追加 |
| R1-07 | §10.1 / §10.3 | timestamp の timezone ポリシー（local + offset を許容、organization-wide では UTC 推奨）を明記 |
| R1-08 | §0 / §32 注釈 | 重複は維持しつつ、§1 から §0/§32 への内部参照を明示（読者の混乱回避） |

これらを Round 1 修正で適用する。

## Round 1 適用済み修正（実装後）

| # | 章 | 修正内容 | 状態 |
|---|----|---------|------|
| R1-01 | §10.2 | schema event ↔ Falco field の型変換注釈を追加 | 適用 |
| R1-02 | §10.2 | `permission_mode` の値リストを Claude Code 公式値に修正、`dontAsk` 誤記を訂正 | 適用 |
| R1-03 | §29 付録B | T-013/T-015/T-016/T-017/T-018 のルールと関連 lists/macros を追加。スコープ説明を冒頭に追記 | 適用 |
| R1-04 | §27 | §27.1（基本値）と §27.2（ツールチェーン/プラットフォーム要件）に分割し、SDK / Go / Falco / OS / GLIBC / アーキテクチャ / License を追加 | 適用 |
| R1-05 | §18.4 新設 | PROBLEM_PATTERNS（P001〜P021 + A コード）と要件項目の対応マッピング表を追加 | 適用 |
| R1-06 | §6.3 FP-003 | Plugin ID 衝突確認の注意書きと Registry 登録手順を追記 | 適用 |
| R1-07 | §10.1 | timestamp の timezone ポリシー（local + offset 許容、組織横断は UTC 推奨）を注釈 | 適用 |

合計 7 件適用。**Round 1 で抽出された 8 件中 7 件を解消。R1-08（§0/§32 重複の整理）は Round 4 以降で扱う。**

新たに Round 2 で扱うべき残課題:
- §6.1.3 matcher 例の冗長性と推奨初期値
- §22 における health check / 監視戦略
- §22.3 における container/k8s 言及
- §17 redaction の具体パターン例
- SBOM / signing
- §28 hook event 名の公式仕様確認注釈
- TOC 追加


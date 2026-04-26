# Round 3 レビュー所見

対象: Round 1+2 適用済みの要件書
観点: 数値整合・ID 体系・条件式の論理・章間細部

## A. 数値整合の確認
- A1: §8.2 SLO の sub-budget 合計
  - 目標: 20 + 250 + 500 = 770ms < 1s ✅
  - 最低条件: 50 + 1000 + 2000 = 3050ms < 5s ✅
- A2: §27.1 buffer size macOS 1024 / Linux 4096 と §8.2 throughput 100 events/sec → 10〜40 秒分の緩衝、十分。
- A3: §14.2 D-004 bounded input 10〜64KB と §6.1.1 HL-012 raw 64KB / evidence 2KB と §27.1 Max event size 64KB / Evidence max 2KB → 全て整合 ✅。
- A4: §6.1.1 HL-013 logger latency p95 20ms と §8.2 SLO は同値。OK。

## B. ID 体系の重複・抜け
- B1: §11.4 主要リスクと対策表に独立 ID 無し。`TR-*`（Threat-Risk）として付与し、§23.1 RK-* との役割（脅威モデル由来 vs 運用全般）を冒頭で明示する。**修正対象**。
- B2: §23.1 RK-* と §11.4 内容に部分重複あり（hook 無効化 / MCP 追加 / artifact 混同 / cross-source 等）。両表の関係を本文で説明し、相互参照を入れる。**修正対象**。
- B3: PROBLEM_PATTERNS.md の A コードは §18.4 で「RK-* として吸収」と書いた。番号体系（A001〜A???）は PROBLEM_PATTERNS 側で管理されているはずなので要件書では参照のみ維持。Round 4 でも特に追加修正不要。

## C. 論理ミス / 条件式
- C1: §12.3 初期ルール例の condition で
  ```
  claude_code.command icontains "curl" and claude_code.command icontains "| sh" or
  claude_code.command icontains "chmod 777" or ...
  ```
  括弧無しのため `(curl AND |sh) OR chmod777 OR ...` と評価されることになり結果としては意図通り。だが可読性とミスマッチ防止のため括弧明示が望ましい。**修正対象**。
- C2: §10.2 で boolean 系 string field の比較について Falco rule 側の運用注意が抜けている（`= "true"` / `= "false"` リテラル）。§13 に R-011 を追加。**修正対象**。
- C3: §12.1 priority 列で `NOTICE/WARNING` 表記の意味が不明確（条件依存？ 運用判断？）。表の前に凡例追加。**修正対象**。

## D. 章間細部
- D1: §24 「10+ rules」と §29 18 rules（付録B）の関係が曖昧。§24 を「最低 10、目標 18（§29 付録B 全網羅）」に明示する。**修正対象**。
- D2: §3.3 OpenClaw との比較表は対比のみで「移植不可項目」を別段で示していない。schema / permission model / MCP / Hook 概念は別物であり、コードの safe な再利用範囲を明示する。**修正対象**。
- D3: §22.4 OPS-005 doctor CLI の exit code 仕様（events.jsonl が空のとき）は実装詳細に近いので Round 4 で検討。
- D4: §0 の表と §1 / §32 の散文重複は読み手の理解を助ける構造であり、Round 4 で軽量化検討（強い修正は避ける）。

## Round 3 で適用する修正リスト

| # | 章 | 修正内容 |
|---|----|---------|
| R3-01 | §11.4 | リスク表に `TR-001`〜`TR-010` ID を付与し、冒頭で §23.1 RK-* との役割分担と相互参照を明示 |
| R3-02 | §13 | R-011 を新設: boolean 由来の string field 比較は文字列リテラル `"true"`/`"false"` を使う旨 |
| R3-03 | §12.1 直前 | priority 凡例を追加（`NOTICE/WARNING` の意味、運用閾値、エスカレーション条件） |
| R3-04 | §12.3 | 初期ルール例 condition を括弧で明示し、`(curl AND \|sh) OR chmod777 OR ...` の意図を明確化 |
| R3-05 | §3.3 末尾 | OpenClaw から流用してはいけない領域（schema / permission / MCP / Hook 概念）を列挙 |
| R3-06 | §24 | 「10+ rules」を「最低 10、目標 18（§29 付録B 全 T-001〜T-018）」に修正 |

合計 6 件を Round 3 で適用する。

## Round 3 適用済み修正

| # | 章 | 修正内容 | 状態 |
|---|----|---------|------|
| R3-01 | §11.4 | リスク表に `TR-001`〜`TR-010` ID を付与し、§23.1 RK-* との関係と相互参照を明示 | 適用 |
| R3-02 | §13 R-011 | boolean 由来の string field 比較は `"true"`/`"false"` リテラル必須を追加 | 適用 |
| R3-03 | §12.1 直前 | priority 凡例（CRITICAL/WARNING/NOTICE/NOTICE-WARNING の意味）を追加 | 適用 |
| R3-04 | §12.3 | 初期ルール例 condition の `(curl AND \|sh)` を括弧で明示 | 適用 |
| R3-05 | §3.3.1 新設 | OpenClaw からの「流用不可領域」と「安全再利用領域」を列挙 | 適用 |
| R3-06 | §24 | rules を「最低 10、目標 18」に修正 | 適用 |

合計 6 件適用。**Round 3 で抽出された 6 件すべてを解消。**

Round 4 への繰越:
- §0 / §1 / §32 の重複の整理（軽量化）
- doctor CLI の exit code 仕様
- A コードの参照体系（PROBLEM_PATTERNS 側で番号管理されているなら参照のみで十分）


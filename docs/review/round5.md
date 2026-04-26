# Round 5 レビュー所見（最終）

対象: Round 1+2+3+4 適用済み（1595 行）
観点: 全章通読での最終 sanity check、ID 連続性、ロードマップ整合、未完了マーカー検出

## A. 機械的検査
- A1: TODO / FIXME / TBD / XXX 残存マーカー: **0 件**（grep 結果空）。OK。
- A2: 各 ID 体系の連続性
  - HL-001〜HL-014: 連続 ✅
  - ES-001〜ES-007: 連続 ✅
  - FP-001〜FP-013: 連続 ✅
  - MAC-001〜MAC-008: 連続 ✅
  - OTEL-001〜OTEL-007: 連続 ✅
  - SEC-001〜SEC-012: 連続 ✅
  - OPS-001〜OPS-006: 連続 ✅（Round 2 新設）
  - SC-001〜SC-006: 連続 ✅（Round 2 新設）
  - B-001〜B-007: 連続 ✅
  - R-001〜R-011: 連続 ✅（R-011 を Round 3 で追加）
  - P-001〜P-007 (Parser): 連続 ✅
  - D-001〜D-008 (Detector): 連続 ✅
  - PRV-001〜PRV-005: 連続 ✅
  - G-001〜G-010: 連続 ✅
  - NG-001〜NG-008: 連続 ✅
  - T-001〜T-018: 連続 ✅
  - TR-001〜TR-010: 連続 ✅（Round 3 で新設）
  - RK-001〜RK-012: 連続 ✅
  - AC-001〜AC-014: 連続 ✅
  - TEST-001〜TEST-008: 連続 ✅
- A3: ID 出現延べ数 215 件、ID 行 173 件。重複・乱れなし。

## B. ロードマップと新規節の整合
- B1: §22.5 で「Container Claude Code 監視は v0.3 以降」と書いたが、§25 ロードマップ表に container/k8s 行が無い。**修正対象**。
- B2: §21.3 SC-002〜SC-004 で SBOM / cosign / SLSA を v0.2 以降必須化と書いたが、§25 ロードマップ表に supply-chain 行が無い。**修正対象**。
- B3: §22.4 OPS-* （health check）と §25 ロードマップは v0.3 で「health endpoint」と整合。OK。

## C. README 強調文言と新規追加内容の整合
- C1: §30 の必須文言 3 つ（events.jsonl 出所 / detect-first / OTel 補助）は本文と一致。OK。
- C2: SBOM / 署名 / health check を README にどう書くかは §30 の対象外（v0.1 強調文言は最低 3 つ）。Round 5 では §30 末尾に「v0.2 以降に署名検証手順を README に追加する」の注記を入れる。**修正対象**。

## D. 細部 spot-check
- D1: §10.2 boolean 型変換注釈（Round 1 R1-01 で追加）と §13 R-011（Round 3 で追加）は整合。OK。
- D2: §3.3.1 OpenClaw 流用不可領域と §31 付録 D #15 の「OpenClaw: 流用してよいのは I/O 骨格のみ」は整合。OK。
- D3: §17.1 redaction 表と §6.1.1 HL-006 / HL-012 と §27.1 Max event size 64KB / Evidence max 2KB は整合。OK。
- D4: §22.4 OPS-* と §6.3 FP-011 (counters)、§8.2 (drop counter) は整合。OK。
- D5: §4.2 G-006 (latency p95 ≤ 1s 目標 / 5s 最低) と §8.2 SLO 表、AC-009 は完全一致。OK。
- D6: §28 必須 hook 9 つと §24 v0.1 必須 events 9 つは一致。OK。
- D7: §29 付録 B の 18 ルールと §12.1 T-001〜T-018 と §24 「最低 10 / 目標 18」は整合。OK。

## E. Round 5 で適用する修正リスト

| # | 章 | 修正内容 |
|---|----|---------|
| R5-01 | §25 ロードマップ | v0.2 行に supply-chain（SBOM/cosign 必須化）、v0.3 行に container/k8s 監視検討、v0.4 行に Falco DaemonSet 統合の文言を追加 |
| R5-02 | §30 末尾 | 「v0.2 以降に SBOM / 署名検証手順も README へ追加する」の注記 |

合計 **2 件** のみ。Round 4 までに大半の論点が解消され、Round 5 はロードマップとの最終整合のみ。

## Round 5 適用済み修正

| # | 章 | 修正内容 | 状態 |
|---|----|---------|------|
| R5-01 | §25 ロードマップ | v0.2 SBOM/cosign 必須化、v0.3 container 内 Claude Code 監視、v0.4 Falco DaemonSet 統合、v0.6 SLSA L3 到達を追記 | 適用 |
| R5-02 | §30 末尾 | v0.2 以降の SBOM / 署名検証 README 注記、v0.1 でも `sha256sum -c` 案内を明記 | 適用 |

合計 2 件適用。**Round 5 で抽出された 2 件すべてを解消。**

## F. 収束評価
- 抽出件数推移: 8 → 8 → 6 → 7 → **2**
- 適用件数推移: 7 → 8 → 6 → 7 → **2**
- 累積適用: 30 件
- Round 5 で抽出された論点はすべてロードマップ整合の細部であり、構造・論理・章間整合・命名・抜け漏れは Round 4 の時点で収束済み。
- **収束済みと判定**: 5 ラウンド以上を実施しても新規高優先度の論点は出ない見込み。

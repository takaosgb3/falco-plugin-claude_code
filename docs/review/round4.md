# Round 4 レビュー所見

対象: Round 1+2+3 適用済み（1578 行）
観点: 構造的整合・章番号・残課題のクロージング・チェックリストの最新化

## A. 構造・章レベルの不整合
- A1: §18.4 が `## 18.4`（h2）で記述されている。§18 の sub-section（`### 18.4`）が正しい。**修正対象**。
- A2: 目次の §18 行で `§18.4 PROBLEM_PATTERNS との対応マッピング` を別個記載しているが §18 配下に統合される形になるので OK。
- A3: §0 / §1 / §32 の役割が冒頭で説明されておらず読者が重複と感じる可能性。§32 冒頭に役割説明を追加。**修正対象**（軽量）。

## B. Round 3 までの繰越
- B1: §22.4 OPS-005 doctor CLI の exit code 仕様が空欄。**修正対象**。
- B2: §22.4 OPS-004 self-check rule の運用注意（idle session / SessionStart で抑止 など）が未記載。**修正対象**。
- B3: §6.1.3 で tool 名 `Agent`（サンプル列挙）と `Task`（推奨初期値内）が混在。Claude Code 公式 tool 名の差異を注釈。**修正対象**。
- B4: §31 付録D のチェックリストが Round 1〜3 で追加した内容（PROBLEM_PATTERNS マッピング、SBOM、Health check、Container 非対象、redaction patterns）を反映していない。**修正対象**。

## C. Round 4 で発見した新規課題
- C1: §18.4 の P004 行で「§14.1 P-005」を参照。§14.1 に P-005（map 初期化）が存在することは確認済み。整合 OK。
- C2: §18.4 で「P001 macOS バイナリを Linux 用に配布」とあるが、PROBLEM_PATTERNS の正確な P001 タイトルは確認できていない。要件側では「Pコード」を整理メモとして扱う。Round 5 で実装時に PROBLEM_PATTERNS の本文と突合し更新。
- C3: §27.1 / §27.2 / §10.2 / §17.1 / §21.3 / §22.4 / §22.5 / §18.4 / §3.3.1 と新規セクションが多数追加された。目次（TOC）にこれらが反映されていない。**修正対象**。
- C4: §29 付録 B の rule 名がコメント `# T-001` 〜 `# T-018` で対応付けされているが、§12.1 の T-001 順序と並んでいるか再確認。
  - 付録B の順序: T-001, T-002, T-003, T-004, T-005, T-006, T-007, T-008, T-009, T-010, T-011, T-012, T-013, T-014, T-015, T-016, T-017, T-018
  - §12.1 の順序: T-001, T-002, T-003, T-004, T-005, T-006, T-007, T-008, T-009, T-010, T-011, T-012, T-013, T-014, T-015, T-016, T-017, T-018
  - 一致 ✅
- C5: §15.2 PRV-001〜PRV-005 と §25 ロードマップの対応:
  - PRV-001〜PRV-004 = optional PoC（v0.1 で実装可能だが optional）
  - PRV-005 = v0.2 以降
  - §25 v0.2 行に「optional prevention hook policy」と記載 → 整合 ✅
- C6: §10.2 末尾の field 一覧と §29 付録 B の `lists/macros` で参照する `claude_code.*` field はすべて §10.2 にあるか確認。spot-check:
  - claude_code.command → ✅
  - claude_code.tool_name → ✅
  - claude_code.event_name → ✅
  - claude_code.permission_mode → ✅
  - claude_code.permission_destination → ✅
  - claude_code.mcp_server_name / mcp_tool_name / mcp_scope → ✅
  - claude_code.risk_score → ✅
  - claude_code.session_id → ✅
  整合 ✅。

## D. Round 4 で適用する修正リスト

| # | 章 | 修正内容 |
|---|----|---------|
| R4-01 | §18.4 | h2 → h3 にレベル降格（`## 18.4` → `### 18.4`） |
| R4-02 | §22.4 OPS-005 | doctor CLI の exit code 仕様（0 / 1 / 2 / 3）を追記 |
| R4-03 | §22.4 OPS-004 | self-check rule の運用注釈（SessionEnd で N 分タイマーを reset、idle 中の偽陽性抑止） |
| R4-04 | §6.1.3 末尾 | tool 名 `Agent` / `Task` の混在に対する統一注釈 |
| R4-05 | §31 | チェックリストに PROBLEM_PATTERNS マッピング、SBOM、Health check、Container 非対象、redaction patterns を追加（10→14 項目） |
| R4-06 | §32 冒頭 | §1 / §0 / §32 の役割分担を冒頭に明示（重複ではなく役割の違い） |
| R4-07 | TOC | Round 1-3 で追加した節（§3.3.1, §17.1, §18.4, §21.3, §22.4, §22.5, §27.1, §27.2）を目次に追記 |

合計 7 件を Round 4 で適用する。

## Round 4 適用済み修正

| # | 章 | 修正内容 | 状態 |
|---|----|---------|------|
| R4-01 | §18.4 | h2 → h3 にレベル降格 | 適用 |
| R4-02 | §22.4 OPS-005 | doctor CLI の exit code 仕様（0/1/2/3, `--max-age`）追記 | 適用 |
| R4-03 | §22.4 OPS-004 | self-check rule の運用注釈（idle / SessionEnd / 業務時間帯）追記 | 適用 |
| R4-04 | §6.1.3 末尾 | tool 名 `Agent` / `Task` 統一の注釈追記、正規化方針明記 | 適用 |
| R4-05 | §31 | チェックリスト 10 → 15 項目に拡張（PROBLEM_PATTERNS マッピング、Redaction、Supply chain、Health check、Container、OpenClaw 流用範囲） | 適用 |
| R4-06 | §32 冒頭 | §0 / §1 / §32 の役割分担を明示 | 適用 |
| R4-07 | TOC | 新規節（§3.3.1, §17.1, §18.4, §21.3, §22.4, §22.5, §27.1, §27.2）と ID 体系（TR-*, RK-*, R-001〜R-011）を追記 | 適用 |

合計 7 件適用。**Round 4 で抽出された 7 件すべてを解消。**

Round 5 への繰越:
- 全章通読での残課題確認（最終 sanity check）
- PROBLEM_PATTERNS の Pコード番号と要件マッピング表の妥当性確認（Round 5 では本文 spot-check）
- 収束評価の最終集計


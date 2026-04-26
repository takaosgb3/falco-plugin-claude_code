# Round 2 レビュー所見

対象: Round 1 適用済みの要件書（行数増加、§18.4 / §22 / §27.1/27.2 が新設・拡張済み）
観点: 章間整合・参照・前提条件・残課題・新たな抜け漏れ

## A. Round 1 修正後の整合性確認
- A1: §10.2 型変換注釈と §6.3 FP-008/FP-009 (Fields/Extract 1:1) は整合。§14.2 D-* も影響なし。OK。
- A2: §10.2 `permission_mode` 値リスト改訂に伴い、§12.1 T-003 の検知例（`bypassPermissions`, `--dangerously-skip-permissions`, `skipDangerousModePermissionPrompt`）と整合。OK。
- A3: §29 付録B の T-013〜T-018 ルール追加に伴い、§24 「10+ rules」と §12.1（18 カテゴリ）の整合は強化。OK。
- A4: §27.2 で SDK バージョンを `v0.8.x（CLAUDE.md で 0.8.1 を想定）` としたが、要件書として確定値にすべき。**修正対象**。
- A5: §18.4 の P002/P008 等の参照先 §27.2/§18.3 が正しく解決される。OK。
- A6: §6.3 FP-003 の Plugin ID 注意書きと §27.1 の補足は整合。OK。

## B. 残課題（Round 1 で繰越された中〜低優先度）
- B1: §6.1.3 の hook 設定例 matcher `Bash|Read|Write|Edit|WebFetch|WebSearch|Agent|.*` が冗長で、推奨初期値の意図が不明確。**修正対象**。
- B2: §17 SEC-006 で redact 対象を列挙するのみで、最小 regex 例なし。実装ガイドラインとして弱い。**修正対象**。
- B3: §22 (運用要件) に health check / 監視戦略が未定義。**修正対象**。
- B4: §22.3 (企業導入) に container / k8s への言及なし。Claude Code は CLI 実行が主だが、Falco を k8s DaemonSet で運用する組織もあるため明記が必要。**修正対象**。
- B5: §21 に SBOM / artifact signing の方針未記載。supply-chain 対策として弱い。**修正対象**。
- B6: §28 で hook event 名（`UserPromptExpansion` 等）の公式仕様確認注釈が未記載。**修正対象**。
- B7: 1378 行超の文書に TOC（目次）が無く navigate 不能。**修正対象**。

## C. Round 2 で新たに発見した課題
- C1: §6.1.3 の hook 設定例には `SubagentStart`/`SubagentStop`/`FileChanged`/`PostToolBatch`/`PermissionDenied` が含まれず、§28 で「必須/推奨」とした hook と例の網羅性が乖離。最小例として割り切る場合、明示が必要。**修正対象**。
- C2: §6.1.3 matcher が正規表現として Claude Code に解釈されることを明示する脚注がない（`.*` の意味、`Bash|Edit` の OR 構文の意味）。**修正対象**。
- C3: §10.2 type 変換注釈の補足として、Falco rule における「false 文字列の比較は文字列リテラル `"false"` を使う」という運用注意を §13 R-* または §10.2 末尾に書くと親切。Round 3 候補。
- C4: §18.4 の表で「A コード」を「dev-kit Phase 6 の feedback で整理」とあるが、A001 / A002 等の具体内容を後続で参照するための番号体系が未定義。Round 3 候補。
- C5: §3.1 の「FALCOYA の既存資産」表で `falco-plugin-openclaw` を「近い参照実装」とするが、本要件書がそこから何を引き継ぎ、何を捨てるかの差分表が薄い。§3.3 で対比済みだが、移植不可項目（schema、permission model、MCP）の理由を一段強める。Round 3 候補（軽微）。
- C6: §11.4 主要リスクと対策表に `RK-*` ID が振られていない（§23.1 の RK-001〜012 だけ ID）。一貫性のため §11.4 にも ID を振るか、両者を統合。Round 3 候補。
- C7: §15.2 の PRV-* と §25 のロードマップで PRV-005 (Falco alert → hook policy フィードバック) を v0.2 以降と書いているが、§15.2 表の該当行も「v0.2以降」にしている。整合 OK。

## D. Round 2 で適用する修正リスト

| # | 章 | 修正内容 |
|---|----|---------|
| R2-01 | §0 直前 | TOC（目次）を追加 |
| R2-02 | §6.1.3 | matcher 推奨初期値の整理。最小例である旨と §28 への参照、SubagentStart/PostToolBatch/FileChanged の追加示唆、matcher 文法の補足 |
| R2-03 | §17 SEC-006 | redaction の最小 regex 例（AWS / GCP / Slack / GitHub PAT / JWT / RSA private key / OAuth bearer 等）を表で追加 |
| R2-04 | §22 新節 §22.4 | Health check / 監視戦略（counter exposure、self-check rule、freshness check）を追加 |
| R2-05 | §22 新節 §22.5 | container / Kubernetes 環境での扱い（v0.1 非対象、v0.3 以降検討）を明記 |
| R2-06 | §21 新節 §21.3 | SBOM / artifact signing の v0.1/v0.2 方針を追加 |
| R2-07 | §28 冒頭 | hook event 名の公式仕様確認に関する注釈を追加 |
| R2-08 | §27.2 | SDK バージョンを `v0.8.1` に確定 |

合計 8 件を Round 2 で適用する。

## Round 2 適用済み修正

| # | 章 | 修正内容 | 状態 |
|---|----|---------|------|
| R2-01 | 表紙直後 | 目次（§0〜§32 ナビゲーション）を追加 | 適用 |
| R2-02 | §6.1.3 末尾 | matcher 文法と推奨初期値、最小例である旨と §28 への参照を注釈で追加 | 適用 |
| R2-03 | §17.1 新設 | redaction 最小 regex 例（AWS / GCP / Slack / GitHub PAT / OAuth / JWT / RSA / `.env` / Cookie / Cloud creds）を表で追加 | 適用 |
| R2-04 | §22.4 新設 | Health check / 監視戦略（OPS-001〜OPS-006: selftest, counter, self-check rule, doctor CLI, OS service 任せの責務分離） | 適用 |
| R2-05 | §22.5 新設 | Container / Kubernetes 環境の v0.1 非対象方針と v0.3+ 検討事項 | 適用 |
| R2-06 | §21.3 新設 | SBOM / 署名 / SLSA / 脆弱性スキャン（SC-001〜SC-006、v0.1 必須/推奨と v0.2+ 必須化） | 適用 |
| R2-07 | §28 冒頭 | hook event 名の公式仕様確認注釈と fallback 方針 | 適用 |
| R2-08 | §27.2 | SDK バージョンを `v0.8.1` に確定 | 適用 |

合計 8 件適用。**Round 2 で抽出された 8 件すべてを解消。**

Round 3 への繰越:
- C3 false 文字列リテラルの運用注意
- C4 A コードの番号体系
- C5 OpenClaw との差分（移植不可項目）
- C6 §11.4 主要リスクへの ID 付与
- §0 / §32 重複の整理


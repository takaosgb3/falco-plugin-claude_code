# 詳細タスク定義書 レビュー作業計画（5 ラウンド）

対象:
- `docs/tasks/detailed_task_definition.md`（1987 行）
- `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`（claude-code 要件 v3、約 1614 行）

参考:
- `docs/review/devkit_v2_req.md`（dev-kit v2 要件書 v5.6, 1609 行のローカルコピー）
- `docs/review/devkit_task_defs.md`（dev-kit 既存タスク定義書 1762 行のローカルコピー）

## レビュー観点
1. **論理整合**: タスク内の Before/After / 完了基準が一貫しているか
2. **章間整合**: §1〜§8 と Step §2〜§6 の参照が崩れていないか（行番号・名称）
3. **抜け漏れ**: 要件 v2/v3 のいずれかに記載があり、本書から落ちていないか
4. **誤り**: dev-kit 既存テンプレートとの整合（行範囲・テンプレート名）
5. **MECE**: 29 タスクが要件 ID と 1:1 になっているか
6. **コンテキスト復元の十分性**: 各タスクで参照箇所が明示され、新セッションで再開可能か
7. **claude-code 要件 v3 との接続**: §18.4 マッピング・受入テスト・WF-Phase の整合
8. **PROBLEM_PATTERNS 整合**: P001〜P021 と本書の対応
9. **ID 連続性**: T1-1〜T5-9, AT-*, ET-*, TR-D*, PR-* の連続性
10. **読みやすさ**: 重複表現、冗長、不明瞭な箇所

## 各ラウンドの目標
- **Round 1**: 全章通読・基本誤り（行番号・テンプレート名）と論理整合
- **Round 2**: 章間整合・参照リンク・タスク依存
- **Round 3**: 数値・Pコード・SC コード・WF-Phase の整合
- **Round 4**: 抜け漏れ補完（テスト・運用・検証）
- **Round 5**: 全体通読で残課題確認、収束レポート

## 出力
- `docs/review/tasks/round{N}.md` … 各ラウンドの所見と修正リスト
- `docs/review/tasks/convergence_report.md` … 5 ラウンドの修正件数推移と収束評価
- GitHub Issue #1 にコメントで進捗

## ルール
- 修正は本体（detailed_task_definition.md）を直接 Edit
- 修正前後で意味が変わらないか確認
- 件数を round{N}.md に列挙し、convergence_report.md で集計

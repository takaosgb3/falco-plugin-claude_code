# 作業計画書: falco-plugin-dev-kit v2 要件定義

## 目的

falco-plugin-openclaw の開発経験から得られた知見を falco-plugin-dev-kit にフィードバックし、
次回以降のプラグイン生成品質を向上させるための要件定義書を作成する。

## 分析に使用したプロンプト

### プロンプト 1: リポジトリ理解

```
リポジトリを理解してください。
```

### プロンプト 2: openclaw からのフィードバック可能性調査

```
/Users/takaos/lab/falco-plugin-openclaw は、falco-plugin-dev-kitのagents、skillsで作られました。
falco-plugin-dev-kitのagents、skillsの改善に役に立ちますでしょうか？
```

上記プロンプトに対し、3つの並行エージェントで以下を調査:

1. **コアコード差分分析**: plugin.go, parser.go, regex_simple.go, config.go のテンプレートとの乖離
2. **テスト・ルール・CI/CD分析**: ルール、E2Eパターン、テスト、Makefile、ワークフローの差分
3. **独自進化・ドキュメント分析**: CLAUDE.md、Dockerfile、BUILD.md、CHANGELOG.md 等テンプレートにない追加物

### プロンプト 3: 要件定義書作成

```
はい。よみやすく、わかりやすく、理解しやすく、MECEで親切丁寧に詳細に要件定義書を作成してください。
参照すべき最も適切であるドキュメントを適切に特定し、必ず参照して徹底的に理解して作業を行なってください。
多くの作業を行うことも想定されるので、適切にタスクを詳細にドキュメント化し、
そのドキュメントを参照しながら適切に作業を進めてください。
作業の途中でコンテキストがなくならないように気をつけてください。
必要に応じて、context7を活用してください。
結果は、新しいブランチを作成して、Issuesを作成し、コメントに記述してください。
プロンプトも全て記録しておきたいです
```

## 参照ドキュメント一覧

### falco-plugin-dev-kit (テンプレート側)

| ファイル | 役割 |
|---------|------|
| `.claude/templates/plugin/plugin.go.tmpl` | プラグイン本体テンプレート |
| `.claude/templates/plugin/parser.go.tmpl` | パーサーテンプレート |
| `.claude/templates/plugin/regex_simple.go.tmpl` | セキュリティ検出テンプレート |
| `.claude/templates/plugin/config.go.tmpl` | 設定テンプレート |
| `.claude/templates/plugin/parser_test.go.tmpl` | パーサーテストテンプレート |
| `.claude/templates/plugin/plugin_rules.yaml.tmpl` | ルールテンプレート |
| `.claude/templates/plugin/Makefile.tmpl` | Makefile テンプレート |
| `.claude/templates/plugin/ci.yml.tmpl` | CI/CD テンプレート |
| `.claude/templates/plugin/falco.yaml.tmpl` | Falco 設定テンプレート |
| `.claude/templates/plugin/e2e_pattern.json.tmpl` | E2E パターンテンプレート |
| `.claude/skills/plugin-scaffold/SKILL.md` | スキャフォールドスキル |
| `.claude/skills/plugin-parser/SKILL.md` | パーサースキル |
| `.claude/skills/plugin-rules/SKILL.md` | ルールスキル |
| `.claude/skills/plugin-test/SKILL.md` | テストスキル |
| `.claude/skills/plugin-build/SKILL.md` | ビルドスキル |
| `.claude/agents/plugin-dev-workflow.md` | 統合ワークフローエージェント |
| `PROBLEM_PATTERNS.md` | 過去の失敗パターン集 |

### falco-plugin-openclaw (生成物側)

| ファイル | 役割 |
|---------|------|
| `cmd/plugin-sdk/plugin.go` | 実装済みプラグイン本体 |
| `cmd/plugin-sdk/plugin_test.go` | Level 2 パイプラインテスト (36 TC) |
| `pkg/parser/parser.go` | 実装済みパーサー |
| `pkg/parser/parser_test.go` | パーサーユニットテスト (40関数) |
| `pkg/parser/regex_simple.go` | AI セキュリティ検出実装 |
| `pkg/parser/config.go` | 設定実装 |
| `rules/openclaw_rules.yaml` | AI セキュリティルール |
| `test/e2e/e2e_pattern_test.go` | Level 1 E2E テスト (9 TC) |
| `test/e2e/patterns/categories/*.json` | 11 カテゴリ E2E パターン |
| `Makefile` | OS 自動検出 + E2E ターゲット付き |
| `.github/workflows/ci.yml` | CI ワークフロー |
| `.github/workflows/e2e-test.yml` | E2E + Allure ワークフロー |
| `.github/workflows/release.yml` | マルチプラットフォームリリース |
| `falco.yaml` / `falco-local.yaml` / `falco-docker.yaml` | 3 環境 Falco 設定 |
| `CLAUDE.md` | プロジェクトガイド |
| `e2e/scripts/inject_patterns.sh` | E2E パターン注入スクリプト |
| `e2e/scripts/batch_analyzer.py` | E2E 結果分析スクリプト |

## 作業ステップ

1. [x] Git リポジトリ初期化 + GitHub リポジトリ作成
2. [x] ブランチ作成 (`feat/requirements-v2`)
3. [x] 作業計画書作成 (本ファイル)
4. [x] 要件定義書作成 (`dev-kit-v2-requirements.md`)
5. [x] GitHub Issue 作成 + 要件定義書をコメントに記載
6. [x] コミット + プッシュ

## 作成された GitHub Issues

| Issue | タイトル | URL |
|-------|---------|-----|
| #1 | falco-plugin-dev-kit v2: 要件定義 (親Issue) | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/1 |
| #2 | [A] テンプレートの改善 | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/2 |
| #3 | [B] スキル定義の改善 | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/3 |
| #4 | [C] ワークフローエージェントの改善 | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/4 |
| #5 | [D] 新規追加項目 | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/5 |
| #6 | [E] 非機能要件 | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/6 |

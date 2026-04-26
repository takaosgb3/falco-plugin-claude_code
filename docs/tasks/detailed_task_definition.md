# falco-plugin-dev-kit v2 詳細タスク定義書（claude-code 視点）

| 項目 | 内容 |
|------|------|
| 文書 ID | TASK-CC-V2-001 |
| 作成日 | 2026-04-26 |
| 改訂 | v1.1（2026-04-26: 9 ファイル分割版を 1 ファイルに統合） |
| 基盤要件 | `falco-plugin-dev-kit` 側 `docs/requirements/dev-kit-v2-requirements.md` v5.6（1609 行） |
| 親 Issue | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/1 |
| 進捗管理 | https://github.com/takaosgb3/falco-plugin-claude_code/issues/1 |
| 対象実装 | dev-kit のテンプレート/スキル/エージェントの改善（**claude-code プラグイン実装はこの完了後** Phase 1 scaffold に進む） |
| 既存版 | dev-kit 側 `docs/requirements/TASK_DEFINITIONS.md`（1762 行）の知見を取り込み、claude-code 視点で再整理 |

---

## 目次

- [この文書の目的](#この文書の目的)
- [タスク一覧（Step 順）](#タスク一覧step-順)
- [読み方の推奨](#読み方の推奨)
- [§1 全体概要・改善対象マップ・共通事項](#1-全体概要改善対象マップ共通事項)
  - 1.1 改善の目的
  - 1.2 改善対象マップ（MECE）
  - 1.3 5 ステップ実装計画
  - 1.4 依存関係図
  - 1.5 共通事項（ブランチ / 復元 / 検証 / 主要リポジトリ / Pコードマッピング / 受入テスト / 用語集）
- [§2 Step 1: 基盤改善【P0 Critical】](#2-step-1-基盤改善p0-critical)
- [§3 Step 2: テスト強化・セキュリティ修正【P1 High】](#3-step-2-テスト強化セキュリティ修正p1-high)
- [§4 Step 3: CI/CD・ビルド改善【P1 High】](#4-step-3-cicdビルド改善p1-high)
- [§5 Step 4: ドメイン非依存化【P1 High】](#5-step-4-ドメイン非依存化p1-high)
- [§6 Step 5: ドキュメント・仕組み化【P2 Medium】](#6-step-5-ドキュメント仕組み化p2-medium)
- [§7 クロスリファレンス](#7-クロスリファレンス)
- [§8 リスク・受入テスト・移行戦略](#8-リスク受入テスト移行戦略)
- [改訂履歴](#改訂履歴)

---

## この文書の目的

### 1. 「単独で再開できる」タスク定義
Claude Code がコンテキスト消失・新セッションで再開しても、各タスクの「コンテキスト復元」セクションを読めば必要情報がすべて揃うように設計する。

### 2. claude-code 要件 v3 との接続
本リポジトリの要件 §18.4「PROBLEM_PATTERNS との対応マッピング」(`docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md`) と本タスク定義書の P001〜P021 を双方向参照可能にする。

### 3. MECE な改善カバレッジ
A（テンプレート）/ B（スキル）/ C（新規追加）/ E（非機能）の 4 カテゴリを 5 つの実装ステップに分解。29 タスクで全要件を網羅。

---

## タスク一覧（Step 順）

### Step 1: 基盤改善【P0 Critical】 — §2
| ID | タスク | 要件 ID |
|----|--------|--------|
| T1-1 | parseLine() と parser パッケージの接続 | A1-1 |
| T1-2 | Makefile OS/Arch 自動検出 | A4-1 |
| T1-3 | Level 2 パイプラインテストテンプレート作成 | A7-1 |
| T1-4 | plugin-test スキル更新 | B4 |

### Step 2: テスト強化・セキュリティ修正【P1 High】 — §3
| ID | タスク | 要件 ID |
|----|--------|--------|
| T2-1 | 入力サイズ超過時の挙動修正（truncate 化） | A3-1 |
| T2-2 | JSON パーサーのデフォルト実装 | A2-1 |
| T2-3 | Level 1 E2E パターンテストテンプレート作成 | A7-2 |
| T2-4 | E2E パターン JSON 拡張（benign + edge_cases） | A7-3 |
| T2-5 | PROBLEM_PATTERNS.md への知見追加（P001〜P021） | C2 |

### Step 3: CI/CD・ビルド改善【P1 High】 — §4
| ID | タスク | 要件 ID |
|----|--------|--------|
| T3-1 | Makefile に build-release ターゲット追加 | A4-2 |
| T3-2 | Makefile に E2E テストターゲット追加 | A4-3 |
| T3-3 | CI/CD 3 ワークフロー分離（ci/e2e-test/release） | A5-1 |
| T3-4 | 3 環境 Falco 設定テンプレート作成（local/docker/prod） | A6-1 |
| T3-5 | config.go.tmpl ドメイン非依存化 | A9-1 |
| T3-6 | plugin-build スキル更新 | B5 |

### Step 4: ドメイン非依存化【P1 High】 — §5
| ID | タスク | 要件 ID |
|----|--------|--------|
| T4-1 | PluginEvent のドメイン非依存化 | A1-2 |
| T4-2 | LogEntry のドメイン非依存化（T4-1 とセット） | A2-3 |
| T4-3 | フォーマット自動検出モード追加（auto） | A2-2 |
| T4-4 | plugin-scaffold スキル更新（フィールド対話収集） | B1 |
| T4-5 | plugin-parser スキル更新 | B2 |

### Step 5: ドキュメント・仕組み化【P2 Medium】 — §6
| ID | タスク | 要件 ID |
|----|--------|--------|
| T5-1 | `~/`パス展開ロジック + パストラバーサル防止 | A1-3 / E6 |
| T5-2 | Extract() 冗長 nil チェック削除 | A1-4 |
| T5-3 | URL デコード重複排除 ※ T1-1 に統合済み | A3-2（統合） |
| T5-4 | CLAUDE.md.tmpl 新規作成 | A8-1 |
| T5-5 | CHANGELOG.md.tmpl 新規作成 | A8-2 |
| T5-6 | README.md.tmpl 更新 | A8-3 |
| T5-7 | dev-kit-feedback スキル新規作成 | C1 |
| T5-8 | plugin-dev-workflow エージェント更新 | B6 |
| T5-9 | plugin-rules スキル更新（ドメイン非依存ガイド） | B3 |

**合計: 29 タスク（5 Step）**
※ T5-3 (URL デコード重複排除) は T1-1 に統合済みで実装作業はない。**実装タスク数は実質 28**、ID 整合性のため番号は維持する。

---

## 読み方の推奨

1. 初回は §1 の全体概要 → §7 のクロスリファレンスを先に読み、改善全体像と依存関係を把握する。
2. 実装開始時は対象 Step（§2〜§6）を開き、タスク内 **コンテキスト復元** を先に読む。
3. コンテキストが切れた場合は、Step 共通の「ブランチ・検証」を再読し、タスク内の参照ドキュメント行番号を順に見直すと再開できる。
4. リスクや受入条件は §8 を参照。

---

## §1 全体概要・改善対象マップ・共通事項

### 1.1 改善の目的

`falco-plugin-dev-kit` は Falco プラグインを生成するテンプレート/スキル/エージェントの集合体である。`falco-plugin-openclaw` 開発で発見された問題を吸収し、**生成直後のコードが最小限の手修正で動く**状態へ引き上げる。

> **本リポジトリ（claude-code）との関係**: 本タスク定義書は **dev-kit 側の改善** を対象とする。dev-kit v2 が完了した後に、本リポジトリで `/plugin-scaffold claude-code json` を起動して claude-code プラグインの実装を進める。dev-kit v2 が未完了の状態で claude-code 実装を進めると、テンプレートの問題（macOS で `make build` 失敗、parser 未接続、JSONL パーサー欠如など）に直接ぶつかる。

### 1.2 改善対象マップ（MECE）

```
falco-plugin-dev-kit v2 改善項目
├── A. テンプレートの改善（生成されるコードを直接改善）
│   ├── A1. plugin.go.tmpl
│   │   ├── A1-1: parseLine() と parser の接続 ★Step 1
│   │   ├── A1-2: PluginEvent ドメイン非依存化 ★Step 4
│   │   ├── A1-3: ~/パス展開 + パストラバーサル防止 ★Step 5
│   │   └── A1-4: Extract() 冗長 nil チェック削除 ★Step 5
│   ├── A2. parser.go.tmpl
│   │   ├── A2-1: JSON パーサーのデフォルト実装 ★Step 2
│   │   ├── A2-2: フォーマット自動検出 ★Step 4
│   │   └── A2-3: LogEntry ドメイン非依存化 ★Step 4
│   ├── A3. regex_simple.go.tmpl
│   │   ├── A3-1: 入力サイズ超過時の挙動修正（truncate） ★Step 2
│   │   └── A3-2: URL デコード重複排除 ※T1-1 に統合
│   ├── A4. Makefile.tmpl
│   │   ├── A4-1: OS/Arch 自動検出 ★Step 1
│   │   ├── A4-2: build-release ターゲット ★Step 3
│   │   └── A4-3: E2E ターゲット（pattern/pipeline/e2e） ★Step 3
│   ├── A5. CI/CD ワークフロー
│   │   └── A5-1: 3 ワークフロー分離（ci/e2e-test/release） ★Step 3
│   ├── A6. Falco 設定
│   │   └── A6-1: 3 環境設定（falco.yaml / falco-local.yaml / falco-docker.yaml） ★Step 3
│   ├── A7. テスト系テンプレート（新規）
│   │   ├── A7-1: Level 2 パイプラインテスト ★Step 1
│   │   ├── A7-2: Level 1 E2E パターンテスト ★Step 2
│   │   └── A7-3: benign + edge_cases パターン拡張 ★Step 2
│   ├── A8. ドキュメント系テンプレート（新規）
│   │   ├── A8-1: CLAUDE.md.tmpl ★Step 5
│   │   ├── A8-2: CHANGELOG.md.tmpl ★Step 5
│   │   └── A8-3: README.md.tmpl 更新 ★Step 5
│   └── A9. config.go.tmpl
│       └── A9-1: Config 構造体ドメイン非依存化 ★Step 3
│
├── B. スキル・エージェント定義の改善
│   ├── B1. plugin-scaffold スキル ★Step 4
│   ├── B2. plugin-parser スキル ★Step 4
│   ├── B3. plugin-rules スキル ★Step 5
│   ├── B4. plugin-test スキル ★Step 1
│   ├── B5. plugin-build スキル ★Step 3
│   └── B6. plugin-dev-workflow エージェント ★Step 5
│
├── C. 新規追加
│   ├── C1. dev-kit-feedback スキル（新規） ★Step 5
│   └── C2. PROBLEM_PATTERNS.md 知見追加（P001〜P021） ★Step 2
│
└── E. 非機能要件（横断的）
    ├── E1. 後方互換性
    ├── E2. ドメイン非依存性
    ├── E3. ドキュメント整合性
    ├── E4. 互換性（Go 1.22+, Falco 0.43+, plugin-sdk-go 0.7.4+）
    ├── E5. 性能（100 events/sec, 5s 起動）
    ├── E6. セキュリティ（ReDoS / nil map / 入力長 / パストラバーサル）
    ├── E7. テンプレート変数仕様（既存 13 + v2 新規 2）
    └── E8. 受入テスト（AT-1〜AT-5）
```

凡例: `★Step N` = 該当タスクの実装ステップ。

### 1.3 5 ステップ実装計画

| Step | 期間目安 | タスク数 | 主目的 | 主要要件 ID |
|---|---|---:|---|---|
| 1 | 〜0.5 週間 | 4 | 基盤: 生成直後のコードが go vet/test/build 通過 | A1-1, A4-1, A7-1, B4 |
| 2 | 〜1 週間 | 5 | テスト・セキュリティ強化 | A2-1, A3-1, A7-2, A7-3, C2 |
| 3 | 〜1 週間 | 6 | CI/CD・ビルド・3 環境 Falco 設定 | A4-2/3, A5-1, A6-1, A9-1, B5 |
| 4 | 〜1 週間 | 5 | ドメイン非依存化（HTTP 以外で生成可能に） | A1-2, A2-2, A2-3, B1, B2 |
| 5 | 〜1 週間 | 9 | ドキュメント・仕組み化 | A1-3, A1-4, A8-1〜3, C1, B3, B6 |

> **進め方**: 各 Step を 1 PR としてマージ。Step 内のタスクはセットで実装すべきもの（例: T4-1 と T4-2）以外は 1 コミットずつ。各 Step 完了時に **検証用プラグイン**（test-plugin）を生成して `go vet / go test / make build / make e2e` で確認する。

### 1.4 依存関係図

```
Step 1 (P0 Critical)
  T1-1 (parser 接続) ────────────────────┐
  T1-2 (OS 自動検出)                      │
  T1-3 (Level 2 テスト) ─ depends on T1-1│
  T1-4 (test スキル)                      │
                                          │
Step 2 (P1 High)                          │
  T2-1 (input truncate)                   │
  T2-2 (JSON パーサー) ────────┐          │
  T2-3 (Level 1 テスト)         │          │
  T2-4 (パターン拡張)           │          │
  T2-5 (PROBLEM_PATTERNS)      │          │
                                │          │
Step 3 (P1 High)                │          │
  T3-1 (build-release)          │          │
  T3-2 (E2E target) ─ depends T1-3, T2-3   │
  T3-3 (CI 分離)                │          │
  T3-4 (3 環境設定)              │          │
  T3-5 (config 非依存化)         │          │
  T3-6 (build スキル)            │          │
                                │          │
Step 4 (P1 High)                │          │
  T4-1 (PluginEvent 非依存化) ─ T1-1 への再改修
  T4-2 (LogEntry 非依存化) ─ T4-1 とセット
  T4-3 (auto モード) ─ depends on T2-2 ───┘
  T4-4 (scaffold スキル) ─ depends on T4-1/T4-2
  T4-5 (parser スキル)

Step 5 (P2 Medium)
  T5-1 (パス展開)
  T5-2 (Extract 冗長削除)
  T5-3 (URL デコード) ※統合済
  T5-4〜T5-6 (ドキュメント)
  T5-7 (feedback スキル)
  T5-8 (workflow エージェント) ─ depends on 全 Step
  T5-9 (rules スキル)
```

> **重要なセット実装**:
> - T4-1 と T4-2 は同時に行う（PluginEvent と LogEntry のフィールド整合性のため）
> - T1-3 と T2-3 完了後でないと T3-2（make e2e）が成立しない
> - 図中「T4-1 → T1-1 への再改修」は **T4-1 が T1-1 で書いた parseLine() 等の HTTP 直接マッピングをテンプレート展開化する**意味（T1-1 の実装を Step 4 で書き換える）。新規依存ではなく改修方向の矢印。

### 1.5 共通事項

#### 1.5.1 ブランチ戦略
```
main
  └── feat/v2-step1  (T1-1〜T1-4)
  └── feat/v2-step2  (T2-1〜T2-5)
  └── feat/v2-step3  (T3-1〜T3-6)
  └── feat/v2-step4  (T4-1〜T4-5)
  └── feat/v2-step5  (T5-1〜T5-9)
```
各 Step を PR にしてマージ。Step 内のタスクは原則 1 コミット 1 タスク。セット実装は 1 コミットに同梱。

#### 1.5.2 コンテキスト復元の標準手順
新しい Claude Code セッションでタスクを再開する時:
1. 本ドキュメント先頭の目次と §1 を読む
2. 対象 Step の章（§2〜§6）を開く
3. 対象タスク内 **コンテキスト復元** セクションの参照ドキュメント・行番号をすべて開く
4. dev-kit 側の現在のテンプレートと openclaw の参照実装を比較する
5. 実装に着手する

#### 1.5.3 検証の標準パターン
全 Step に共通する検証コマンド:
```bash
# 環境変数 STEP に Step 番号（1-5）を入れる。例: export STEP=1
# テスト用プラグインを毎回 clean state で新規生成（前 Step の test-plugin は削除して再作成）
TEST_DIR=/tmp/dev-kit-v2-step$STEP-test-plugin
rm -rf "$TEST_DIR" && mkdir -p "$TEST_DIR" && cd "$TEST_DIR"
# (scaffold スキル経由 or 手動でテンプレート展開)

go vet ./...                # E1: 静的解析
go test ./... -v -race      # E1: ユニット + Level 2
make build                  # macOS は .dylib, Linux は .so
make e2e                    # Level 1 + Level 2（Step 3 以降）
make build-release          # サイズ最適化（Step 3 以降）
file lib*.so                # ELF 検証（Linux）
```

> **生成方針**: 各 Step の検証では **必ず新規 clean state** で test-plugin を生成する。前 Step の生成物を再利用しない。理由: 前 Step の状態が残っていると、新規変更が反映されたかの判定が曖昧になる。

#### 1.5.4 主要リポジトリの場所

| リポジトリ | パス / URL | 用途 |
|---|---|---|
| dev-kit（編集対象） | `https://github.com/takaosgb3/falco-plugin-dev-kit`（local clone は dev-kit/`feat/v2-stepN` ブランチ） | 全タスクの直接編集対象 |
| dev-kit v2 要件書 | dev-kit `docs/requirements/dev-kit-v2-requirements.md`（v5.6, 1609 行） | 各タスクの requirements 参照 |
| dev-kit 既存タスク定義書 | dev-kit `docs/requirements/TASK_DEFINITIONS.md`（1762 行） | 既存版（参考） |
| openclaw（参照実装） | `/Users/takaos/lab/falco-plugin-openclaw/` | 全タスクの参照実装 |
| nginx-plugin（回帰確認・任意） | `/Users/takaos/lab/falco-plugin-nginx/`（存在する場合） | 既存プラグイン回帰確認（ET-7）。HTTP ドメインの後方互換性検証 |
| claude-code（本リポ） | `/Users/takaos/lab/falco-plugin-claude_code/` | dev-kit v2 完了後の利用先 |
| claude-code 要件 v3 | 本リポ `docs/claude_code_falco_plugin_requirements_2026-04-26_v3.md` | §18.4 で P001〜P021 と要件項目の対応マッピング |
| PROBLEM_PATTERNS | dev-kit `PROBLEM_PATTERNS.md` | A コード（A303-A334）+ T2-5 で追加する P001〜P021 |

#### 1.5.5 P コード ↔ 要件 ↔ タスクの参照関係（クイック参照）

完全な逆引きは §7.2 を参照。本表は概観用。

> **重要な注意**: P コード番号は dev-kit `PROBLEM_PATTERNS.md`（T2-5 で集約）が原典。
> claude-code 要件 v3 §18.4 は同じ番号体系を使うが、**P006 / P007 / P009 / P012 / P015 / P016 / P020 / P021 で意味解釈が異なる**（要件 v3 側の独自解釈）。
> 本表「要件 v3 §18.4 マッピング」列は **dev-kit 解釈の Pコードに対する claude-code 側要件節への参照** を示し、要件 v3 §18.4 表内の同番号行に書かれた要件節を転記している。両者の意味整合は T2-5 完了後に再見直しする（要件 v3 側を改訂するか、dev-kit Pコードに新番号を振るかは別途検討）。

| P コード | 概要 | 関連タスク | claude-code 要件 v3 §18.4 マッピング |
|---|---|---|---|
| P001 | macOS バイナリ Linux 配布 | T1-2, T3-1, T3-3 | §7.4, §21.1, §21.2 B-003/B-004, AC-013 |
| P002 | -buildmode=c-shared 不指定 | T1-2, T3-1 | §21.2 B-001, §27.2 |
| P003 | rule に source 不指定 | T5-9 | §13 R-001, §18.3, §29 |
| P004 | GOB nil map panic | T1-1, T1-3, T2-2, T4-1 | §14.1 P-005 |
| P005 | rule で evt.type 使用 | T5-9 | §13 R-002, §18.3, §29 |
| P006 | URL エンコードパターン | T2-2 | （rules 側） |
| P007 | rules_files 個別パス | T3-4 | §18.3 |
| P008 | load_plugins 欠落 | T3-4 | §18.3 |
| P009 | レート制限でアラート抑制 | T3-4 | §18.3 |
| P010 | Fields/Extract 不一致 | T1-1, T4-1 | §6.3 FP-008/FP-009, §23.2 |
| P011 | YAML コメント位置 | T5-9 | §13 R-010 |
| P012 | headers 参照は小文字 | T5-9 | §13 R-* |
| P013 | ビルド環境不一致 | T1-2, T3-1, T3-6 | §21.2 B-002/B-005 |
| P014 | ファイル SeekEnd 必須 | T1-1, T1-3 | §6.2 ES-004, §14.1 P-006 |
| P015 | クロスルール干渉 | T5-9 | §13 R-005 |
| P016 | URL エンコード多段 | T2-1, T2-2 | （rules 側） |
| P017 | macOS Falco outputs 拒否 | T3-4, T3-6 | §7.2 MAC-002, §18.1 |
| P018 | macOS -U フラグ必須 | T3-4, T3-6 | §7.2 MAC-003, §18.3 |
| P019 | Falco 1 イベント 1 ルール | T5-9 | §13 R-005/R-006 |
| P020 | 検出 truncate vs 全文返却 | T2-1 | §14.2 D-004 |
| P021 | fsnotify タイミング | T1-3 | §6.3 FP-006 |

#### 1.5.6 受入テスト（要件 §E8）

| TC ID | テストケース | 入力 | 期待結果 | 実施タイミング |
|---|---|---|---|---|
| AT-1 | HTTP プラグイン生成 | format=combined, HTTP 標準フィールド | go vet + go test + make build 成功 | Step 4 完了時 |
| AT-2 | AI プラグイン生成 | format=json, type/tool/args/session_id | 同上 | Step 4 完了時 |
| AT-3 | IoT プラグイン生成 | format=custom, device_id/sensor_type/value(string) | 同上 | Step 4 完了時 |
| AT-4 | macOS ビルド | AT-1 を macOS arm64 で実行 | make build で .dylib 生成成功 | Step 1 / Step 4 |
| AT-5 | E2E テスト | AT-1 で make e2e | Level 1 + Level 2 全通過 | Step 3 / Step 4 |

> **拡張受入テスト**: 本書独自の ET-1〜ET-6（flaky 計測、ET-6 で claude-code scaffold 動作確認）は §8.2.2 を参照。Step 完了タイミングは §8.2.3 にまとめる。

#### 1.5.7 用語集（抜粋）

| 用語 | 説明 |
|---|---|
| dev-kit | falco-plugin-dev-kit。プラグイン生成ツールキット |
| テンプレート | `.claude/templates/plugin/*.tmpl`。`${VAR}` 形式の文字列置換でコードを生成 |
| スキル | `.claude/skills/*/SKILL.md`。Claude Code の `/<skill>` で起動する手順書 |
| エージェント | `.claude/agents/*.md`。複数スキルを Phase 順に実行する自律エージェント |
| WF-Phase | プラグイン開発時のワークフロー段階（0〜6） |
| 実装ステップ Step | 本タスク定義書の改善実装の順序（1〜5） |
| Level 1/2/3 | テストレベル（1=パターン, 2=パイプライン, 3=Falco 統合） |
| Pコード | プラグイン共通の問題パターン ID（P001〜P021）。dev-kit `PROBLEM_PATTERNS.md` |
| GOB | Go 標準の `encoding/gob` バイナリ。プラグイン↔Falco 間 |
| fsnotify | ファイル変更監視ライブラリ |

---

## §2 Step 1: 基盤改善【P0 Critical】

### 2.1 Step 1 の目標

生成直後のプラグインが **`go vet` + `go test` + `make build` を通過する** 状態にする。これは「最低限動くテンプレート」のラインを引くため最優先。

| 指標 | 現状 | Step 1 完了時 |
|---|---|---|
| parser 接続 | TODO のまま | parser.New() / Parse() を呼ぶ |
| macOS で `make build` | クロスコンパイルエラー | `.dylib` 生成成功 |
| Level 2 テスト | 存在せず | テンプレート同梱、生成時点で実行可 |
| plugin-test スキル | ユニットテストのみ | 3 層 E2E アーキテクチャを記載 |

### 2.2 Step 1 共通: ブランチ・検証

#### ブランチ
- `feat/v2-step1`
- 着地先: dev-kit `main`

#### 着手時の必読
1. dev-kit v2 要件書 § 7「Step 1」（L1451-L1465）
2. 本章冒頭（§2.1〜§2.2）
3. dev-kit `PROBLEM_PATTERNS.md` の A コード概観

#### 完了判定（PR マージ前に確認）
| # | 確認項目 | 方法 |
|---|---------|------|
| 1 | テスト用プラグインが生成できる | scaffold スキル or 手動 |
| 2 | `go vet ./...` パス | テスト用プラグインで実行 |
| 3 | `go test ./... -v -race` パス（Level 2 含む） | 同上 |
| 4 | macOS で `make build` 成功 | macOS arm64 / amd64 両方 |
| 5 | Linux で `make build` 成功 | CI または local Linux |
| 6 | 生成バイナリの拡張子が OS に応じて切り替わる | `file lib*.so` / `file lib*.dylib` |

### 2.3 T1-1: parseLine() と parser パッケージの接続

| 項目 | 内容 |
|------|------|
| 要件 ID | A1-1 |
| 優先度 | **P0 Critical** |
| 先行タスク | なし（最初に着手） |
| 後続タスク | T1-3（Level 2 テストは parser 接続が前提）、T4-1（PluginEvent 非依存化で再改修） |
| Pコード関連 | P004（nil map）、P010（Fields/Extract 一致）、P014（SeekEnd） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit `docs/requirements/dev-kit-v2-requirements.md` | L93-L209（A1-1 セクション） | Before/After コード例と設計方針 |
| dev-kit `.claude/templates/plugin/plugin.go.tmpl` | L1-L11（テンプレート変数）、L48-L51（PluginConfig）、L275-L302（parseLine）、L307-L395（Fields/Extract） | 現在のテンプレート構造 |
| dev-kit `.claude/templates/plugin/parser.go.tmpl` | L1-L30（パッケージ）、L60-L95（Parser 構造体と New） | parser パッケージ API |
| dev-kit `.claude/templates/plugin/config.go.tmpl` | 全体（9 行） | parser.Config 構造体 |
| openclaw `cmd/plugin-sdk/plugin.go` | L1-L50（import/構造体）、L100-L150（Init）、L200 付近（parseLine） | 参照実装 |
| 本リポ要件 v3 §14.1 P-005 | — | nil map 防止 |

#### 変更内容（編集対象 = `.claude/templates/plugin/plugin.go.tmpl`）

1. **import に parser パッケージを追加**（L5 付近）
   ```go
   "github.com/${AUTHOR}/${PLUGIN_NAME}/pkg/parser"
   ```
2. **MyPlugin 構造体に parser フィールドを追加**（L79 付近）
   ```go
   parser *parser.Parser
   ```
3. **Init() 内で parser を初期化**（L134 付近）
   ```go
   p.parser = parser.New(parser.Config{
       LogFormat:        "${LOG_FORMAT}",
       SecurityPatterns: true,
   })
   ```
4. **Open() で parser を readLoop に渡す**（L213 付近）
   ```go
   go instance.readLoop(p.parser)
   ```
5. **readLoop / readNewLines のシグネチャに parser 追加**（L219, L239 付近）
6. **parseLine() を実装**（L275-L302） — TODO を削除して parser.Parse(line) を呼ぶ
   - Headers は `entry.Headers` を直接参照（parser 側で `make()` 済みである前提だが、defensive に nil チェック）
   - HTTP 固有フィールドは Step 1 では直接マッピング（Step 4 の T4-1 でテンプレート展開化）

#### 検証
```bash
# テスト用プラグイン生成
mkdir -p /tmp/test-plugin && cd /tmp/test-plugin
# (テンプレート展開：scaffold スキル経由 or sed)

go vet ./...                # PASS 必須
go test ./pkg/parser/ -v    # 既存パーサーテストが parser 接続後も通る
```

#### 完了基準
- [ ] テスト用プラグインで `go vet ./...` がエラー 0
- [ ] `parser.Parse()` が呼ばれることをテストで確認
- [ ] Headers が nil でないことをテストで確認（P004）
- [ ] PR コミット粒度: 1 コミット

### 2.4 T1-2: Makefile OS/Arch 自動検出

| 項目 | 内容 |
|------|------|
| 要件 ID | A4-1 |
| 優先度 | **P0 Critical** |
| 先行タスク | なし |
| 後続タスク | T3-1（build-release）、T3-2（E2E ターゲット）、T3-6（build スキル更新） |
| Pコード関連 | P001（macOS バイナリリリース）、P002（c-shared）、P013（ビルド環境）、P017（macOS outputs）、P018（-U フラグ） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L645-L695（A4-1） | Before/After コード |
| dev-kit `.claude/templates/plugin/Makefile.tmpl` | L1-L5（変数定義） | 現状の Linux 固定 |
| openclaw `Makefile` | 冒頭 OS 自動検出ブロック | 参照実装 |
| 本リポ要件 v3 §27.2 | ビルドフラグ表 | macOS/Linux/GLIBC の最低版 |

#### 変更内容（編集対象 = `.claude/templates/plugin/Makefile.tmpl`）

1. **OS/Arch 自動検出ブロックを冒頭に追加**:
   ```makefile
   UNAME_S := $(shell uname -s)
   UNAME_M := $(shell uname -m)
   ifeq ($(UNAME_S),Darwin)
     ifeq ($(UNAME_M),arm64)
       BINARY := lib$(PLUGIN_NAME)-plugin-darwin-arm64.dylib
       GO_ENV := CGO_ENABLED=1 GOOS=darwin GOARCH=arm64
     else
       BINARY := lib$(PLUGIN_NAME)-plugin-darwin-amd64.dylib
       GO_ENV := CGO_ENABLED=1 GOOS=darwin GOARCH=amd64
     endif
   else
     BINARY := lib$(PLUGIN_NAME)-plugin-linux-amd64.so
     GO_ENV := CGO_ENABLED=1 GOOS=linux GOARCH=amd64
   endif
   ```
2. **GO_BUILD_FLAGS と GO_RELEASE_FLAGS の分離**:
   ```makefile
   GO_BUILD_FLAGS := -buildmode=c-shared
   GO_RELEASE_FLAGS := -buildmode=c-shared -trimpath -ldflags="-s -w"
   ```

#### 検証
```bash
# macOS arm64 で
make build && file lib*.dylib | grep -q 'Mach-O.*arm64'
# Linux amd64 で
make build && file lib*.so | grep -q 'ELF.*64-bit'
```

#### 完了基準
- [ ] macOS arm64 / amd64 / Linux amd64 のすべてで `make build` 成功
- [ ] BINARY 名が OS/Arch を正しく反映
- [ ] PR コミット粒度: 1 コミット

### 2.5 T1-3: Level 2 パイプラインテストテンプレート作成

| 項目 | 内容 |
|------|------|
| 要件 ID | A7-1 |
| 優先度 | **P0 Critical** |
| 先行タスク | T1-1（parser 接続が必要） |
| 後続タスク | T3-2（make e2e で起動）、T1-4（test スキルでドキュメント化） |
| Pコード関連 | P004, P010, P014, P021（fsnotify タイミング） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L884-L940（A7-1 セクション） | TC 一覧・ヘルパー関数仕様 |
| openclaw `cmd/plugin-sdk/plugin_test.go` | 全体（37 テスト関数、6 ヘルパー） | 参照実装 |
| dev-kit `.claude/templates/plugin/plugin.go.tmpl` | NextBatch / Open / Init | テスト対象の API 把握 |
| 本リポ要件 v3 §6.3 FP-* | — | プラグイン側要件 |

#### 変更内容（新規作成 = `.claude/templates/plugin/plugin_test.go.tmpl`）

1. **テストカテゴリ**: ライフサイクル / 取り込み / 性能 / エラー耐性
2. **必須 TC**（要件 v2 §A7-1 で具体的に列挙されたもの）:
   - TC-1-01〜08: Init / Open / Close / SeekEnd / バッファ境界
   - TC-2-01, TC-2-04, TC-2-05, TC-2-06: ログ取り込み / 複数ファイル / GOB ラウンドトリップ / Headers 非 nil（TC-2-02/03 は要件で未指定）
   - TC-5-01, TC-5-02: 100 events/sec, バッファオーバーフロー
   - TC-6-01, TC-6-04: 不正 JSON 設定 / ファイル削除時
3. **ヘルパー関数**:
   ```go
   initPlugin / openAndCleanup / writeToLog / waitForEvent / gobEncode / gobDecode
   ```
4. **fsnotify タイミング対策（P021）**: writeToLog で `time.Sleep(50*time.Millisecond)` のような sleep をコメント付きで入れる

#### 検証
- テスト用プラグインで `go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -count=1 -timeout 120s` パス
- `-race` でデータ競合検出されないこと

#### 負荷生成サンプル（TC-5-01: 100 events/sec, TC-5-02: バッファオーバーフロー）

```go
// TC-5-01: 100 events/sec を 10 秒間継続して投入
func TestPipeline_TC5_01_Throughput(t *testing.T) {
    plugin := initPlugin(t, []string{eventsPath})
    inst := openAndCleanup(t, plugin)

    target := 1000
    interval := 10 * time.Millisecond  // 100 events/sec
    start := time.Now()
    for i := 0; i < target; i++ {
        writeToLog(t, eventsPath, sampleHookEventJSON(i))
        time.Sleep(interval)
    }
    elapsed := time.Since(start)

    events := drainEvents(t, inst, target, 30*time.Second)
    require.GreaterOrEqual(t, len(events), target*95/100,  // 95% 以上の取得
        "100 events/sec to throughput drop=%d", target-len(events))
    require.LessOrEqual(t, elapsed, 12*time.Second)  // 投入時間自体が想定内
}

// TC-5-02: バッファサイズより多く一気に投入し drop counter が上がること
func TestPipeline_TC5_02_BufferOverflow(t *testing.T) {
    plugin := initPlugin(t, []string{eventsPath})  // EventBufferSize=100 等の小さい値
    inst := openAndCleanup(t, plugin)

    // 1000 行を一気に書く（plugin が読み出す前にバッファ溢れ）
    burst := 1000
    for i := 0; i < burst; i++ {
        writeToLog(t, eventsPath, sampleHookEventJSON(i))
    }

    // drop counter が増えていること（hang していないこと）
    require.Eventually(t, func() bool {
        return plugin.DroppedCounter() > 0
    }, 10*time.Second, 100*time.Millisecond)
}
```

#### 完了基準
- [ ] TC-1-01〜TC-6-04 がテンプレートに含まれる
- [ ] 6 ヘルパー関数が定義
- [ ] 100 events/sec の TC-5-01 が安定パス（10 連続実行で flaky 0）
- [ ] TC-5-02 で drop counter 増加を確認、hang しない
- [ ] PR コミット粒度: 1 コミット

### 2.6 T1-4: plugin-test スキル更新

| 項目 | 内容 |
|------|------|
| 要件 ID | B4 |
| 優先度 | **P0 Critical** |
| 先行タスク | T1-3 |
| 後続タスク | T3-6（build スキル）、T5-8（workflow エージェント） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1217-L1247（B4 セクション） | スキル更新指示 |
| dev-kit `.claude/skills/plugin-test/SKILL.md` | 全体 | 現状のスキル本文 |
| 本リポ要件 v3 §20 | — | 3 層 E2E テストアーキテクチャ |

#### 変更内容（編集対象 = `.claude/skills/plugin-test/SKILL.md`）

1. **3 層 E2E テストアーキテクチャ説明を追加**:
   - Level 1: パターンカバレッジ（Falco 不要、`make e2e-pattern`）
   - Level 2: プラグインパイプライン（Falco 不要、CGO 必要、`make e2e-pipeline`）
   - Level 3: Falco 統合（Falco 必要、`make e2e-ci` / `make e2e-native`）
2. **Level 2 テスト生成手順**を追加（T1-3 のテンプレートを利用）
3. **Level 1 テスト生成手順**は T2-3 完了時に追記（プレースホルダーを残す）
4. **成功基準 SC-030〜033** を更新:
   - SC-031: Level 2 テスト全 TC 通過
   - SC-032: 100 events/sec 達成

#### 検証
- スキル本文に 3 層の説明が含まれる
- `make e2e-pipeline` の実行手順が記述
- Level 1 / Level 3 のプレースホルダーがある

#### 完了基準
- [ ] SKILL.md に 3 層 E2E アーキテクチャが記述
- [ ] SC-030〜033 が更新されている
- [ ] PR コミット粒度: 1 コミット

### 2.7 Step 1 完了の最終確認

| 確認項目 | 検証コマンド / 方法 | 期待結果 |
|---|---|---|
| テスト用プラグインの生成 | scaffold スキル | 生成成功 |
| 生成プラグインの go vet | `go vet ./...` | エラー 0 |
| 生成プラグインの go test | `go test ./... -v -race` | 全 PASS |
| macOS arm64 ビルド | `make build` | `.dylib` 生成 |
| Linux amd64 ビルド | `make build` | `.so` 生成 |
| Level 2 テスト | `go test ./cmd/plugin-sdk/ -run TestPipeline` | 全 PASS |
| plugin-test SKILL.md | 目視 | 3 層 E2E 説明あり |
| PR | dev-kit `feat/v2-step1` | 4 コミット（T1-1〜T1-4） |

---

## §3 Step 2: テスト強化・セキュリティ修正【P1 High】

### 3.1 Step 2 の目標

セキュリティ検出の盲点（10KB 超で検出スキップ）を修正し、JSONL ログのデフォルト対応と E2E パターンテストの基盤を整える。openclaw 開発で得た知見を `PROBLEM_PATTERNS.md` に集約する。

### 3.2 Step 2 共通: ブランチ・検証

- ブランチ: `feat/v2-step2`（dev-kit `main` から派生）
- 着手前必読: dev-kit v2 要件書 § 7「Step 2」（L1466-L1480）
- 完了判定:
  - 10KB 超入力でセキュリティ検出が動く
  - `make e2e-pattern` パス
  - `PROBLEM_PATTERNS.md` に P001〜P021 が記載されている

### 3.3 T2-1: 入力サイズ超過時の挙動修正（truncate 化）

| 項目 | 内容 |
|------|------|
| 要件 ID | A3-1 |
| 優先度 | **P1 High** |
| 先行タスク | なし |
| 後続タスク | T2-3（パターンテストで検証） |
| Pコード関連 | P016（URL エンコード多段）、P020（検出 truncate vs 全文返却） |
| セキュリティ関連 | E6（10KB 超は切り詰め、スキップしない） |
| 注 | A3-2（URL デコード重複排除）は T1-1 で `detectSecurityPatterns()` 一箇所に集約済み。本タスクと衝突しない。 |

#### 背景
現在の実装は `if len(input) > maxInputLength { return "", false }` で **検出をスキップ**する。攻撃者は意図的に大きなペイロードを送ることでセキュリティ検出を回避できる。**先頭 10KB に対して検出を続行**するのが正解。

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L587-L626（A3-1 セクション） | Before/After |
| dev-kit `.claude/templates/plugin/regex_simple.go.tmpl` | L27-L29 | 現状のスキップロジック |
| openclaw `pkg/parser/regex_simple.go` | DetectSecurityThreat() | 参照実装 |
| 本リポ要件 v3 §14.2 D-004 | — | bounded input 上限 |

#### 変更内容（編集対象 = `.claude/templates/plugin/regex_simple.go.tmpl`）
```go
// Before
if len(input) > d.maxInputLength {
    return "", false   // ← 攻撃者が悪用可能
}

// After
if len(input) > d.maxInputLength {
    input = input[:d.maxInputLength]   // ← 切り詰めて検出続行
}
```

> **注**: `Extract()` は切り詰めずに全文を Falco に返却する（P020 に該当する意図的な設計）。本タスクは検出側のみを変更する。

#### 検証
- テスト追加: 11KB の悪意ある入力（先頭 10KB に SQLi パターン）→ 検出される
- テスト追加: 11KB の悪意ある入力（11KB 目に SQLi パターン）→ 検出されない（既知の制約として明示）

#### 完了基準
- [ ] テンプレートが truncate 方式に変更
- [ ] テストケースが追加
- [ ] PR コミット粒度: 1 コミット

### 3.4 T2-2: JSON パーサーのデフォルト実装

| 項目 | 内容 |
|------|------|
| 要件 ID | A2-1 |
| 優先度 | **P1 High** |
| 先行タスク | T1-1 |
| 後続タスク | T4-3（auto モード）、T4-2（LogEntry 非依存化） |
| Pコード関連 | P004（Headers 初期化） |

#### 背景
現在の `parseJSON()` は `"JSON format not yet implemented"` エラーを返すだけ。多くの現代アプリ（Claude Code 含む）は JSONL を出すため、デフォルト実装が必須。

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L368-L435（A2-1 セクション） | 実装例とタイムスタンプ多形式対応 |
| dev-kit `.claude/templates/plugin/parser.go.tmpl` | L198-L201（parseJSON 部分） | 現状の TODO 実装 |
| openclaw `pkg/parser/parser.go` | parseJSON() | 参照実装 |
| 本リポ要件 v3 §10.1 | timestamp ポリシー | RFC3339 多形式対応 |

#### 変更内容（編集対象 = `.claude/templates/plugin/parser.go.tmpl`）

1. **parseJSON() を実装**:
   ```go
   func (p *Parser) parseJSON(line string) (*LogEntry, error) {
       var raw map[string]interface{}
       if err := json.Unmarshal([]byte(line), &raw); err != nil {
           return nil, fmt.Errorf("invalid JSON: %w", err)
       }
       entry := &LogEntry{ Headers: make(map[string]string) }   // P004
       if v, ok := raw["timestamp"].(string); ok {
           entry.Timestamp = p.parseTimestamp(v)
       }
       if v, ok := raw["level"].(string); ok {
           entry.Headers["level"] = v
       }
       if v, ok := raw["message"].(string); ok {
           entry.Headers["message"] = v
       }
       ${DOMAIN_FIELDS_PARSE_JSON}    // ← T4-2 で展開
       return entry, nil
   }
   ```
2. **parseTimestamp() を追加** — RFC3339, RFC3339Nano, `2006-01-02T15:04:05`, `2006-01-02 15:04:05`, `p.timeLayout` を順に試行

#### 検証
- テスト追加: `parseJSON({"timestamp":"2026-04-26T12:34:56Z","level":"info","message":"hi"})` → 正常 parse
- テスト追加: 不正 JSON → エラー
- テスト追加: timestamp が RFC3339 と RFC3339Nano の両方で parse される

#### 完了基準
- [ ] parseJSON が JSONL を parse 成功
- [ ] Headers が `make()` で初期化（P004）
- [ ] `${DOMAIN_FIELDS_PARSE_JSON}` プレースホルダーが入っている（T4-2 で展開予定）
- [ ] PR コミット粒度: 1 コミット

### 3.5 T2-3: Level 1 E2E パターンテストテンプレート作成

| 項目 | 内容 |
|------|------|
| 要件 ID | A7-2 |
| 優先度 | **P1 High** |
| 先行タスク | T1-1, T2-2 |
| 後続タスク | T3-2（make e2e で起動）、T2-4（パターン拡張） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L944-L961（A7-2） | テストフレームワーク仕様 |
| openclaw `test/e2e/e2e_pattern_test.go` | 全体（9 テストケース） | 参照実装 |
| dev-kit `.claude/templates/plugin/e2e_pattern.json.tmpl` | 既存 5 カテゴリ x 4 パターン | 入力形式 |

#### 変更内容（新規作成 = `.claude/templates/plugin/e2e_pattern_test.go.tmpl`）

1. **TC-3-01〜05**:
   - 全攻撃カテゴリの True Positive
   - True Negative
   - 10KB 入力境界
   - 大文字小文字非依存
2. **動的読み込み**: `test/e2e/patterns/categories/*.json` を走査して全パターンを実行
3. **`format` フィールド対応**: json/plaintext/combined を切り替えて parser を呼ぶ

#### 検証
```bash
go test ./test/e2e/ -v -race -run TestPattern -count=1
```

#### 完了基準
- [ ] テンプレートが正しく展開され生成プラグインで `make e2e-pattern` パス
- [ ] benign カテゴリへの拡張余地（T2-4 用）を残す
- [ ] PR コミット粒度: 1 コミット

### 3.6 T2-4: E2E パターン JSON 拡張（benign + edge_cases）

| 項目 | 内容 |
|------|------|
| 要件 ID | A7-3 |
| 優先度 | **P1 High** |
| 先行タスク | T2-3 |
| 後続タスク | なし |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L964-L991（A7-3） | 追加カテゴリと JSON スキーマ拡張 |
| dev-kit `.claude/templates/plugin/e2e_pattern.json.tmpl` | 全体 | 既存テンプレート構造 |
| openclaw `test/e2e/patterns/categories/*.json` | benign + edge_cases | 参照実装（既存ファイル名から推定） |

#### 変更内容（編集対象 = `.claude/templates/plugin/e2e_pattern.json.tmpl`）

1. **`benign` カテゴリ追加**: 5+ 件の正常リクエスト
2. **`edge_cases` カテゴリ追加**: 10239/10240/10241 バイト、空文字、空白のみ
3. **JSON スキーマ拡張**: `format`, `expected_threat`, `note` フィールドを追加

#### 検証
```bash
go test ./test/e2e/ -v -run TestPattern_Benign
go test ./test/e2e/ -v -run TestPattern_EdgeCases
```

#### 完了基準
- [ ] benign 5+, edge_cases 5+ パターンが追加
- [ ] スキーマ拡張がドキュメント化（テンプレートヘッダコメント）
- [ ] T2-3 のテストフレームワークが新パターンを自動取り込む
- [ ] PR コミット粒度: 1 コミット

### 3.7 T2-5: PROBLEM_PATTERNS.md への知見追加（P001〜P021）

| 項目 | 内容 |
|------|------|
| 要件 ID | C2 |
| 優先度 | **P1 High** |
| 先行タスク | なし |
| 後続タスク | T5-9（rules スキルから参照）, T5-7（dev-kit-feedback の比較基盤） |

#### 背景
現状 `PROBLEM_PATTERNS.md` には A コード（A303-A334, nginx-proxy 由来）のみ。P001〜P016 はスキル定義に散在、P017〜P021 は openclaw 開発で発見済みだが未集約。**スキル横断の参照点**として集約する。

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1294-L1336（C2 セクション） | P001〜P021 一覧と発見経緯 |
| dev-kit `PROBLEM_PATTERNS.md` | 全体 | 現状の A コード構造 |
| dev-kit 各 `.claude/skills/*/SKILL.md` | P*-* インライン記載部 | 集約元 |
| 本リポ要件 v3 §18.4 | マッピング表 | 双方向参照 |

#### 変更内容（編集対象 = `dev-kit/PROBLEM_PATTERNS.md`）

1. **新セクション「P コード: プラグイン共通パターン」を追加**
2. **P001〜P016 を集約**（既存スキルから移植、要件書 L1311-L1326 を参照）
3. **P017〜P021 を新規追加**（要件書 L1329-L1336）
   - P017: macOS Falco outputs 拒否
   - P018: macOS -U フラグ必須
   - P019: Falco 1イベント1ルール制約
   - P020: 検出 truncate vs 全文返却
   - P021: fsnotify タイミング
4. **各 P コードに以下を含める**:
   - 概要 / 影響 / 検出方法 / 対策 / 関連スキル / 関連タスク（本タスク定義書の T*-* ID）

#### 検証
- 各スキル `SKILL.md` から `PROBLEM_PATTERNS.md` への参照リンクを追加（または既存を確認）
- 本リポ要件 v3 §18.4 のマッピング表と内容が整合

#### 完了基準
- [ ] P001〜P021 全 21 件が集約
- [ ] 各項目に「関連タスク」欄が含まれる
- [ ] 既存 A コードと共存（順序: P コード → A コード）
- [ ] **要件 v3 §18.4 との Pコード意味食い違い 8 件（P006/P007/P009/P012/P015/P016/P020/P021）の整合プランを別 Issue（claude-code 側）で起票**。整合は (a) 要件 v3 §18.4 を改訂、(b) dev-kit Pコードに新番号を振り直す、(c) 並行運用 + マッピング表追加 の 3 案を提示し、レビュー後決定する。Issue 雛形:

```
Title: P コード番号体系の整合（dev-kit ↔ claude-code 要件 v3 §18.4）

## 背景
T2-5 で dev-kit `PROBLEM_PATTERNS.md` に P001〜P021 を集約した結果、
claude-code 要件 v3 §18.4 で同番号が異なる意味で使われている 8 件が判明:

| Pコード | dev-kit 解釈 | 要件 v3 §18.4 解釈 |
|---|---|---|
| P006 | URL エンコードパターン | counters / observability 不足 |
| P007 | rules_files 個別パス | NextBatch busy loop |
| P009 | レート制限でアラート抑制 | rules_files 個別指定不足 |
| P012 | headers 参照は小文字 | rules ファイル分割の重複定義 |
| P015 | クロスルール干渉 | log rotation でファイル追従不可 |
| P016 | URL エンコード多段 | log line truncation / 大型イベントで OOM |
| P020 | 検出 truncate vs 全文返却 | required_plugin_versions 未指定 |
| P021 | fsnotify タイミング | release artifact 整合性チェック欠如 |

## 整合プラン（3 案）
- (a) 要件 v3 §18.4 を dev-kit 解釈に揃えて改訂
- (b) dev-kit Pコードに新番号を振り直し、要件 v3 §18.4 を残す
- (c) 並行運用：要件 v3 §18.4 を「claude-code 専用 Pコード（CC-Pxxx）」へ改名、本書で
      dev-kit Pコード ↔ CC-Pコードのマッピング表を維持

## 推奨
dev-kit 解釈が原典（先行）であり PROBLEM_PATTERNS.md として既に別リポで運用されているため、
**(a) 要件 v3 §18.4 改訂** を推奨。レビュー後決定する。
```

- [ ] PR コミット粒度: 1 コミット

### 3.8 Step 2 完了の最終確認

| 確認項目 | 検証コマンド / 方法 | 期待結果 |
|---|---|---|
| 入力 truncate | テスト用プラグインで 11KB 入力テスト | 検出される |
| JSON パース | `parseJSON({"timestamp":"...","message":"..."})` | LogEntry 取得 |
| Level 1 テスト | `make e2e-pattern` | 全 PASS |
| benign / edge_cases | パターン JSON カテゴリ存在 | 5+ 件ずつ |
| PROBLEM_PATTERNS | P001〜P021 集約 | 全 21 件 |
| PR | dev-kit `feat/v2-step2` | 5 コミット（T2-1〜T2-5） |

---

## §4 Step 3: CI/CD・ビルド改善【P1 High】

### 4.1 Step 3 の目標

ビルド最適化（リリース）と E2E ターゲットを Makefile に整え、CI を 3 ワークフローに分離。Falco 設定を本番 / macOS / Docker の 3 環境に対応。Config 構造体をドメイン非依存化して Step 4 への土台を作る。

### 4.2 Step 3 共通: ブランチ・検証

- ブランチ: `feat/v2-step3`（dev-kit `main` から派生）
- 着手前必読: dev-kit v2 要件書 § 7「Step 3」（L1481-L1497）
- 完了判定:
  - `make build-release` 成功（バイナリサイズが build より小）
  - `make e2e` で Level 1 + Level 2 が一気通貫
  - `.github/workflows/` に 3 ファイルが存在
  - `falco.yaml` / `falco-local.yaml` / `falco-docker.yaml` の 3 ファイルが存在

### 4.3 T3-1: Makefile に build-release ターゲット追加

| 項目 | 内容 |
|------|------|
| 要件 ID | A4-2 |
| 優先度 | **P1 High** |
| 先行タスク | T1-2（OS 自動検出が前提） |
| 後続タスク | T3-3（CI release ジョブで使用） |
| Pコード関連 | P001, P013 |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L696-L711（A4-2） | フラグ説明 |
| openclaw `Makefile` | build-release ターゲット | 参照実装 |
| dev-kit `.claude/templates/plugin/Makefile.tmpl` | T1-2 適用後の状態 | 編集対象 |

#### 変更内容（編集対象 = `.claude/templates/plugin/Makefile.tmpl`）
```makefile
# リリース最適化ビルド（デバッグ情報除去、バイナリサイズ削減 30-50%）
build-release:
	$(GO_ENV) go build $(GO_RELEASE_FLAGS) -o $(BINARY) $(SRC_DIR)/
```
`GO_RELEASE_FLAGS` は T1-2 で導入済み（`-buildmode=c-shared -trimpath -ldflags="-s -w"`）。

#### 検証
```bash
make build && ls -l lib*.{so,dylib}
make build-release && ls -l lib*.{so,dylib}    # バイナリサイズが build より小さいこと
```

#### 完了基準
- [ ] build-release で生成バイナリが build バイナリの 50-70% 程度のサイズ
- [ ] Linux で `file lib*.so` が ELF を返す
- [ ] PR コミット粒度: 1 コミット

### 4.4 T3-2: Makefile に E2E テストターゲット追加

| 項目 | 内容 |
|------|------|
| 要件 ID | A4-3 |
| 優先度 | **P1 High** |
| 先行タスク | T1-3（Level 2）, T2-3（Level 1） |
| 後続タスク | T3-3（e2e-test ワークフローで利用） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L713-L737（A4-3） | ターゲット仕様 |
| openclaw `Makefile` | e2e-pattern / e2e-pipeline / e2e | 参照実装 |

#### 変更内容（編集対象 = `.claude/templates/plugin/Makefile.tmpl`）
```makefile
.PHONY: e2e-pattern e2e-pipeline e2e vet

# Level 1
e2e-pattern:
	go test ./test/e2e/ -v -race -run TestPattern -count=1

# Level 2
e2e-pipeline:
	go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -count=1 -timeout 120s

# Level 1 + 2（CI 高速パス）
e2e: e2e-pattern e2e-pipeline

vet:
	go vet ./...
```

#### 検証
- `make e2e` が Level 1 + Level 2 を順次実行
- どちらかが落ちると後続が実行されないこと

#### 完了基準
- [ ] e2e-pattern / e2e-pipeline / e2e の 3 ターゲットが追加
- [ ] vet ターゲットが追加
- [ ] PR コミット粒度: 1 コミット

### 4.5 T3-3: CI/CD 3 ワークフロー分離

| 項目 | 内容 |
|------|------|
| 要件 ID | A5-1 |
| 優先度 | **P1 High** |
| 先行タスク | T1-2（OS 自動検出, release.yml の matrix で必須）, T3-1（build-release）, T3-2（E2E ターゲット） |
| 後続タスク | T3-6（build スキル更新） |
| Pコード関連 | P001, P013 |

#### 背景
現在は単一 `ci.yml` に test + build + release が同居。E2E テストとマルチプラットフォームリリースが運用しにくい。

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L744-L854（A5-1） | 3 ワークフローの仕様、テンプレート例 |
| dev-kit `.claude/templates/plugin/ci.yml.tmpl` | 全体（84 行） | 現状の単一ワークフロー |
| openclaw `.github/workflows/ci.yml`, `e2e-test.yml`, `release.yml` | — | 参照実装 |

#### 変更内容
1. **既存 `ci.yml.tmpl` を更新**: release ジョブを削除、ランナーを `ubuntu-24.04` 固定、`-race` 追加、golangci-lint をバージョン固定
2. **新規 `e2e-test.yml.tmpl`** 作成: Level 1 + Level 2 を実行（要件 L778-L804）
3. **新規 `release.yml.tmpl`** 作成: matrix で linux/amd64 + darwin/arm64 ビルド、`softprops/action-gh-release@v2` で SHA256 チェックサム付きリリース（要件 L806-L853）

#### 検証
- `gh workflow list` で 3 ワークフローが認識される
- 試験的な PR で ci.yml が green、push で e2e-test.yml が動く
- `workflow_dispatch` で release.yml が動く

#### 完了基準
- [ ] 既存 ci.yml.tmpl が更新（race、ピン留め、ルール検証）
- [ ] e2e-test.yml.tmpl 新規作成
- [ ] release.yml.tmpl 新規作成
- [ ] テンプレート変数 `${PLUGIN_NAME}`, `${VERSION}` が正しく展開される
- [ ] PR コミット粒度: 1 コミット（または ci.yml 更新と新規 2 ファイルで 2-3 コミット）

### 4.6 T3-4: 3 環境 Falco 設定テンプレート作成

| 項目 | 内容 |
|------|------|
| 要件 ID | A6-1 |
| 優先度 | **P1 High** |
| 先行タスク | なし |
| 後続タスク | T3-6（build スキル）、T5-8（workflow） |
| Pコード関連 | P007（rules_files 個別パス）、P008（load_plugins）、P009（rate: 0）、P017（macOS outputs）、P018（-U） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L862-L881（A6-1） | 3 環境テーブル |
| dev-kit `.claude/templates/plugin/falco.yaml.tmpl` | 全体 | 現状の単一ファイル |
| openclaw `falco.yaml`, `falco-local.yaml`, `falco-docker.yaml` | — | 参照実装 |
| 本リポ要件 v3 §18.1 / §18.2 / §18.3 | — | claude-code 側の設定例（参考） |

#### 変更内容
1. **既存 `falco.yaml.tmpl` 更新**: Linux 本番、library_path = `/usr/share/falco/plugins/lib${PLUGIN_NAME}-plugin.so`
2. **新規 `falco-local.yaml.tmpl`**: macOS、library_path = `./lib${PLUGIN_NAME}-plugin-darwin-arm64.dylib`、`outputs:` セクションは含めない（P017）、`-U` フラグ必須注記
3. **新規 `falco-docker.yaml.tmpl`**: container、library_path = `/plugins/lib${PLUGIN_NAME}-plugin.so`、`json_output: true`
4. **3 ファイル共通の必須**:
   - `load_plugins: [${PLUGIN_NAME}]`（P008）
   - `rate: 0`, `max_burst: 0`（P009）
   - `rules_files` は個別パス（P007）

#### 検証
- 3 ファイルすべてに上記必須項目が含まれること
- macOS で `falco -c falco-local.yaml --disable-source syscall -U` がプラグインをロードできる（手動）

#### 完了基準
- [ ] falco.yaml / falco-local.yaml / falco-docker.yaml の 3 テンプレート存在
- [ ] 各ファイルにヘッダコメントで「対象環境」明記
- [ ] PR コミット粒度: 1 コミット（3 ファイル同梱）

### 4.7 T3-5: config.go.tmpl ドメイン非依存化

| 項目 | 内容 |
|------|------|
| 要件 ID | A9-1 |
| 優先度 | **P1 High** |
| 先行タスク | T2-1（MaxFieldLength 配線の前提） |
| 後続タスク | T4-1〜T4-3（PluginEvent / LogEntry 非依存化のため） |
| Pコード関連 | P020（検出 truncate vs 全文返却。MaxFieldLength のフォールバック既定値 = 10KB を踏襲） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1058-L1119（A9-1） | Before/After と配線指示 |
| dev-kit `.claude/templates/plugin/config.go.tmpl` | 全体（9 行） | 現状の構造体 |
| dev-kit `.claude/templates/plugin/parser.go.tmpl` | New() 関数 | 配線先 |
| dev-kit `.claude/templates/plugin/regex_simple.go.tmpl` | maxInputSize 定数 | 配線先 |

#### 変更内容
1. **`config.go.tmpl`**:
   - `LogFormat` の選択肢に `"auto"` を追加（T4-3 で実装する auto モード対応）
   - `LargeResponseThreshold` → `MaxFieldLength` にリネーム
2. **`parser.go.tmpl` の New() で `cfg.MaxFieldLength` を読み取る配線を追加**
3. **`SimpleSecurityDetector` 初期化時に `maxInputLength` フィールドへ伝搬**
4. **デフォルト値（0 の場合は 10KB）のフォールバック処理**

#### 検証
- `parser.New(parser.Config{MaxFieldLength: 5000})` で SimpleSecurityDetector の `maxInputLength` が 5000 になる
- `MaxFieldLength: 0` で 10KB が使われる

#### 完了基準
- [ ] LogFormat に "auto" 追加
- [ ] LargeResponseThreshold → MaxFieldLength
- [ ] parser.New() と SimpleSecurityDetector への配線完了
- [ ] フォールバックロジックがある
- [ ] PR コミット粒度: 1 コミット

### 4.8 T3-6: plugin-build スキル更新

| 項目 | 内容 |
|------|------|
| 要件 ID | B5 |
| 優先度 | **P1 High** |
| 先行タスク | T1-2, T3-1, T3-3, T3-4 |
| 後続タスク | T5-8 |
| Pコード関連 | P017, P018 |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1249-L1255（B5） | スキル更新指示 |
| dev-kit `.claude/skills/plugin-build/SKILL.md` | 全体 | 現状本文 |

#### 変更内容（編集対象 = `.claude/skills/plugin-build/SKILL.md`）
1. **macOS ネイティブビルド可能を明記**（`.dylib`）
2. **`make build-release` の説明追加**（バイナリサイズ最適化）
3. **macOS 制約の追加**: P017（outputs セクション除外）、P018（-U フラグ必須）
4. **CI/CD 3 ワークフロー** の説明追加
5. **成功基準 SC-040〜043 の更新**

#### 完了基準
- [ ] macOS ビルド手順記載
- [ ] build-release 説明
- [ ] P017/P018 注記
- [ ] PR コミット粒度: 1 コミット

### 4.9 Step 3 完了の最終確認

| 確認項目 | 方法 | 期待結果 |
|---|---|---|
| build-release | `make build-release` | サイズ最適化 |
| make e2e | `make e2e` | Level 1+2 PASS |
| CI 3 ワークフロー | `gh workflow list` | 3 件 |
| Falco 3 環境設定 | ファイル存在 | 3 ファイル |
| config 非依存化 | parser.New() 配線確認 | OK |
| build スキル | SKILL.md | macOS 手順 + P017/P018 |
| PR | dev-kit `feat/v2-step3` | 6〜8 コミット（T3-3 で ci.yml 更新と新規 2 ファイルを 2-3 コミットに分けることを許容） |

---

## §5 Step 4: ドメイン非依存化【P1 High】

### 5.1 Step 4 の目標

テンプレートを HTTP 専用から **ドメイン非依存** に進化させる。WF-Phase 0 で対話的にフィールド定義を収集し、テンプレート展開時に PluginEvent / LogEntry / Fields() / Extract() / parseLine() / parseJSON() を動的生成する。

これにより HTTP / AI ログ / IoT センサーログなど任意のドメインで生成可能になる。

> **claude-code との関係**: この Step が完了して初めて、本リポ要件 v3 §10 / §12 で定義した `claude_code.*` フィールドを scaffold 時に正しく生成できる。

### 5.2 Step 4 共通: ブランチ・検証

- ブランチ: `feat/v2-step4`（dev-kit `main` から派生）
- 着手前必読: dev-kit v2 要件書 § 7「Step 4」（L1498-L1511）
- 重要前提: T4-1 と T4-2 は **同時実装** が必須（PluginEvent と LogEntry のフィールド整合性のため）
- 完了判定: 受入テスト AT-1（HTTP）, AT-2（AI）, AT-3（IoT）が全て PASS

### 5.3 テンプレート展開機構（共通）

5 つのプレースホルダーを scaffold が WF-Phase 0 で収集したフィールド定義に基づいて展開する:

| プレースホルダー | 展開先 | 生成内容 |
|------|------|------|
| `${DOMAIN_FIELDS_STRUCT}` | plugin.go (PluginEvent), parser.go (LogEntry) | Go 構造体フィールド |
| `${DOMAIN_FIELDS_DEFS}` | plugin.go Fields() | sdk.FieldEntry 配列 |
| `${DOMAIN_FIELDS_EXTRACT}` | plugin.go Extract() | switch/case 文 |
| `${DOMAIN_FIELDS_MAPPING}` | plugin.go parseLine() | LogEntry → PluginEvent |
| `${DOMAIN_FIELDS_PARSE_JSON}` | parser.go parseJSON() | JSON → LogEntry |

scaffold の WF-Phase 0 で対話的に収集する項目（要件書 §E7 v2 新規変数 + 拡張）:
- フィールド名（snake_case）
- Go 型（string / uint64 / 必要なら time.Time）
- Falco フィールド名（`<source>.<field>`）
- JSON キー名（必要なら）
- 説明文

### 5.4 T4-1: PluginEvent のドメイン非依存化

| 項目 | 内容 |
|------|------|
| 要件 ID | A1-2 |
| 優先度 | **P1 High** |
| 先行タスク | T1-1（parser 接続済）, T3-5（config 非依存化） |
| 後続タスク | T4-2（セット）, T4-4（scaffold スキル更新） |
| Pコード関連 | P004, P010 |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L211-L296（A1-2） | 設計方針、テンプレート展開機構、設計選択肢 |
| dev-kit v2 要件書 | L271-L290（プレースホルダー表） | 展開先テンプレート、生成例 |
| dev-kit `.claude/templates/plugin/plugin.go.tmpl` | L58-L74（PluginEvent）, L307-L395（Fields/Extract）, L275-L302（parseLine） | 編集対象 |
| openclaw `cmd/plugin-sdk/plugin.go` | PluginEvent 構造体 | 参照実装 |
| 本リポ要件 v3 §10.2 | claude_code.* フィールド一覧 | dev-kit v2 完了後に生成する具体例 |

#### 変更内容（編集対象 = `.claude/templates/plugin/plugin.go.tmpl`）

1. **PluginEvent を 2 層化**:
   ```go
   type PluginEvent struct {
       // 共通フィールド（全プラグイン）
       LogPath   string
       Raw       string
       Timestamp time.Time
       Headers   map[string]string

       // ドメイン固有フィールド（scaffold が生成）
       ${DOMAIN_FIELDS_STRUCT}
   }
   ```
2. **Fields() を `${DOMAIN_FIELDS_DEFS}` プレースホルダーで置換**
3. **Extract() の switch/case を `${DOMAIN_FIELDS_EXTRACT}` で置換**
4. **parseLine() の HTTP 直接マッピングを `${DOMAIN_FIELDS_MAPPING}` に置換**
5. **Info() の Description フィールドを `${PLUGIN_DESCRIPTION}` で動的化**（v2 新規変数、要件 §E7）

#### 検証
- HTTP プラグイン生成時、従来通りの構造体・Fields/Extract が生成されること（後方互換）
- AI プラグイン生成時、`SessionID`, `Tool`, `Args`, `ThreatLevel` 等が生成されること
- Info() の Description が ${PLUGIN_DESCRIPTION} で展開されること

#### 完了基準
- [ ] テンプレートに 4 つのプレースホルダー + ${PLUGIN_DESCRIPTION} が配置
- [ ] 共通フィールドはハードコーディング維持
- [ ] Info().Description が動的化
- [ ] PR コミット粒度: T4-2 とセットで 1 コミット

### 5.5 T4-2: LogEntry のドメイン非依存化

| 項目 | 内容 |
|------|------|
| 要件 ID | A2-3 |
| 優先度 | **P1 High** |
| 先行タスク | T2-2、**T4-1（同一コミットで実装するセット）** |
| 後続タスク | T4-3, T4-4, T4-5 |
| Pコード関連 | P004 |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L497-L580（A2-3） | LogEntry 2 層化、TimeLocal 統合方針、parser_test 更新指示 |
| dev-kit `.claude/templates/plugin/parser.go.tmpl` | L41-L59（LogEntry）, parseCombined / parseCommon / parseJSON | 編集対象 |
| dev-kit `.claude/templates/plugin/parser_test.go.tmpl` | 248 行、33 TC | 更新対象（HTTP 固有 TC を移植） |

#### 変更内容（編集対象 = `.claude/templates/plugin/parser.go.tmpl`, `parser_test.go.tmpl`）

1. **LogEntry を 2 層化**:
   ```go
   type LogEntry struct {
       Timestamp      time.Time
       Raw            string
       Headers        map[string]string
       SecurityThreat SecurityThreatType
       ${DOMAIN_FIELDS_STRUCT}
   }
   ```
2. **TimeLocal を Timestamp に統合**（HTTP 固有名を排除）
3. **parseJSON()** の `${DOMAIN_FIELDS_PARSE_JSON}` を確認（T2-2 で導入済）
4. **parseCombined() / parseCommon()** を「scaffold が選択されたフォーマットに応じて生成」へ変更
   - parseCustom() は TODO のまま（ユーザー定義）
5. **`parser_test.go.tmpl` を更新**: HTTP 固有 TC は scaffold で展開時に生成、共通 TC は維持

#### 検証
- AI プラグイン生成時、`SessionID time.Time` などが LogEntry に含まれない（型安全）
- HTTP プラグイン生成時、従来通り

#### 完了基準
- [ ] LogEntry が 2 層化
- [ ] TimeLocal 削除、Timestamp 統合
- [ ] parser_test.go.tmpl が新構造に追従
- [ ] PR コミット粒度: T4-1 とセットで 1 コミット

### 5.6 T4-3: フォーマット自動検出モード追加（auto）

| 項目 | 内容 |
|------|------|
| 要件 ID | A2-2 |
| 優先度 | **P1 High**（要件書では P2 Medium だが Step 4 へ前倒し） |
| 先行タスク | T2-2 |
| 後続タスク | T4-5 |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L438-L494（A2-2） | parseAuto 実装例 |
| dev-kit `.claude/templates/plugin/parser.go.tmpl` | L96-L105（switch case） | 編集対象 |
| openclaw `pkg/parser/parser.go` | parseAuto | 参照実装 |

#### 変更内容
```go
// switch に "auto" を追加
case "auto":
    p.parseFunc = p.parseAuto
```
```go
// parseAuto は先頭文字で JSON / テキスト判定
func (p *Parser) parseAuto(line string) (*LogEntry, error) {
    trimmed := strings.TrimSpace(line)
    if len(trimmed) > 0 && trimmed[0] == '{' {
        return p.parseJSON(trimmed)
    }
    return p.parseCombined(line)
}
```
`config.go.tmpl` の Config.LogFormat 説明を更新（`"auto"` を選択肢に明示） — T3-5 で実施済の場合は確認のみ。

#### 検証
- 設定 `LogFormat: "auto"` で JSON 行と plaintext 行が混在しても両方 parse 成功

#### 完了基準
- [ ] auto モード実装、`{` を起点に JSON / 既定 fallback
- [ ] テスト追加
- [ ] PR コミット粒度: 1 コミット

### 5.7 T4-4: plugin-scaffold スキル更新

| 項目 | 内容 |
|------|------|
| 要件 ID | B1 |
| 優先度 | **P1 High** |
| 先行タスク | T4-1, T4-2 |
| 後続タスク | T5-8 |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1142-L1200（B1） | フィールド数 (14→18→20)、対応表 |
| dev-kit `.claude/skills/plugin-scaffold/SKILL.md` | 全体 | 編集対象 |
| 本リポ要件 v3 §10.2 | claude_code.* フィールド全 28 項目 | 対話収集する domain field 一覧 |
| **本リポ要件 v3 §27.3** | scaffold 入力ワークシート 19 項目 | scaffold が対話で聞く全変数の正解セット（claude-code 案件で実装者が答える内容） |

#### 変更内容（編集対象 = `.claude/skills/plugin-scaffold/SKILL.md`）
1. **WF-Phase 0 にフィールド対話収集を追加** — 名前 / 型 / Falco フィールド名 / JSON キー名 / 説明
2. **生成ファイル一覧を更新**: 14 → 18（Step 4 完了時点）→ 20（Step 5 完了時 CLAUDE.md/CHANGELOG.md 追加後）
3. **5 プレースホルダーの展開ルール**を記載
4. **テンプレート→スキル対応表を更新**（要件書 L1167-L1195 を参照）
5. **e2e/scripts/ ディレクトリの作成手順を追加**（Level 3 はテンプレート化対象外、ディレクトリのみ）

#### 完了基準
- [ ] フィールド対話収集の手順が明記
- [ ] 18 ファイル一覧と対応表（Step 5 完了時 20）
- [ ] 5 プレースホルダー説明
- [ ] **成功基準 SC-001〜SC-008 を新仕様に合わせて更新**（フィールド収集、ドメイン非依存生成、go vet パス、ディレクトリ構造）
- [ ] PR コミット粒度: 1 コミット

### 5.8 T4-5: plugin-parser スキル更新

| 項目 | 内容 |
|------|------|
| 要件 ID | B2 |
| 優先度 | **P1 High** |
| 先行タスク | T2-2, T4-3 |
| 後続タスク | T5-8 |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1202-L1209（B2） | スキル更新指示 |
| dev-kit `.claude/skills/plugin-parser/SKILL.md` | 全体 | 編集対象 |

#### 変更内容
1. **JSON パーサー** が「未実装」から「デフォルト実装あり、ドメイン固有フィールドは展開で追加」に変更
2. **フォーマット検出** に `"auto"` モードを追加
3. **入力サイズ超過** を「スキップ」から「先頭 10KB に切り詰めて続行」に修正

#### 完了基準
- [ ] JSON, auto, truncate 方針が SKILL.md に記載
- [ ] **成功基準 SC-010〜SC-013 を更新**（パーサー実装、セキュリティ検出、auto モード、入力サイズ）
- [ ] PR コミット粒度: 1 コミット

### 5.9 Step 4 完了の最終確認（受入テスト）

| TC ID | テスト | コマンド | 期待結果 |
|---|---|---|---|
| AT-1 | HTTP プラグイン生成 | scaffold で format=combined, HTTP 標準 | go vet/test/build PASS |
| AT-2 | AI プラグイン生成 | scaffold で format=json, type/tool/args/session_id | 同上 |
| AT-3 | IoT プラグイン生成 | scaffold で format=custom, device_id/sensor_type/value | 同上 |
| AT-4 | macOS ビルド | AT-1 を macOS arm64 で | .dylib 生成 |
| AT-5 | E2E | AT-1 で `make e2e` | Level 1+2 PASS |

| 確認項目 | 方法 | 期待 |
|---|---|---|
| プレースホルダー展開 | 3 種類で生成 | 全フィールドが正しく展開 |
| 後方互換 | HTTP 生成 | 従来テスト全 PASS |
| scaffold スキル | SKILL.md | 18 ファイル + 5 プレースホルダー説明 |
| parser スキル | SKILL.md | JSON / auto / truncate 反映 |
| PR | dev-kit `feat/v2-step4` | T4-1+T4-2 同梱、その他 1 コミットずつ |

---

## §6 Step 5: ドキュメント・仕組み化【P2 Medium】

### 6.1 Step 5 の目標

ドキュメント整備、軽微なコード改善、フィードバック自動化、ワークフローエージェントの最終仕上げを行う。本 Step が完了すると dev-kit v2 のすべての改善が完成し、本リポジトリ（claude-code）で `/plugin-scaffold claude-code json` を起動して claude-code プラグイン実装に進める状態になる。

### 6.2 Step 5 共通: ブランチ・検証

- ブランチ: `feat/v2-step5`（dev-kit `main` から派生）
- 着手前必読: dev-kit v2 要件書 § 7「Step 5」（L1512-L1528）
- 完了判定: 全 Step の品質ゲート通過 + Step 5 個別タスクの完了

### 6.3 T5-1: `~/`パス展開ロジック + パストラバーサル防止

| 項目 | 内容 |
|------|------|
| 要件 ID | A1-3 / E6 |
| 優先度 | **P2 Medium** |
| 先行タスク | T1-1 |
| 後続タスク | なし |
| Pコード関連 | E6（パストラバーサル防止） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L301-L325（A1-3） | 実装例 |
| dev-kit v2 要件書 | L1378-L1387（E6 セキュリティ要件） | パストラバーサル基準 |
| dev-kit `.claude/templates/plugin/plugin.go.tmpl` | Open() 内 | 編集対象 |

#### 変更内容（編集対象 = `plugin.go.tmpl` の Open() 内）
```go
for _, logPath := range p.config.LogPaths {
    // パストラバーサル防止（E6）
    if strings.Contains(logPath, "..") {
        return nil, fmt.Errorf("path traversal not allowed: %s", logPath)
    }
    // ~/ 展開
    if strings.HasPrefix(logPath, "~/") {
        if home, err := os.UserHomeDir(); err == nil {
            logPath = filepath.Join(home, logPath[2:])
        }
    }
    // 既存のファイルオープン処理...
}
```

#### 検証
- 設定 `log_paths: ["~/test/app.log"]` で `$HOME/test/app.log` が監視される
- `log_paths: ["../../etc/passwd"]` でエラー

#### 完了基準
- [ ] `..` 含むパスでエラー
- [ ] `~/` が正しく展開
- [ ] テスト追加（TC-6-01 相当）
- [ ] PR コミット粒度: 1 コミット

### 6.4 T5-2: Extract() 冗長 nil チェック削除

| 項目 | 内容 |
|------|------|
| 要件 ID | A1-4 |
| 優先度 | **P3 Low**（要件書では P2 Medium 群に） |
| 先行タスク | なし |
| 後続タスク | なし |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L329-L361（A1-4） | Before/After |
| dev-kit `.claude/templates/plugin/plugin.go.tmpl` | L332-L343（Extract）| 編集対象 |

#### 変更内容（編集対象 = `plugin.go.tmpl` の Extract()）
```go
// Before: data := evt.EventData() の nil チェックを行ってから Reader() を使う冗長コード
// After:
func (p *MyPlugin) Extract(req sdk.ExtractRequest, evt sdk.EventReader) error {
    var event PluginEvent
    decoder := gob.NewDecoder(evt.Reader())   // 直接 Reader を使う
    if err := decoder.Decode(&event); err != nil {
        return fmt.Errorf("failed to decode event: %w", err)
    }
    // ...
}
```

#### 完了基準
- [ ] 不要な EventData() 呼び出しと nil チェックを削除
- [ ] PR コミット粒度: 1 コミット

### 6.5 T5-3: URL デコード重複排除（※ T1-1 に統合済み）

| 項目 | 内容 |
|------|------|
| 要件 ID | A3-2 |
| ステータス | **統合済み（R5-001）** |
| 注記 | T1-1 の parser 接続で URL デコードが `detectSecurityPatterns()` 一箇所に集約された結果、本タスクは不要となった |

本 Step 5 で実施することはない。dev-kit 側 TASK_DEFINITIONS.md と本リポ要件の整合のために ID は維持する。

### 6.6 T5-4: CLAUDE.md.tmpl 新規作成

| 項目 | 内容 |
|------|------|
| 要件 ID | A8-1 |
| 優先度 | **P2 Medium** |
| 先行タスク | なし |
| 後続タスク | T4-4（scaffold ファイル一覧 18→20） |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L998-L1027（A8-1） | テンプレート概要 |
| openclaw `CLAUDE.md` | 全体 | 参照実装 |
| 本リポ `CLAUDE.md` | 全体 | claude-code 側 CLAUDE.md の例 |

#### 変更内容（新規作成 = `.claude/templates/plugin/CLAUDE.md.tmpl`）
要件書 L1006-L1024 のテンプレート構造を採用:
- Project Overview
- Build & Development Commands
- Architecture（plugin / parser 層の説明）
- Critical Constraints（P002, P004, P008, P010 等）

テンプレート変数:
- `${PLUGIN_NAME}`, `${LOG_SOURCE}`（v2 新規変数）

#### 完了基準
- [ ] テンプレート作成
- [ ] 主要 P コードの注意書きを含む
- [ ] **`.claude/skills/plugin-scaffold/SKILL.md` のファイル一覧を 18 → 20 に更新（CLAUDE.md.tmpl と CHANGELOG.md.tmpl を反映）**
- [ ] PR コミット粒度: 1 コミット（T5-5 と同梱可）

### 6.7 T5-5: CHANGELOG.md.tmpl 新規作成

| 項目 | 内容 |
|------|------|
| 要件 ID | A8-2 |
| 優先度 | **P2 Medium** |
| 先行タスク | なし |
| 後続タスク | なし |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1030-L1036（A8-2） | スケルトン仕様 |
| openclaw `CHANGELOG.md` | 全体 | 参照実装 |

#### 変更内容
[Keep a Changelog](https://keepachangelog.com/) 準拠スケルトン:
- `## [Unreleased]`
- 各カテゴリ（Added / Changed / Deprecated / Removed / Fixed / Security）のプレースホルダー
- 初回リリースとして `## [${VERSION}] - YYYY-MM-DD` の雛形

#### 完了基準
- [ ] テンプレート作成
- [ ] Keep a Changelog 準拠
- [ ] **`.claude/skills/plugin-scaffold/SKILL.md` のファイル一覧を 20 化（T5-4 と同梱で 1 度に更新可）**
- [ ] PR コミット粒度: 1 コミット（T5-4 と同梱可）

### 6.8 T5-6: README.md.tmpl 更新

| 項目 | 内容 |
|------|------|
| 要件 ID | A8-3 |
| 優先度 | **P2 Medium** |
| 先行タスク | T1-2, T3-2, T3-4, T1-3, T2-3 |
| 後続タスク | なし |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1038-L1050（A8-3） | 影響表 |
| dev-kit `.claude/templates/plugin/README.md.tmpl` | 全体 | 編集対象 |

#### 変更内容（編集対象 = `README.md.tmpl`）
| 影響元 | 追記内容 |
|--------|---------|
| A4-1 | macOS / Linux 両対応のビルドコマンド説明 |
| A4-3 | `make e2e`, `make e2e-pattern`, `make e2e-pipeline` の説明 |
| A6-1 | 3 環境設定（falco.yaml / falco-local.yaml / falco-docker.yaml）の使い分け |
| A7-1/A7-2 | 3 層テストアーキテクチャの概要 |

#### 完了基準
- [ ] 4 セクションが追加 or 更新されている
- [ ] PR コミット粒度: 1 コミット

### 6.9 T5-7: dev-kit-feedback スキル新規作成

| 項目 | 内容 |
|------|------|
| 要件 ID | C1 |
| 優先度 | **P2 Medium** |
| 先行タスク | T2-5（PROBLEM_PATTERNS への追加候補提示の比較基盤） |
| 後続タスク | なし |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1273-L1289（C1） | 処理フロー |
| dev-kit `.claude/skills/`（既存スキルの構造） | SKILL.md ファイル形式 | 雛形 |

#### 変更内容（新規作成 = `.claude/skills/dev-kit-feedback/SKILL.md`）
1. **コマンド**: `/dev-kit-feedback [plugin-path]`
2. **処理フロー**:
   1. 指定パスのプラグインコードを読み込む
   2. dev-kit テンプレートとの差分を検出（`diff` ベース）
   3. 差分を分類（テンプレート改善 / スキル改善 / 新規パターン）
   4. 改善提案レポートを出力
   5. PROBLEM_PATTERNS.md への追加候補を提示
3. **入力**: プラグイン root ディレクトリ
4. **出力**: Markdown レポート（標準出力 + 任意で `feedback-report.md`）

#### 完了基準
- [ ] SKILL.md がフォーマットに従う
- [ ] 処理フロー 5 ステップを記載
- [ ] PR コミット粒度: 1 コミット

### 6.10 T5-8: plugin-dev-workflow エージェント更新

| 項目 | 内容 |
|------|------|
| 要件 ID | B6 |
| 優先度 | **P2 Medium** |
| 先行タスク | **全タスク（T1-1〜T5-7 + T5-9）が完了していること**。T5-9（rules スキル）も WF-Phase 3 反映に必要なため含む |
| 後続タスク | なし |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1257-L1267（B6） | 変更項目 |
| dev-kit `.claude/agents/plugin-dev-workflow.md` | 全体 | 編集対象 |

#### 変更内容（編集対象 = `plugin-dev-workflow.md`）
| 変更内容 | 対象 WF-Phase |
|---------|------------|
| ドメイン固有フィールド対話収集を追加 | WF-Phase 0 |
| parser 統合を自動実行 | WF-Phase 2 |
| Level 2 テスト生成を追加 | WF-Phase 4 |
| macOS ネイティブビルド | WF-Phase 5 |
| 品質ゲートに Level 2 通過を追加 | WF-Phase 4→5 |
| 完了報告に E2E 結果サマリー | WF-Phase 6 |
| 新規テンプレートファイル一覧（20 ファイル）に更新 | WF-Phase 1 |

#### 完了基準
- [ ] 7 項目すべて反映
- [ ] 全 Step の最終確認後に PR
- [ ] PR コミット粒度: 1 コミット

### 6.11 T5-9: plugin-rules スキル更新

| 項目 | 内容 |
|------|------|
| 要件 ID | B3 |
| 優先度 | **P2 Medium** |
| 先行タスク | T2-5（PROBLEM_PATTERNS） |
| 後続タスク | なし |
| Pコード関連 | P003, P005, P011, P012, P015, P019 |

#### コンテキスト復元
| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| dev-kit v2 要件書 | L1211-L1215（B3） | スキル更新指示 |
| dev-kit `.claude/skills/plugin-rules/SKILL.md` | 全体 | 編集対象 |
| dev-kit `PROBLEM_PATTERNS.md` | T2-5 後の P コード一覧 | 参照 |
| 本リポ要件 v3 §13 R-001〜R-011 | — | rules ベストプラクティス |

#### 変更内容（編集対象 = `plugin-rules/SKILL.md`）
1. **ドメイン非依存ルール構造ガイド** を追加（HTTP に限らない例: AI ログ、IoT）
2. **priority 使い分け基準**:
   - CRITICAL: 即時対応（バイパスモード、機密流出など）
   - WARNING: 高頻度許容（設定変更など）
   - NOTICE: 集計対象（subagent 起動、tool 多発など）
3. **PROBLEM_PATTERNS への参照追加**（P003, P005, P011, P012, P015, P019）
4. **boolean field の比較ルール**（本リポ §13 R-011: `claude_code.dropped = "true"` の文字列リテラル）

#### 完了基準
- [ ] ドメイン非依存ガイド追加
- [ ] priority 凡例追加
- [ ] PROBLEM_PATTERNS リンク
- [ ] PR コミット粒度: 1 コミット

### 6.12 Step 5 完了の最終確認（dev-kit v2 全体）

| 確認項目 | 方法 | 期待結果 |
|---|---|---|
| パス展開 / トラバーサル防止 | テスト | OK |
| Extract() 冗長削除 | コード差分 | OK |
| CLAUDE.md / CHANGELOG.md 生成 | scaffold | 20 ファイル一覧 |
| README 更新 | 目視 | 4 セクション追加 |
| dev-kit-feedback スキル | `/dev-kit-feedback ...` | レポート出力 |
| workflow エージェント | 全 Phase の改修反映 | OK |
| rules スキル | priority / PROBLEM_PATTERNS 反映 | OK |
| **受入テスト AT-1〜AT-5** | Step 4 で実施したものを再実行 | 全 PASS |
| **dev-kit v2 全体の go vet/test/build** | 任意のドメインで生成 | 全 PASS |
| PR | dev-kit `feat/v2-step5` | 7〜8 コミット（T5-3 は統合済みでスキップ、T5-4+T5-5 は同梱可） |

### 6.13 dev-kit v2 完了後の claude-code 側ロードマップ

dev-kit v2 が main にマージされた後、本リポジトリで claude-code プラグインの実装に進む。

詳細は **§8.3.2 dev-kit v2 完了後の claude-code 実装ロードマップ** および **本リポ要件 v3 §19（開発ワークフロー要件）** を参照。

---

## §7 クロスリファレンス

### 7.1 要件 ID → タスク 逆引き

dev-kit v2 要件書の改善 ID を、本タスク定義書のどのタスクで実装するかの逆引き表。

#### 7.1.1 A. テンプレート改善

| 要件 ID | 名称 | 担当タスク | Step | 優先度 |
|---|---|---|---:|---|
| A1-1 | parseLine() と parser 接続 | T1-1 | 1 | P0 |
| A1-2 | PluginEvent ドメイン非依存化 | T4-1 | 4 | P1 |
| A1-3 | `~/`パス展開 + トラバーサル防止 | T5-1 | 5 | P2 |
| A1-4 | Extract() 冗長 nil チェック削除 | T5-2 | 5 | P3 |
| A2-1 | JSON パーサーのデフォルト実装 | T2-2 | 2 | P1 |
| A2-2 | フォーマット自動検出 | T4-3 | 4 | P1 |
| A2-3 | LogEntry ドメイン非依存化 | T4-2 | 4 | P1 |
| A3-1 | 入力サイズ超過時 truncate | T2-1 | 2 | P1 |
| A3-2 | URL デコード重複排除 | T1-1 に統合 | 1 | — |
| A4-1 | OS/Arch 自動検出 | T1-2 | 1 | P0 |
| A4-2 | build-release ターゲット | T3-1 | 3 | P1 |
| A4-3 | E2E ターゲット（pattern/pipeline/e2e） | T3-2 | 3 | P1 |
| A5-1 | CI 3 ワークフロー分離 | T3-3 | 3 | P1 |
| A6-1 | 3 環境 Falco 設定 | T3-4 | 3 | P1 |
| A7-1 | Level 2 パイプラインテストテンプレート | T1-3 | 1 | P0 |
| A7-2 | Level 1 E2E パターンテストテンプレート | T2-3 | 2 | P1 |
| A7-3 | benign + edge_cases パターン拡張 | T2-4 | 2 | P1 |
| A8-1 | CLAUDE.md.tmpl 新規 | T5-4 | 5 | P2 |
| A8-2 | CHANGELOG.md.tmpl 新規 | T5-5 | 5 | P2 |
| A8-3 | README.md.tmpl 更新 | T5-6 | 5 | P2 |
| A9-1 | config.go.tmpl ドメイン非依存化 | T3-5 | 3 | P1 |

#### 7.1.2 B. スキル・エージェント

| 要件 ID | 名称 | 担当タスク | Step | 優先度 |
|---|---|---|---:|---|
| B1 | plugin-scaffold スキル | T4-4 | 4 | P1 |
| B2 | plugin-parser スキル | T4-5 | 4 | P1 |
| B3 | plugin-rules スキル | T5-9 | 5 | P2 |
| B4 | plugin-test スキル | T1-4 | 1 | P0 |
| B5 | plugin-build スキル | T3-6 | 3 | P1 |
| B6 | plugin-dev-workflow エージェント | T5-8 | 5 | P2 |

#### 7.1.3 C. 新規追加

| 要件 ID | 名称 | 担当タスク | Step | 優先度 |
|---|---|---|---:|---|
| C1 | dev-kit-feedback スキル | T5-7 | 5 | P2 |
| C2 | PROBLEM_PATTERNS.md 知見追加 | T2-5 | 2 | P1 |

#### 7.1.4 E. 非機能要件

| 要件 ID | 名称 | 関連タスク | 検証方法 |
|---|---|---|---|
| E1 | 後方互換性 | 全タスク | テスト用プラグインで go vet + go test |
| E2 | ドメイン非依存性 | T4-1, T4-2, T4-4 | 受入テスト AT-1〜AT-3 |
| E3 | ドキュメント整合性 | T1-4, T3-6, T4-4, T4-5, T5-6, T5-8, T5-9 | SKILL.md 全数チェック |
| E4 | 互換性（Go 1.22+ / Falco 0.43+） | T3-3 | CI のビルド確認 |
| E5 | 性能（100 events/sec） | T1-3 | TC-5-01 |
| E6 | セキュリティ | T2-1, T5-1 | テスト |
| E7 | テンプレート変数仕様 | T4-4, T5-4 | scaffold スキル仕様 |
| E8 | 受入テスト | Step 4 / Step 5 | AT-1〜AT-5 |

### 7.2 P コード → タスク 逆引き

> **注**: §1.5.5 と同じ重要注意（dev-kit Pコード ≠ 要件 v3 §18.4 の解釈）が適用される。Pコードの意味解釈は dev-kit `PROBLEM_PATTERNS.md`（T2-5 集約）を真とし、要件 v3 §18.4 列は claude-code 側の同番号要件節への参照を示す。

| P コード | 概要 | 対応タスク | claude-code 要件 v3 §18.4 |
|---|---|---|---|
| P001 | macOS バイナリ Linux 配布 | T1-2, T3-1, T3-3 | §7.4, §21.1, §21.2 B-003/B-004, AC-013 |
| P002 | -buildmode=c-shared 不指定 | T1-2, T3-1 | §21.2 B-001, §27.2 |
| P003 | rule に source 不指定 | T5-9 | §13 R-001, §18.3, §29 |
| P004 | GOB nil map panic | T1-1, T1-3, T2-2, T4-1 | §14.1 P-005 |
| P005 | rule で evt.type 使用 | T5-9 | §13 R-002, §18.3, §29 |
| P006 | URL エンコードパターン | T2-2 | （rules 側） |
| P007 | rules_files 個別パス | T3-4 | §18.3 |
| P008 | load_plugins 欠落 | T3-4 | §18.3 |
| P009 | レート制限でアラート抑制 | T3-4 | §18.3 |
| P010 | Fields/Extract 不一致 | T1-1, T4-1 | §6.3 FP-008/FP-009, §23.2 |
| P011 | YAML コメント位置 | T5-9 | §13 R-010 |
| P012 | headers 参照は小文字 | T5-9 | §13 R-* |
| P013 | ビルド環境不一致 | T1-2, T3-1, T3-6 | §21.2 B-002/B-005 |
| P014 | ファイル SeekEnd 必須 | T1-1, T1-3 | §6.2 ES-004, §14.1 P-006 |
| P015 | クロスルール干渉 | T5-9 | §13 R-005 |
| P016 | URL エンコード多段 | T2-1, T2-2 | （rules 側） |
| P017 | macOS Falco outputs 拒否 | T3-4, T3-6 | §7.2 MAC-002, §18.1 |
| P018 | macOS -U フラグ必須 | T3-4, T3-6 | §7.2 MAC-003, §18.3 |
| P019 | Falco 1 イベント 1 ルール | T5-9 | §13 R-005/R-006 |
| P020 | 検出 truncate vs 全文返却 | T2-1 | §14.2 D-004 |
| P021 | fsnotify タイミング | T1-3 | §6.3 FP-006 |

### 7.3 スキル → タスク 逆引き

| スキル | 担当タスク | 主な変更 |
|---|---|---|
| plugin-scaffold | T4-4 | フィールド対話収集、18→20 ファイル、5 プレースホルダー |
| plugin-parser | T4-5 | JSON デフォルト実装、auto モード、truncate |
| plugin-rules | T5-9 | ドメイン非依存ガイド、priority 凡例、PROBLEM_PATTERNS 参照 |
| plugin-test | T1-4 | 3 層 E2E アーキテクチャ |
| plugin-build | T3-6 | macOS ネイティブ、build-release、P017/P018 |
| plugin-dev-workflow | T5-8 | 全 Phase の改修反映 |
| dev-kit-feedback（新規） | T5-7 | フィードバック自動化 |

### 7.4 テンプレート → タスク 逆引き

#### 7.4.1 編集タスク

| テンプレート | 担当タスク | 変更概要 |
|---|---|---|
| `plugin.go.tmpl` | T1-1, T4-1, T5-1, T5-2 | parser 接続 / PluginEvent 非依存化 / パス展開 / Extract 冗長削除 |
| `parser.go.tmpl` | T2-2, T4-2, T4-3 | parseJSON 実装 / LogEntry 非依存化 / auto モード |
| `regex_simple.go.tmpl` | T2-1 | truncate 化 |
| `config.go.tmpl` | T3-5 | LogFormat に "auto"、MaxFieldLength リネーム |
| `Makefile.tmpl` | T1-2, T3-1, T3-2 | OS 自動検出 / build-release / E2E ターゲット |
| `ci.yml.tmpl` | T3-3 | release ジョブ削除、ピン留め、race 追加 |
| `falco.yaml.tmpl` | T3-4 | Linux 本番設定 |
| `e2e_pattern.json.tmpl` | T2-4 | benign + edge_cases 追加 |
| `parser_test.go.tmpl` | T4-2 | LogEntry 非依存化対応 |
| `README.md.tmpl` | T5-6 | macOS ビルド / E2E / 3 環境 / 3 層テスト |

#### 7.4.2 新規作成タスク（合計 8 ファイル）

| テンプレート | 担当タスク | 用途 |
|---|---|---|
| `plugin_test.go.tmpl` | T1-3 | Level 2 パイプラインテスト |
| `e2e_pattern_test.go.tmpl` | T2-3 | Level 1 E2E パターンテスト |
| `falco-local.yaml.tmpl` | T3-4 | macOS ローカル |
| `falco-docker.yaml.tmpl` | T3-4 | Docker |
| `e2e-test.yml.tmpl` | T3-3 | E2E + Allure ワークフロー |
| `release.yml.tmpl` | T3-3 | マルチプラットフォームリリース |
| `CLAUDE.md.tmpl` | T5-4 | プロジェクトガイド |
| `CHANGELOG.md.tmpl` | T5-5 | Keep a Changelog 準拠 |

### 7.5 テンプレート変数 → タスク

| 変数 | 種別 | 関連タスク |
|---|---|---|
| `${PLUGIN_NAME}` | 既存 | 全タスク |
| `${PLUGIN_NAME_UPPER}` | 既存 | 全タスク（plugin_rules.yaml.tmpl, e2e_pattern.json.tmpl 等で使用。値は scaffold が PLUGIN_NAME を upper-case 化して生成） |
| `${PLUGIN_NAME_CAMEL}` | 既存 | T4-4 |
| `${PLUGIN_ID}` | 既存 | T4-4 |
| `${EVENT_SOURCE}` | 既存 | T4-4 |
| `${VERSION}` | 既存 | T3-3, T5-5 |
| `${AUTHOR}` | 既存 | T1-1（import パス） |
| `${LOG_PATH_DEFAULT}` | 既存 | T4-4 |
| `${LOG_FORMAT}` | 既存（v2 で `auto` 拡張） | T1-1, T2-2, T3-5, T4-3 |
| `${SDK_VERSION}` | 既存 | T4-4 |
| `${TIME_FORMAT}` | 既存 | T1-1 |
| `${LICENSE}` | 既存 | T4-4 |
| `${YEAR}` | 既存 | T4-4 |
| `${PLUGIN_DESCRIPTION}` | **v2 新規** | T4-1（PluginEvent 非依存化と同時） |
| `${LOG_SOURCE}` | **v2 新規** | T5-4（CLAUDE.md.tmpl） |
| `${DOMAIN_FIELDS_STRUCT}` | **v2 新規プレースホルダー** | T4-1, T4-2 |
| `${DOMAIN_FIELDS_DEFS}` | **v2 新規プレースホルダー** | T4-1 |
| `${DOMAIN_FIELDS_EXTRACT}` | **v2 新規プレースホルダー** | T4-1 |
| `${DOMAIN_FIELDS_MAPPING}` | **v2 新規プレースホルダー** | T4-1 |
| `${DOMAIN_FIELDS_PARSE_JSON}` | **v2 新規プレースホルダー** | T2-2, T4-2 |

### 7.6 WF-Phase ↔ タスク

dev-kit の `plugin-dev-workflow` エージェント（T5-8 で更新）が実行する WF-Phase 0〜6 と本タスク定義書の関係:

| WF-Phase | 内容 | 関連タスク |
|---|---|---|
| WF-Phase 0 | 要件確認・フィールド対話収集 | T4-4 |
| WF-Phase 1 | scaffold（テンプレート展開・生成） | T1-2（Makefile）, T4-1, T4-2（プレースホルダー展開機構）, T4-4（scaffold スキル本体） |
| WF-Phase 2 | parser 実装 | T1-1, T2-2, T4-2, T4-3, T4-5 |
| WF-Phase 3 | rules 作成 | T5-9 |
| WF-Phase 4 | テスト作成 / 実行 | T1-3, T2-3, T2-4, T1-4 |
| WF-Phase 5 | ビルド / リリース | T1-2, T3-1, T3-2, T3-3, T3-6 |
| WF-Phase 6 | 完了レポート / フィードバック | T5-7, T5-8 |

### 7.7 SC コード → タスク

各スキルの成功基準（SC-001〜SC-053）を更新するタスク:

| SC コード範囲 | スキル | 担当タスク |
|---|---|---|
| SC-001〜SC-008 | plugin-scaffold | T4-4 |
| SC-010〜SC-013 | plugin-parser | T4-5 |
| SC-020〜SC-024 | plugin-rules | T5-9 |
| SC-030〜SC-033 | plugin-test | T1-4 |
| SC-040〜SC-043 | plugin-build | T3-6 |
| SC-050〜SC-053 | plugin-dev-workflow | T5-8 |

> **飛び番について**: SC-009, SC-014〜019, SC-025〜029, SC-034〜039, SC-044〜049 は意図的な空き番号で、各スキルの将来拡張（v3 以降の追加成功基準）用に予約されている。本書では使用しない。

---

## §8 リスク・受入テスト・移行戦略

### 8.1 主要リスク（プロジェクト全体）

#### 8.1.1 技術リスク

| ID | リスク | 影響 | 確率 | 対策 | 発火点となるタスク |
|---|---|---|---|---|---|
| TR-D1 | テンプレート展開の文字列置換が複雑化し、バグが混入 | 高 | 中 | T4-1/T4-2 完了時にプレースホルダーごとの単体テストを追加。3 ドメイン (HTTP/AI/IoT) で生成して回帰テスト | T4-1, T4-2 |
| TR-D2 | T4-1（PluginEvent 非依存化）と T4-2（LogEntry）のフィールド整合が崩れて parseLine() が破綻 | 高 | 中 | 両者を同一コミットで実装。受入テスト AT-1 で go test を必須通過 | T4-1, T4-2 |
| TR-D3 | T1-3 のテストが fsnotify タイミング依存で flaky 化（P021） | 中 | 高 | sleep の理由を必ずコメント化。CI は 10 連続実行で flaky を計測 | T1-3 |
| TR-D4 | macOS arm64 と amd64 の両方をテストできず、Apple Silicon/Intel どちらかが破損 | 中 | 中 | T1-2 では UNAME_M 分岐を明示。CI は arm64 のみだが、開発者の amd64 環境で別途確認を推奨 | T1-2, T3-1 |
| TR-D5 | dev-kit v2 完了前に claude-code 開発を進めてしまい、テンプレート問題に直接ぶつかる | 高 | 中 | 本タスク定義書で明示し、claude-code 側 Phase 1 着手は v2 マージ後とする | プロジェクト管理 |
| TR-D6 | Falco 0.43.0 の outputs/`-U` 仕様（P017/P018）が将来バージョンで変更 | 中 | 低 | CI で複数 Falco バージョンに対するスモークテストを v0.5 で追加検討（v0.1 では非対象） | T3-4, T3-6 |
| TR-D7 | プレースホルダー展開で生成された Go コードが構文エラー | 高 | 低 | scaffold 後に go vet を必須化。失敗時はテンプレートと展開ロジックを Issue 化 | T4-1, T4-4 |
| TR-D8 | URL デコード重複排除（A3-2 → T1-1 統合）の実装で 1 箇所削除を忘れる | 中 | 低 | T1-1 完了時に regex_simple.go.tmpl の該当箇所を削除した diff を確認 | T1-1 |
| TR-D9 | macOS で `make build` 後の `.dylib` を Linux 用としてリリースに含めてしまう（P001） | 高 | 低 | T3-3 release.yml.tmpl で OS 別 matrix 化。SHA256 検証 | T3-3 |
| TR-D10 | Step 5 の workflow エージェント更新（T5-8）漏れで scaffold が古い動作 | 中 | 中 | T5-8 を最後に置き、依存関係チェックリストを PR テンプレート化 | T5-8 |

#### 8.1.2 プロセスリスク

| ID | リスク | 影響 | 対策 |
|---|---|---|---|
| PR-1 | Step 1〜5 のレビュー・マージが滞り、claude-code 実装が長期ブロック | 中 | 各 Step を 1 PR にし 1 週間以内マージを目標。レビュー観点は本書を参照 |
| PR-2 | 既存プラグイン（nginx-plugin / openclaw）への破壊的変更が混入 | 高 | E1 後方互換性: 既存 2 リポを clone して go vet/test を CI に組み込む（v0.2 移行） |
| PR-3 | 要件書バージョンと TASK 定義書のドリフト | 中 | 要件書の改訂履歴（v5.6 以降）を変更したら本書も同期。CI で参照 ID の存在を grep |

### 8.2 受入テスト（dev-kit v2 全体）

要件書 §E8 の AT-1〜AT-5 に加え、本タスク定義書として実施すべき統合受入テストを示す。

#### 8.2.1 必須受入テスト（要件 §E8）

| TC ID | 入力 | 実行コマンド | 期待結果 | 実施 Step |
|---|---|---|---|---|
| AT-1 | format=combined, HTTP 標準フィールド | scaffold → go vet → go test → make build | 全成功 | Step 4 完了 |
| AT-2 | format=json, type/tool/args/session_id | 同上 | 全成功 | Step 4 完了 |
| AT-3 | format=custom, device_id/sensor_type/value(string) | 同上 | 全成功 | Step 4 完了 |
| AT-4 | AT-1 を macOS arm64 で実行 | make build | `.dylib` 生成 | Step 1 / Step 4 |
| AT-5 | AT-1 で make e2e | make e2e | Level 1 + Level 2 全 PASS | Step 3 / Step 4 |

#### 8.2.2 拡張受入テスト（本書追加）

| TC ID | 内容 | コマンド | 期待結果 |
|---|---|---|---|
| ET-1 | Step 完了ごとの go vet | 各 Step 完了時に test-plugin 生成 → go vet | エラー 0 |
| ET-2 | flaky テストの 10 連続実行 | T1-3 完了後 `for i in {1..10}; do make e2e-pipeline; done` | 全 10 回 PASS |
| ET-3 | `.dylib` / `.so` 切り分け | macOS と Linux で make build | OS ごとに正しい拡張子 |
| ET-4 | release ワークフロー | dev-kit `gh workflow run release.yml -f version=v0.0.0-test` | matrix で 2 バイナリ + checksums.sha256 生成 |
| ET-5 | dev-kit-feedback スキル動作 | `/dev-kit-feedback openclaw` | 改善提案レポート出力 |
| ET-6 | claude-code scaffold | dev-kit v2 完了後、本リポで `/plugin-scaffold claude-code json` | `claude-code/` ディレクトリ生成、go vet PASS |
| ET-7 | 既存プラグイン回帰確認 | 各 Step 完了時に `falco-plugin-nginx` / `falco-plugin-openclaw` に対して `go vet ./...` と `go test ./...` を実行 | 全 PASS（テンプレート変更が既存実装を壊していない） |

#### 8.2.3 受入テスト実施タイミング

| Step 完了時 | 実施 TC | スコープ |
|---|---|---|
| Step 1 | AT-4, ET-1, ET-3, **ET-7** | 基盤確認 + 既存プラグイン回帰 |
| Step 2 | ET-1, ET-2, **ET-7** | テスト基盤確認 + 既存プラグイン回帰 |
| Step 3 | AT-5, ET-1, ET-3, ET-4, **ET-7** | CI/CD 含む確認 + 既存プラグイン回帰 |
| Step 4 | **AT-1, AT-2, AT-3**, AT-4, AT-5, ET-1, **ET-7** | ドメイン非依存性確認（最重要） + 既存プラグイン回帰 |
| Step 5 | 全 AT/ET、特に ET-5, ET-6, ET-7 | dev-kit v2 完成確認 |

### 8.3 移行戦略

#### 8.3.1 既存プラグインへの影響

| 既存プラグイン | 影響 | 対応 |
|---|---|---|
| `falco-plugin-nginx` | テンプレートを直接使っていない（既存コード） | dev-kit v2 マージ後に scaffold ベースで再生成は不要。E1 後方互換性で確認 |
| `falco-plugin-openclaw` | テンプレートを直接使っていない | 同上 |
| `falco-plugin-claude_code`（本リポ） | 未生成。dev-kit v2 完了後に scaffold で初回生成 | 本リポ要件 v3 §19 のワークフローに従う |

#### 8.3.2 dev-kit v2 完了後の claude-code 実装ロードマップ

```
[dev-kit feat/v2-step5 マージ]
       ↓
[dev-kit v0.2.0 タグ]
       ↓
本リポ (/Users/takaos/lab/falco-plugin-claude_code/) で /plugin-scaffold claude-code json
       ↓
WF-Phase 0: 対話入力 — 完全な入力ワークシートは要件 v3 §27.3 を参照
       ↓
WF-Phase 1: scaffold 実行
   ⇒ 本リポの **root 直下** にプラグイン構造を展開（既存の docs/ や .claude/ と並列）
   ⇒ `claude-code/` というサブディレクトリは作らない（リポジトリ自体がプラグインルート）
   ⇒ 生成されるディレクトリ例: `cmd/plugin-sdk/`, `pkg/parser/`, `rules/`,
      `cmd/claude-code-security-logger/`（hook logger は手動追加、要件 §6.1）
       ↓
WF-Phase 2〜6: 本リポ要件 v3 §19 の Phase 0〜6 に従って実装
```

> **注**: 要件 §27.1 で Repository 名が `falco-plugin-claude-code`（ハイフン区切り）と書かれているが、本リポ名は `falco-plugin-claude_code`（アンダースコア）。GitHub 上での `gh repo rename` を v0.1 リリース前に実施するか、現名のまま運用するかは別途判断する。Go module path も同様（`github.com/takaosgb3/falco-plugin-claude_code` か `...claude-code` か）。

#### 8.3.3 ロールバック戦略

| Step | ロールバック手段 |
|---|---|
| Step 1〜5 PR | revert commit でテンプレートを直前状態に戻す |
| dev-kit v0.2.0 タグ後 | 既存プラグインは影響を受けない（新規 scaffold のみ影響）。緊急時は v0.1.x へ降格して generate |
| claude-code scaffold 後 | `claude-code/` ディレクトリを削除して再 scaffold |

### 8.4 運用注意事項

#### 8.4.1 タスク着手時のチェックリスト

各タスクに着手する前に、以下を必ず確認する:

- [ ] 当該 Step（§2〜§6）の **コンテキスト復元** セクションを開く
- [ ] dev-kit v2 要件書 (`dev-kit-v2-requirements.md`) の該当行範囲を読む
- [ ] dev-kit 側の編集対象テンプレートを開く
- [ ] openclaw 側の参照実装を開く（必要に応じて）
- [ ] 本リポ要件 v3 の関連節（§18.4 マッピング表など）を確認
- [ ] 先行タスクの完了を確認（§1.4 依存関係図参照）
- [ ] 当該 Step のブランチ（`feat/v2-stepN`）にチェックアウト

#### 8.4.2 タスク完了時のチェックリスト

- [ ] 完了基準のチェックボックスをすべて埋める
- [ ] go vet / go test / make build を実行（該当 Step のもの）
- [ ] 受入テスト（該当 Step のスコープ）を実施
- [ ] PR を出すか、Step 末尾で一括 PR
- [ ] 本タスク定義書の該当タスクに「実装済」マーク（任意）

#### 8.4.3 コンテキスト消失時の復帰手順

新しい Claude Code セッション、あるいは長時間休憩後の復帰時:

1. **本タスク定義書** `docs/tasks/detailed_task_definition.md` を開く
2. 目次 → §1 を読み直す
3. dev-kit リポジトリで `git status` / `git log -5` を実行し、現在の Step とコミット状況を把握
4. 該当 Step（§2〜§6）を開き、未完了タスクから再開
5. 各タスク内の「コンテキスト復元」セクションを順に開く
6. 不明点があれば本リポ Issue #1 のコメントを遡る

#### 8.4.4 Claude Code 利用時の推奨

- 本タスク定義書の **目次と §1** を最初に Read で読む
- 1 タスクあたり 1 セッションを目安（コンテキスト消費を抑える）
- 大きな構造変更（T4-1, T4-2 など）は **Plan 系エージェント**で設計をまとめてから実装
- テスト追加・修正は **Explore エージェント**で類似パターンを参考

### 8.5 進捗追跡

| 進捗管理 | URL / ファイル |
|---|---|
| 親 Issue（dev-kit 側） | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/1 |
| 進捗報告（本リポ） | https://github.com/takaosgb3/falco-plugin-claude_code/issues/1 |
| Step 別 PR | dev-kit `feat/v2-stepN` |
| 改訂履歴 | dev-kit v2 要件書 §9（v5.6 まで）、本タスク定義書 §改訂履歴 |

### 8.6 受入完了判定

dev-kit v2 が「完了」したと判定する条件:

1. ✅ Step 1〜5 のすべての PR が main にマージ済み
2. ✅ 受入テスト AT-1, AT-2, AT-3, AT-4, AT-5 すべて PASS
3. ✅ 拡張受入テスト ET-1〜ET-7 すべて PASS（ET-7 = 既存プラグイン回帰確認を含む）
4. ✅ dev-kit v0.2.0 タグ作成
5. ✅ 本リポで `/plugin-scaffold claude-code json` が成功（ET-6）
6. ✅ 既存プラグイン（nginx, openclaw）の go vet/test に影響なし
7. ✅ PROBLEM_PATTERNS.md に P001〜P021 + 既存 A コードが整理されている
8. ✅ dev-kit-feedback スキルが動作する
9. ✅ 全スキル `SKILL.md` が最新テンプレートと整合している（要件 §E3）
10. ✅ ドキュメント整合性: 本タスク定義書の参照が要件書・テンプレート・スキルと一致。**要件 §E1〜E8 全件のチェック項目を消化**
11. ✅ **互換性確認（要件 §E4）**: Go 1.22 以上、Falco 0.43 以上、plugin-sdk-go 0.7.4 以上で生成プラグインが動作。CI で minimum-version マトリクス確認
12. ✅ **性能（§E5）**: AT-5 の TC-5-01 で 100 events/sec 通過。NextBatch 間でのメモリ非増加を Level 2 テストで観測
13. ✅ **セキュリティ（§E6）**: T2-1（入力 truncate）、T1-1（nil map 防止）、T5-1（パストラバーサル）の各単体テストで合格
14. ✅ **テンプレート変数仕様（§E7）**: AT-1〜AT-3 の 3 ドメイン生成が成功し、`${PLUGIN_DESCRIPTION}` `${LOG_SOURCE}` `${DOMAIN_FIELDS_*}` がすべて展開できる

完了後は本リポ要件 v3 §19 / Phase 0〜6 に従って claude-code プラグイン実装を進める。

---

## 改訂履歴

| 日付 | バージョン | 概要 |
|------|-----------|------|
| 2026-04-26 | 1.0 | 初版作成（9 ファイル分割版）。dev-kit v2 要件定義書 v5.6 と既存 TASK_DEFINITIONS.md（1762 行）を統合し、claude-code 視点で再構成 |
| 2026-04-26 | 1.1 | レビュー容易性向上のため 9 ファイルを 1 ファイルに統合。内部リンクをアンカー参照に変更 |
| 2026-04-26 | 1.2 | レビュー Round 1〜3 反映（26 件修正）。Pコードマッピング統一（§1.5.5/§7.2）、依存関係補正（T4-2/T5-7/T5-8）、scaffold ファイル一覧 18→20 のフォロー（T5-4/T5-5）、SC コード飛び番注記（§7.7）、互換性受入項目追加（§8.6）、Info() Description 動的化（T4-1）、dev-kit Pコード ≠ 要件 v3 §18.4 解釈の食い違いを明示注記 |
| 2026-04-26 | 1.3 | レビュー Round 4 反映（7 件修正）。受入完了判定に E5/E6/E7 を追加（§8.6 項目 12〜14）、test-plugin 標準化（§1.5.3）、ET-7 既存プラグイン回帰確認（§8.2.2/§8.2.3）、T2-5 完了基準に Pコード食い違い別 Issue 起票、T1-3 TC 表記を要件準拠に正確化、nginx-plugin パス追加 |
| 2026-04-26 | 1.4 | レビュー Round 5 反映（3 件修正）。§8.6 項目 3 を「ET-1〜ET-7」に更新、§1.5.3 で `$STEP` 環境変数の使い方を注記、改訂履歴に v1.4 を記録。**5 ラウンド完了 = 累計 36 件修正、収束達成** |

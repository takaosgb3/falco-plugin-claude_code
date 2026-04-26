# falco-plugin-dev-kit v2 詳細タスク定義書

| 項目 | 内容 |
|------|------|
| 文書ID | TASK-DEVKIT-V2-001 |
| 作成日 | 2026-03-07 |
| 基盤文書 | `docs/requirements/dev-kit-v2-requirements.md` (v5.6) |
| 親 Issue | https://github.com/takaosgb3/falco-plugin-dev-kit/issues/1 |

---

## この文書の目的

要件定義書（`dev-kit-v2-requirements.md` v5.6）の全改善項目を、**Claude Code が単独で実行可能な粒度**の
タスクに分解した定義書です。各タスクには以下を含みます:

- **コンテキスト復元に必要な参照ドキュメント** — 作業途中でコンテキストが失われても、指定ファイルを読めば作業を再開できる
- **具体的な変更手順** — 何を、どのファイルの、どの箇所に変更するか
- **検証手順** — タスク完了の判定方法
- **依存関係** — 先行タスクと後続タスク

---

## タスク体系

```
Step 1: 基盤改善【P0 Critical】（4 タスク）
  T1-1: parseLine() と parser パッケージの接続
  T1-2: Makefile OS/Arch 自動検出
  T1-3: Level 2 パイプラインテストテンプレート作成
  T1-4: plugin-test スキル更新

Step 2: テスト強化・セキュリティ修正【P1 High】（5 タスク）
  T2-1: 入力サイズ超過時の挙動修正（truncate）
  T2-2: JSON パーサーのデフォルト実装
  T2-3: Level 1 E2E パターンテストテンプレート作成
  T2-4: E2E パターン JSON 拡張（benign + edge_cases）
  T2-5: PROBLEM_PATTERNS.md への知見追加（P001-P021）

Step 3: CI/CD・ビルド改善【P1 High】（6 タスク）
  T3-1: Makefile に build-release ターゲット追加
  T3-2: Makefile に E2E テストターゲット追加
  T3-3: CI/CD 3 ワークフロー分離
  T3-4: 3 環境 Falco 設定テンプレート作成
  T3-5: config.go.tmpl ドメイン非依存化
  T3-6: plugin-build スキル更新

Step 4: ドメイン非依存化【P1 High】（5 タスク）
  T4-1: PluginEvent のドメイン非依存化
  T4-2: LogEntry のドメイン非依存化
  T4-3: フォーマット自動検出モード追加
  T4-4: plugin-scaffold スキル更新
  T4-5: plugin-parser スキル更新

Step 5: ドキュメント・仕組み化【P2 Medium】（9 タスク）
  T5-1: ~/パス展開ロジック追加
  T5-2: Extract() 冗長チェック削除
  T5-3: URL デコード重複排除 ※ T1-1 に統合済み (R5-001)
  T5-4: CLAUDE.md テンプレート作成
  T5-5: CHANGELOG.md テンプレート作成
  T5-6: README.md.tmpl 更新
  T5-7: dev-kit-feedback スキル新規作成
  T5-8: plugin-dev-workflow エージェント更新
  T5-9: plugin-rules スキル更新
```

**合計: 29 タスク**（5 Step）

---

## 共通事項

### ブランチ戦略

```
main
  └── feat/v2-step1  ← Step 1 の全タスク
  └── feat/v2-step2  ← Step 2 の全タスク
  └── feat/v2-step3  ← Step 3 の全タスク
  └── feat/v2-step4  ← Step 4 の全タスク
  └── feat/v2-step5  ← Step 5 の全タスク
```

各 Step は PR 単位でマージする。Step 内のタスクは原則 1 コミットずつ。ただし、セットで実装すべきタスク（例: T4-1 と T4-2）は同一コミットとする。

### コンテキスト復元の共通手順

どの Step から作業を再開する場合でも、以下を最初に読むこと:

1. **本文書**（`docs/requirements/TASK_DEFINITIONS.md`）— タスク全体像と現在の進捗
2. **要件定義書**（`docs/requirements/dev-kit-v2-requirements.md`）— 該当 Step のセクション
3. **対象テンプレートファイル** — 該当タスクの「編集対象」に記載のファイル
4. **openclaw の対応ファイル** — 該当タスクの「参照実装」に記載のファイル

### スキル定義内の成功基準（SC コード）

各スキルには成功基準が定義されています。タスク完了時にこれらも参照すること:

| スキル | SC コード範囲 | 主な基準 |
|--------|-------------|---------|
| plugin-scaffold | SC-001〜SC-008 | ファイル生成、go vet パス、ディレクトリ構造 |
| plugin-parser | SC-010〜SC-013 | パーサー実装、セキュリティ検出 |
| plugin-rules | SC-020〜SC-024 | ルール構文、source 指定、URL エンコード対応 |
| plugin-test | SC-030〜SC-033 | テスト通過、E2E パターン 20+ |
| plugin-build | SC-040〜SC-043 | ビルド成功、ELF バイナリ検証 |
| plugin-dev-workflow | SC-050〜SC-053 | 全 Phase 完了、品質ゲート通過 |

### openclaw リポジトリの場所

```
/Users/takaos/lab/falco-plugin-openclaw/
```

### 検証の共通パターン

テンプレート変更後は、テスト用プラグインを生成して以下を確認する:

```bash
# テスト用プラグインのディレクトリで実行
go vet ./...          # 静的解析
go test ./... -v      # テスト実行
make build            # ビルド成功
```

---

## Step 1: 基盤改善【P0 Critical】

**目標**: 生成直後のプラグインが `go vet` + `go test` + `make build` を通過する

**要件定義書の該当セクション**: セクション 7「Step 1」(L1451-L1465)

---

### T1-1: parseLine() と parser パッケージの接続

| 項目 | 内容 |
|------|------|
| 要件ID | A1-1 |
| 優先度 | P0 Critical |
| 依存関係 | なし（最初に着手） |
| 後続タスク | T1-3（Level 2 テストはparser接続が前提）、T4-1（PluginEvent 非依存化） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L93-L209（A1-1 セクション） | Before/After コード例と設計方針 |
| `.claude/templates/plugin/plugin.go.tmpl` | L1-L11（テンプレート変数）、L48-L51（PluginConfig）、L275-L302（parseLine）、L307-L395（Fields/Extract） | 現在のテンプレート構造 |
| `.claude/templates/plugin/parser.go.tmpl` | L1-L30（パッケージ構造）、L60-L95（Parser構造体とNew関数） | parser パッケージの API |
| `.claude/templates/plugin/config.go.tmpl` | 全体（9行） | parser.Config 構造体 |
| `/Users/takaos/lab/falco-plugin-openclaw/cmd/plugin-sdk/plugin.go` | L1-L50（import/構造体）、L100-L150付近（Init）、L200付近（parseLine） | 参照実装 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/plugin.go.tmpl`（456行）

1. **import に parser パッケージを追加** (L5付近)
   ```go
   "github.com/${AUTHOR}/${PLUGIN_NAME}/pkg/parser"
   ```

2. **MyPlugin 構造体に parser フィールドを追加** (L79付近)
   ```go
   parser *parser.Parser
   ```

3. **Init() 内で parser を初期化** (L134付近)
   ```go
   p.parser = parser.New(parser.Config{
       LogFormat:        "${LOG_FORMAT}",
       SecurityPatterns: true,
   })
   ```

4. **Open() で parser を readLoop に渡す** (L213付近)
   ```go
   go instance.readLoop(p.parser)
   ```

5. **readLoop() のシグネチャに parser を追加** (L219付近)
   ```go
   func (inst *MyInstance) readLoop(p *parser.Parser) {
   ```

6. **readNewLines() のシグネチャに parser を追加** (L239付近)
   ```go
   func (inst *MyInstance) readNewLines(path string, p *parser.Parser) {
   ```

7. **parseLine() を実装** (L275-L302) — TODO コメントを削除し、実際の parser 呼び出しに置換
   ```go
   func parseLine(line, path string, p *parser.Parser) *PluginEvent {
       entry, err := p.Parse(line)
       if err != nil {
           debugLog("Parse error: %v (line: %.100s)", err, line)
           return nil
       }
       event := &PluginEvent{
           // 共通フィールド
           LogPath:   path,
           Raw:       entry.Raw,
           Timestamp: entry.Timestamp,
           Headers:   entry.Headers,              // P004: parser 側で初期化済み
           // HTTP ドメイン固有フィールドの完全マッピング
           RemoteAddr:  entry.RemoteAddr,
           RemoteUser:  entry.RemoteUser,
           TimeLocal:   entry.TimeLocal.Format("${TIME_FORMAT}"),
           Method:      entry.Method,
           Path:        entry.Path,
           QueryString: entry.QueryString,
           Protocol:    entry.HTTPVersion,         // フィールド名の差異に注意
           Status:      uint64(entry.Status),      // int → uint64 型変換
           BytesSent:   uint64(entry.BodyBytes),   // int → uint64 型変換
           Referer:     entry.Referer,
           UserAgent:   entry.UserAgent,
       }
       return event
   }
   ```

#### 注意事項

- この段階（Step 1）では **ドメイン非依存化は行わない**。HTTP フィールドのまま parser を接続する
- ドメイン非依存化は Step 4（T4-1, T4-2）で実施。Step 4 では T1-1 で追加した parseLine() 内の HTTP 固有マッピング（`TimeLocal.Format("${TIME_FORMAT}")` 等）が `${DOMAIN_FIELDS_MAPPING}` プレースホルダーに置換される（R5-003 対応）
- `${LOG_FORMAT}` テンプレート変数は scaffold スキルの Phase 1 で収集済み（E7 参照、要件定義書 L1393）

8. **URL デコード重複排除を同時実施**（R5-001 対応: T5-3 から前倒し）
   `regex_simple.go.tmpl` の `DetectSecurityThreat()` 内（L34-L42）の URL デコード処理を削除する。
   parser 接続後は `detectSecurityPatterns()` 内でデコード済み文字列が `DetectSecurityThreat()` に渡されるため、
   `DetectSecurityThreat()` 内の URL デコードは二重実行となり最大6段階のデコードが発生する。

#### 検証手順

```bash
# plugin.go.tmpl の変数を手動展開してテスト用ファイルを作成
# または scaffold スキルでテスト用プラグインを生成
go vet ./...          # parser import が正しく解決される
go build ./...        # コンパイルが通る
```

#### 完了条件

- [ ] plugin.go.tmpl の parseLine() が parser.Parse() を呼び出している
- [ ] MyPlugin 構造体に `parser *parser.Parser` フィールドがある
- [ ] Init() 内で parser.New() による初期化がある
- [ ] readLoop/readNewLines に parser が引数として渡されている
- [ ] TODO コメントが削除されている
- [ ] regex_simple.go.tmpl の DetectSecurityThreat() 内の URL デコード処理（L34-L42）が削除されている（R5-001）

---

### T1-2: Makefile OS/Arch 自動検出

| 項目 | 内容 |
|------|------|
| 要件ID | A4-1 |
| 優先度 | P0 Critical |
| 依存関係 | なし |
| 後続タスク | T3-1（build-release）、T3-6（build スキル更新）、T5-6（README 更新） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L649-L694（A4-1 セクション） | Before/After コード例 |
| `.claude/templates/plugin/Makefile.tmpl` | 全体（60行） | 現在の Makefile テンプレート |
| `/Users/takaos/lab/falco-plugin-openclaw/Makefile` | L6-L19（OS自動検出）、全体（151行） | 参照実装: `uname -s` + `uname -m` による Darwin/Linux/arm64/amd64 判定 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/Makefile.tmpl`（60行）

1. **L1-L5 を OS 自動検出ロジックに置換**:
   - `uname -s` で Darwin/Linux を判定
   - `uname -m` で arm64/amd64 を判定
   - Darwin: `.dylib` バイナリ、Linux: `.so` バイナリ
   - Go 環境変数（GOOS, GOARCH）を自動設定

2. **`GO_RELEASE_FLAGS` 変数を追加**（Step 3 の T3-1 で使用）:
   ```makefile
   GO_RELEASE_FLAGS := -buildmode=c-shared -trimpath -ldflags="-s -w"
   ```

#### 検証手順

```bash
# macOS で確認
make build   # → .dylib が生成される
# Linux で確認（CI で検証）
make build   # → .so が生成される
```

#### 完了条件

- [ ] `UNAME_S` / `UNAME_M` による OS/Arch 判定がある
- [ ] macOS (arm64/amd64) → `.dylib`、Linux → `.so` の分岐がある
- [ ] `GOOS`, `GOARCH` が自動設定される
- [ ] `GO_RELEASE_FLAGS` が定義されている

#### 注意事項

- Linux arm64 は現時点で対象外（E4 互換性要件: Linux amd64, macOS arm64/amd64）

---

### T1-3: Level 2 パイプラインテストテンプレート作成

| 項目 | 内容 |
|------|------|
| 要件ID | A7-1 |
| 優先度 | P0 Critical |
| 依存関係 | T1-1（parser 接続完了が前提） |
| 後続タスク | T1-4（test スキル更新）、T3-2（Makefile E2E ターゲット）、T5-6（README 更新） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L888-L942（A7-1 セクション） | テストケース一覧とヘルパー関数仕様 |
| `.claude/templates/plugin/plugin.go.tmpl` | 全体 | テスト対象のプラグイン構造の理解 |
| `.claude/templates/plugin/parser_test.go.tmpl` | 全体（248行） | 既存テストテンプレートのスタイル参照 |
| `/Users/takaos/lab/falco-plugin-openclaw/cmd/plugin-sdk/plugin_test.go` | 全体（1015行） | **最重要参照**: 37 テスト関数 + 6 ヘルパー関数の実装例 |

#### 変更内容

**新規作成**: `.claude/templates/plugin/plugin_test.go.tmpl`

要件定義書 L901-L938（A7-1 テストケーステーブルとヘルパー関数仕様）に定義された以下を含むテストテンプレートを新規作成する:

**ヘルパー関数**（6 関数）:
1. `initPlugin(t, logPaths)` — プラグイン初期化
2. `openAndCleanup(t, plugin)` — Open + Cleanup 登録
   注: `Open()` は SDK インターフェース `source.Instance` を返す。テストコードで `eventCh` にアクセスするには `instance.(*MyInstance)` への型アサーションが必要。`openAndCleanup` ヘルパー内で `inst := instance.(*MyInstance)` のキャストを行い、`*MyInstance` を返す設計とする。
3. `writeToLog(t, path, line)` — ログ書き込み + fsnotify 待機
4. `waitForEvent(t, ch, timeout)` — イベント待機
5. `gobEncode(t, event)` — GOB エンコード
6. `gobDecode(t, data)` — GOB デコード

**テストケース**（最低 14 TC）:

| カテゴリ | TC ID | テスト内容 |
|---------|-------|-----------|
| ライフサイクル | TC-1-01 | デフォルト設定での Init |
| ライフサイクル | TC-1-02 | カスタム設定での Init |
| ライフサイクル | TC-1-03 | バッファサイズ境界値 |
| ライフサイクル | TC-1-06 | Open() ファイル自動作成 |
| ライフサイクル | TC-1-07 | Open() SeekEnd (P014) |
| ライフサイクル | TC-1-08 | Close() リソース解放 |
| 取り込み | TC-2-01 | 基本ログ取り込み |
| 取り込み | TC-2-04 | 複数ファイル監視 |
| 取り込み | TC-2-05 | GOB ラウンドトリップ |
| 取り込み | TC-2-06 | Headers 非 nil (P004) |
| 性能 | TC-5-01 | スループット 100 events/sec |
| 性能 | TC-5-02 | バッファオーバーフロー |
| エラー耐性 | TC-6-01 | 不正 JSON 設定 |
| エラー耐性 | TC-6-04 | ファイル削除時 |

#### 注意事項

- テスト内の `time.Sleep` に P021（fsnotify タイミング）の注記コメントを入れること
- `Headers` マップの非 nil チェック（P004）を TC-2-06 で必ずテストすること
- openclaw の `plugin_test.go`（37 テスト関数）を **構造のリファレンス** として使うが、
  HTTP 固有のフィールドはテンプレート変数に合わせる（この段階では HTTP デフォルト）
- openclaw のテスト関数名一覧（参考）:
  TestPipelineInitDefault, TestPipelineInitCustom, TestPipelineInitBufferBoundary,
  TestPipelineInfo, TestPipelineInitSchema, TestPipelineOpenFileCreation,
  TestPipelineOpenSeekEnd, TestPipelineClose, TestPipelineFieldsExtractConsistency,
  TestPipelineJSONIngestion, TestPipelinePlaintextIngestion, TestPipelineMultiLineWrite,
  TestPipelineMultiFileWatch, TestPipelineGOBRoundTrip, TestPipelineHeadersNonNil,
  TestPipelineHeadersCopy, TestPipelineHeadersFieldExists, TestPipelineTimestampFormats,
  TestPipelineSourceFile, TestPipelineEmptyLineSkip, TestPipelineFsnotifyWrite,
  TestPipelineThroughput, TestPipelineBufferOverflow, TestPipelineDropCount,
  TestPipelineLargeInput, TestPipelineLongRunning, TestPipelineInvalidJSON,
  TestPipelineEmptyLine, TestPipelineSuperLongLine, TestPipelineFileDeleted,
  TestPipelineFileRecreated, TestPipelinePermissionError, TestPipelineInitInvalidConfig,
  TestPipelineGOBDecodeError, TestPipelineUnknownField, TestPipelineBinaryData,
  TestPipelineNonTargetFile

#### 検証手順

```bash
go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -count=1 -timeout 120s
```

#### 完了条件

- [ ] `.claude/templates/plugin/plugin_test.go.tmpl` が新規作成されている
- [ ] 6 ヘルパー関数が定義されている
- [ ] 最低 14 テストケースが含まれている
- [ ] テンプレート変数（`${PLUGIN_NAME}` 等）が正しく使われている
- [ ] P004, P014, P021 の知見がテストに反映されている
- [ ] テスト関数名が `TestPipeline` プレフィックスで始まること（T3-2 の `make e2e-pipeline` が `-run TestPipeline` でフィルタするため）

---

### T1-4: plugin-test スキル更新

| 項目 | 内容 |
|------|------|
| 要件ID | B4 |
| 優先度 | P0 Critical |
| 依存関係 | T1-3（テストテンプレートの作成完了が前提） |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1217-L1248（B4 セクション） | 変更内容と 3 層テストアーキテクチャ |
| `.claude/skills/plugin-test/SKILL.md` | 全体（247行） | 現在のスキル定義 |

#### 変更内容

**編集対象**: `.claude/skills/plugin-test/SKILL.md`（247行）

1. **3 層 E2E テストアーキテクチャの説明を追加**:
   ```
   Level 1: パターンカバレッジテスト（Falco 不要）
   Level 2: プラグインパイプラインテスト（Falco 不要、CGO_ENABLED=1 必要）
   Level 3: Falco 統合テスト（Falco 必要）
   ```

2. **Level 2 テスト生成手順を追加**:
   - `plugin_test.go.tmpl` からの生成手順
   - テストヘルパー関数の説明
   - 実行コマンド: `make e2e-pipeline`

3. **Level 1 テスト生成手順を追加**:
   - `e2e_pattern_test.go.tmpl` からの生成手順
   - パターン JSON の配置方法
   - 実行コマンド: `make e2e-pattern`

4. **偽陽性テストの手順を追加**:
   - `benign.json` パターンの作成方法

5. **成功基準を更新**:
   - ユニットテスト通過 → Level 1 + Level 2 テスト通過

#### 完了条件

- [ ] 3 層テストアーキテクチャの説明がある
- [ ] Level 2 テスト生成手順がある
- [ ] 成功基準に Level 2 テスト通過が含まれている

---

### Step 1 統合検証

Step 1 の全タスク（T1-1〜T1-4）完了後に実施:

```bash
# 0. テンプレートの手動展開（scaffold スキルに依存しない検証経路、R-315 対応）
#    sed や envsubst 等でテンプレート変数を展開してテスト用プラグインを作成
#    または scaffold スキルでテスト用プラグインを生成

# 1. E4 互換性検証（R-307 対応）
grep "^go " go.mod              # Go 1.22 以上
grep "plugin-sdk-go" go.mod     # v0.7.4 以上

# 2. 生成されたプラグインディレクトリで以下を実行
go vet ./...                    # 静的解析パス
go test ./... -v                # 全テスト通過（parser + pipeline）
make build                      # macOS/Linux でビルド成功
```

---

## Step 2: テスト強化・セキュリティ修正【P1 High】

**目標**: セキュリティ検出の修正と E2E テスト基盤の整備

**要件定義書の該当セクション**: セクション 7「Step 2」(L1466-L1480)

---

### T2-1: 入力サイズ超過時の挙動修正（truncate）

| 項目 | 内容 |
|------|------|
| 要件ID | A3-1 |
| 優先度 | P1 High |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L587-L626（A3-1 セクション） | Before/After と設計意図 |
| `.claude/templates/plugin/regex_simple.go.tmpl` | L27-L29（DetectSecurityThreat 冒頭） | 変更箇所 |
| `/Users/takaos/lab/falco-plugin-openclaw/pkg/parser/regex_simple.go` | L25-L35付近 | 参照実装（truncate 方式） |

#### 変更内容

**編集対象**: `.claude/templates/plugin/regex_simple.go.tmpl`（226行）

**L27-L29 の変更**:

```go
// Before:
if len(input) > d.maxInputLength {
    return "", false   // ← スキップ（セキュリティリスク）
}

// After:
if len(input) > d.maxInputLength {
    input = input[:d.maxInputLength]  // ← 切り詰めて続行
}
```

#### 検証手順

- 10KB 超の入力で `DetectSecurityThreat()` が呼ばれた場合、先頭 10KB 内の脅威パターンが検出されること
- 10KB 以下の入力では動作が変わらないこと

#### 完了条件

- [ ] `return "", false` が `input = input[:d.maxInputLength]` に変更されている

---

### T2-2: JSON パーサーのデフォルト実装

| 項目 | 内容 |
|------|------|
| 要件ID | A2-1 |
| 優先度 | P1 High |
| 依存関係 | なし |
| 後続タスク | T4-3（auto 検出モード）、T4-5（parser スキル更新） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L368-L435（A2-1 セクション） | After コード例と parseTimestamp 実装 |
| `.claude/templates/plugin/parser.go.tmpl` | L197-L201（parseJSON）、L60-L95（Parser/New） | 現在の未実装箇所 |
| `/Users/takaos/lab/falco-plugin-openclaw/pkg/parser/parser.go` | 全体（284行） | 参照実装の JSON パーサー |

#### 変更内容

**編集対象**: `.claude/templates/plugin/parser.go.tmpl`（278行）

1. **L197-L201 の `parseJSON()` を実装**:
   - `json.Unmarshal` で生の JSON をパース
   - 共通フィールド（timestamp, level, message）の汎用マッピング
   - ドメイン固有フィールドの設定箇所に `${DOMAIN_FIELDS_PARSE_JSON}` プレースホルダーを配置（T4-2 でテンプレート展開対応時に使用）
   - `Headers` マップへの残余フィールドの格納

2. **`parseTimestamp()` ヘルパー関数を追加**:
   - RFC3339, RFC3339Nano, ISO 8601 等の複数フォーマットを試行
   - パース失敗時は `time.Now()` にフォールバック

3. **import に `"encoding/json"` を追加**（未追加の場合）

#### 注意事項

- Step 2 時点では HTTP 固有フィールドのマッピングも parseJSON() 内に直接記述する（例: `if v, ok := raw["remote_addr"].(string); ok { entry.RemoteAddr = v }`）。Step 4（T4-2）でこれらが `${DOMAIN_FIELDS_PARSE_JSON}` プレースホルダーに置換される
- `${DOMAIN_FIELDS_PARSE_JSON}` プレースホルダーは Step 2 時点ではコメントとして記述し、Go コンパイルに影響しない形にする（例: `// ${DOMAIN_FIELDS_PARSE_JSON}`）

#### 完了条件

- [ ] `parseJSON()` が JSON 文字列を `LogEntry` に変換できる
- [ ] `parseTimestamp()` が複数の日時フォーマットに対応している
- [ ] "JSON format not yet implemented" エラーが削除されている

---

### T2-3: Level 1 E2E パターンテストテンプレート作成

| 項目 | 内容 |
|------|------|
| 要件ID | A7-2 |
| 優先度 | P1 High |
| 依存関係 | T2-4（benign/edge_cases パターンが TC-3-02, TC-3-04 のテストデータとして必要） |
| 後続タスク | T3-2（Makefile E2E ターゲット）、T5-6（README 更新） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L944-L962（A7-2 セクション） | テストケース一覧 |
| `.claude/templates/plugin/e2e_pattern.json.tmpl` | 全体（213行） | パターン JSON のスキーマ理解 |
| `/Users/takaos/lab/falco-plugin-openclaw/test/e2e/e2e_pattern_test.go` | 全体（462行） | **最重要参照**: E2E テストの実装例 |
| `/Users/takaos/lab/falco-plugin-openclaw/test/e2e/patterns/categories/` | 11 JSON ファイル（56 パターン） | パターン JSON の実例（benign.json, edge_cases.json 含む） |

#### 変更内容

**新規作成**: `.claude/templates/plugin/e2e_pattern_test.go.tmpl`

- パターン JSON ファイルをディレクトリ走査で動的読み込み
- 各パターンをパーサーに通して検出結果を検証
- TC-3-01: True Positive テスト
- TC-3-02: True Negative テスト
- TC-3-04: 10KB 入力サイズ境界テスト
- TC-3-05: 大文字小文字非依存テスト

#### 注意事項

- e2e_pattern.json.tmpl は生成時にカテゴリごとに `test/e2e/patterns/categories/` ディレクトリに分割配置される（openclaw と同じ構造）。テストはこのディレクトリを走査して `*.json` を動的読み込みする
- テスト関数名は `TestPattern` プレフィックスで始めること（T3-2 の `make e2e-pattern` が `-run TestPattern` でフィルタするため）

#### 完了条件

- [ ] `.claude/templates/plugin/e2e_pattern_test.go.tmpl` が新規作成されている
- [ ] ディレクトリ走査によるパターン動的読み込みがある
- [ ] 4 種類のテストケースが含まれている

---

### T2-4: E2E パターン JSON 拡張

| 項目 | 内容 |
|------|------|
| 要件ID | A7-3 |
| 優先度 | P1 High |
| 依存関係 | なし |
| 後続タスク | T2-3（Level 1 テストが benign/edge_cases パターンを使用） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L964-L992（A7-3 セクション） | パターン追加の方針と JSON スキーマ拡張 |
| `.claude/templates/plugin/e2e_pattern.json.tmpl` | 全体（213行） | 現在のパターン構造 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/e2e_pattern.json.tmpl`（213行）

1. **benign カテゴリを追加**（正常リクエスト 5+ パターン）:
   - 正常な HTTP リクエスト
   - 正常な JSON ログ
   - 安全な URL パラメータ
   → 偽陽性テスト用

2. **edge_cases カテゴリを追加**（境界値パターン）:
   - 10239 バイト（10KB 未満）
   - 10240 バイト（ちょうど 10KB）
   - 10241 バイト（10KB 超）
   - 空文字列
   - 空白のみ

**参考**: openclaw の E2E パターンは 11 カテゴリ / 56 パターン:
shell_injection, data_exfiltration, workspace_escape, edge_cases,
unauthorized_model, composite, plaintext_threats, dangerous_command,
suspicious_config, benign, agent_runaway

3. **JSON スキーマに新フィールドを追加**:
   - `format`: ログ形式（json / plaintext / combined）
   - `expected_threat`: 期待脅威タイプ
   - `note`: テスト作成者向け注記

#### 完了条件

- [ ] benign カテゴリに 5+ 正常パターンがある
- [ ] edge_cases カテゴリに 5+ 境界値パターンがある
- [ ] JSON スキーマに `format`, `expected_threat`, `note` フィールドがある

---

### T2-5: PROBLEM_PATTERNS.md への知見追加

| 項目 | 内容 |
|------|------|
| 要件ID | C2 |
| 優先度 | P1 High |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1292-L1338（C2 セクション） | P001-P021 の一覧と集約方針 |
| `PROBLEM_PATTERNS.md` | 先頭 50 行（構造の理解）、末尾（追加位置の特定） | 現在のパターン集 |
| `.claude/skills/plugin-build/SKILL.md` | P001, P002, P013 の記述箇所 | 散在する P コードの原文確認 |
| `.claude/skills/plugin-scaffold/SKILL.md` | P003, P004, P007-P010, P014 の記述箇所 | 散在する P コードの原文確認 |
| `.claude/skills/plugin-rules/SKILL.md` | P003, P005, P006, P011, P012, P015, P016 の記述箇所 | 散在する P コードの原文確認 |
| `.claude/skills/plugin-parser/SKILL.md` | P004, P006 の記述箇所 | 散在する P コードの原文確認 |

#### 変更内容

**編集対象**: `PROBLEM_PATTERNS.md`（19,676行）

1. **「P コード: プラグイン共通パターン」セクションを新設**
   ファイル先頭に新規セクションヘッダー `# P コード: プラグイン共通パターン` を追加し、
   既存の A-code パターン（`## Pattern #A334` で始まる）の前に配置する

2. **P001-P016 をスキル定義から集約**:
   各スキルファイルに散在する P コードの知見を統一フォーマットで記録。
   要件定義書 L1307-L1326 の表に従う。

3. **P017-P021 を新規追加**:
   openclaw 開発で発見された新知見。
   要件定義書 L1328-L1336 の表に従う。

#### 注意事項

- 既存の A-code パターン（A303-A334）は変更しない
- P コードセクションは A-code セクションの前に配置する
- 各 P コードには「参照元スキル」を記載し、トレーサビリティを確保する

#### 完了条件

- [ ] P001-P016 が PROBLEM_PATTERNS.md に記録されている
- [ ] P017-P021 が PROBLEM_PATTERNS.md に記録されている
- [ ] 各 P コードに参照元スキル名が記載されている
- [ ] 既存 A-code パターンは変更されていない

---

### Step 2 統合検証

生成されたプラグインディレクトリで以下を実行:

```bash
go vet ./...                        # 静的解析パス
go test ./... -v                    # 全テスト通過
go test ./test/e2e/ -v -race -run TestPattern -count=1  # Level 1 テスト通過（直接実行）
# make e2e-pattern は T3-2 完了後に利用可能
# 10KB 超入力で検出が動作することを確認
```

---

## Step 3: CI/CD・ビルド改善【P1 High】

**目標**: ビルド・テスト・CI/CD パイプラインの本格整備

**要件定義書の該当セクション**: セクション 7「Step 3」(L1481-L1497)

---

### T3-1: Makefile に build-release ターゲット追加

| 項目 | 内容 |
|------|------|
| 要件ID | A4-2 |
| 優先度 | P1 High |
| 依存関係 | T1-2（OS 自動検出 + GO_RELEASE_FLAGS 定義済み） |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L696-L711（A4-2 セクション） | After コード例 |
| `.claude/templates/plugin/Makefile.tmpl` | 全体（T1-2 で変更済み） | 現在の Makefile 構造 |
| `/Users/takaos/lab/falco-plugin-openclaw/Makefile` | build-release ターゲット | 参照実装 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/Makefile.tmpl`

`build-release` ターゲットを追加:
```makefile
build-release:
	$(GO_ENV) go build $(GO_RELEASE_FLAGS) -o $(BINARY) $(SRC_DIR)/
```

#### 完了条件

- [ ] `make build-release` でサイズ最適化ビルドが成功する

---

### T3-2: Makefile に E2E テストターゲット追加

| 項目 | 内容 |
|------|------|
| 要件ID | A4-3 |
| 優先度 | P1 High |
| 依存関係 | T1-3（Level 2 テスト）、T2-3（Level 1 テスト） |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L713-L738（A4-3 セクション） | ターゲット定義 |
| `/Users/takaos/lab/falco-plugin-openclaw/Makefile` | e2e 関連ターゲット | 参照実装 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/Makefile.tmpl`

追加するターゲット:
```makefile
.PHONY: e2e-pattern e2e-pipeline e2e vet

e2e-pattern:
	go test ./test/e2e/ -v -race -run TestPattern -count=1

e2e-pipeline:
	go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -count=1 -timeout 120s

e2e: e2e-pattern e2e-pipeline

vet:
	go vet ./...
```

#### 完了条件

- [ ] `make e2e-pattern`, `make e2e-pipeline`, `make e2e`, `make vet` ターゲットが存在する

---

### T3-3: CI/CD 3 ワークフロー分離

| 項目 | 内容 |
|------|------|
| 要件ID | A5-1 |
| 優先度 | P1 High |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L744-L855（A5-1 セクション） | 3 ワークフローの設計 |
| `.claude/templates/plugin/ci.yml.tmpl` | 全体（84行） | 現在の CI テンプレート |
| `/Users/takaos/lab/falco-plugin-openclaw/.github/workflows/ci.yml` | 全体（54行） | 参照: CI（push/PR to main、test+build 2ジョブ） |
| `/Users/takaos/lab/falco-plugin-openclaw/.github/workflows/e2e-test.yml` | 全体（334行） | 参照: E2E（push/PR+path filter+workflow_dispatch、go-tests+falco-integration+allure 4ジョブ） |
| `/Users/takaos/lab/falco-plugin-openclaw/.github/workflows/release.yml` | 全体（91行） | 参照: Release（workflow_dispatch、validate+matrix build+release 3ジョブ） |

#### 変更内容

1. **編集**: `.claude/templates/plugin/ci.yml.tmpl`（84行）
   - release ジョブ（L59-L85）を削除
   - ランナーを `ubuntu-24.04` にピン留め
   - テストに `-race` フラグ追加
   - golangci-lint バージョンをピン留め
   - YAML ルール検証ステップ追加

2. **新規作成**: `.claude/templates/plugin/e2e-test.yml.tmpl`
   - Level 1 + Level 2 テスト実行
   - （オプション）Allure レポート生成

3. **新規作成**: `.claude/templates/plugin/release.yml.tmpl`
   - `workflow_dispatch` トリガー
   - matrix ビルド: ubuntu-24.04 (.so) + macos-14 (.dylib)
   - SHA256 チェックサム付き GitHub Release

#### 注意事項

- golangci-lint のピン留めバージョンは実装時点の最新安定版を使用する。openclaw の `.github/workflows/ci.yml` を参照すること

#### 完了条件

- [ ] `ci.yml.tmpl` から release ジョブが分離されている
- [ ] `e2e-test.yml.tmpl` が新規作成されている
- [ ] `release.yml.tmpl` が新規作成されている
- [ ] 3 ファイルのテンプレート変数が正しく使われている

---

### T3-4: 3 環境 Falco 設定テンプレート作成

| 項目 | 内容 |
|------|------|
| 要件ID | A6-1 |
| 優先度 | P1 High |
| 依存関係 | なし |
| 後続タスク | T5-6（README 更新） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L861-L882（A6-1 セクション） | 3 環境の設計と P007-P009 の必須設定 |
| `.claude/templates/plugin/falco.yaml.tmpl` | 全体（23行） | 現在の設定テンプレート |
| `/Users/takaos/lab/falco-plugin-openclaw/falco.yaml` | 全体（22行） | 参照: Linux 本番 |
| `/Users/takaos/lab/falco-plugin-openclaw/falco-local.yaml` | 全体（18行） | 参照: macOS ローカル |
| `/Users/takaos/lab/falco-plugin-openclaw/falco-docker.yaml` | 全体（31行） | 参照: Docker |

#### 変更内容

1. **既存の確認**: `.claude/templates/plugin/falco.yaml.tmpl` — 必要に応じて P007-P009 の設定を追加

2. **新規作成**: `.claude/templates/plugin/falco-local.yaml.tmpl`
   - バイナリパス: `./lib${PLUGIN_NAME}-plugin-darwin-arm64.dylib`
   - `outputs:` セクション除外（P017）
   - `-U` フラグ注記（P018）

3. **新規作成**: `.claude/templates/plugin/falco-docker.yaml.tmpl`
   - バイナリパス: `/plugins/lib${PLUGIN_NAME}-plugin.so`
   - `json_output: true`

#### 全設定ファイル共通の必須設定（P007-P009）

```yaml
load_plugins: [${PLUGIN_NAME}]      # P008
rate: 0                              # P009
max_burst: 0                         # P009
rules_files:
  - /path/to/${PLUGIN_NAME}_rules.yaml  # P007
```

#### 完了条件

- [ ] `falco-local.yaml.tmpl` が新規作成されている
- [ ] `falco-docker.yaml.tmpl` が新規作成されている
- [ ] 3 ファイル全てに `load_plugins`, `rate: 0`, `max_burst: 0` がある
- [ ] `falco-local.yaml.tmpl` に `outputs:` が含まれていない

---

### T3-5: config.go.tmpl ドメイン非依存化

| 項目 | 内容 |
|------|------|
| 要件ID | A9-1 |
| 優先度 | P1 High |
| 依存関係 | なし |
| 後続タスク | T4-1（ドメイン非依存化の一部として） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1058-L1121（A9-1 セクション） | After コードと 2 つの Config の関係 |
| `.claude/templates/plugin/config.go.tmpl` | 全体（9行） | 現在の Config |
| `.claude/templates/plugin/parser.go.tmpl` | L60-L108（New() 関数） | MaxFieldLength の読み取り追加箇所 |
| `.claude/templates/plugin/regex_simple.go.tmpl` | L8（maxInputSize 定数）、L17付近（コンストラクタ） | 定数削除とコンストラクタ修正箇所 |
| `/Users/takaos/lab/falco-plugin-openclaw/pkg/parser/config.go` | 全体（8行） | 参照実装 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/config.go.tmpl`（9行）

```go
// Before:
type Config struct {
    LogFormat              string
    CustomFormat           string
    SecurityPatterns       bool
    LargeResponseThreshold int
}

// After:
type Config struct {
    LogFormat          string // "auto", "combined", "common", "json", "custom"
    CustomFormat       string // Custom regex pattern
    SecurityPatterns   bool   // Enable security threat detection
    MaxFieldLength     int    // Threshold for large field truncation (bytes)
}
```

変更点:
- `LogFormat` コメントに `"auto"` を追加
- `LargeResponseThreshold` → `MaxFieldLength`（ドメイン非依存名）

3. **MaxFieldLength の配線を実装**（R-302 対応）:
   現在 `LargeResponseThreshold` は config.go.tmpl にのみ存在し、どこからも参照されていない。
   `regex_simple.go.tmpl` は独自の `maxInputSize` 定数（L8: `const maxInputSize = 10 * 1024`）を使用している。

   **parser.go.tmpl の `New()` 関数に追加**:
   ```go
   // MaxFieldLength のフォールバック処理
   maxFieldLen := cfg.MaxFieldLength
   if maxFieldLen <= 0 {
       maxFieldLen = 10 * 1024 // デフォルト 10KB
   }
   ```

   **SimpleSecurityDetector の初期化に渡す**:
   ```go
   detector := NewSimpleSecurityDetector(maxFieldLen)
   ```

   **regex_simple.go.tmpl の変更**:
   - `const maxInputSize = 10 * 1024` を削除
   - `NewSimpleSecurityDetector(maxInputLength int)` コンストラクタを追加（または既存の初期化を修正）
   - `d.maxInputLength` を渡された値で初期化

4. **parser_test.go.tmpl の更新**:
   `NewSimpleSecurityDetector()` の呼び出し箇所（6箇所: L85, L106, L128, L150, L172, L197）を
   `NewSimpleSecurityDetector(10 * 1024)` に変更する

#### 注意事項

- `regex_simple.go.tmpl` 内で `LargeResponseThreshold` を参照している箇所があれば `MaxFieldLength` に更新
- `parser.go.tmpl` 内の参照も更新
- **重要**: 単なるリネームではなく、config → parser → detector の配線を実装すること
- T1-1 で `DetectSecurityThreat()` 内の URL デコード処理（L34-L42）が削除済みの前提で作業する。regex_simple.go.tmpl の構造が T1-1 実施前と異なることに注意

#### 完了条件

- [ ] `LargeResponseThreshold` が `MaxFieldLength` にリネームされている
- [ ] `LogFormat` のコメントに `"auto"` が含まれている
- [ ] 他テンプレートからの参照が更新されている
- [ ] `parser.go.tmpl` の `New()` が `cfg.MaxFieldLength` を読み取っている
- [ ] `regex_simple.go.tmpl` のハードコード定数 `maxInputSize` が削除されている
- [ ] `SimpleSecurityDetector` が外部から `maxInputLength` を受け取るようになっている
- [ ] parser_test.go.tmpl 内の NewSimpleSecurityDetector() 呼び出しが全て更新されている

---

### T3-6: plugin-build スキル更新

| 項目 | 内容 |
|------|------|
| 要件ID | B5 |
| 優先度 | P1 High |
| 依存関係 | T1-2, T3-1 の完了 |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1249-L1255（B5 セクション） | 変更内容 |
| `.claude/skills/plugin-build/SKILL.md` | 全体（259行） | 現在のスキル定義 |

#### 変更内容

**編集対象**: `.claude/skills/plugin-build/SKILL.md`（259行）

1. macOS ネイティブビルド（`.dylib`）が可能であることを明記
2. `build-release` ターゲットの説明を追加
3. P017（macOS outputs 拒否）、P018（-U フラグ）の注意事項を追加

#### 完了条件

- [ ] macOS ビルドの手順が記載されている
- [ ] `build-release` の説明がある
- [ ] P017, P018 が記載されている

---

## Step 4: ドメイン非依存化【P1 High】

**目標**: HTTP 以外のログソース（AI, IoT 等）にも対応できるテンプレートにする

**要件定義書の該当セクション**: セクション 7「Step 4」(L1498-L1511)

### Step 4 前提: テンプレート展開メカニズムの明確化 (R5-002)

Step 4 では `${DOMAIN_FIELDS_STRUCT}` 等 5 つのプレースホルダーによる**複数行 Go コードブロック**の
動的生成が必要です。現在のテンプレートシステムは `${VARIABLE}` 形式の単純な文字列置換ですが、
複数行コードブロック（構造体フィールド定義、switch/case 文等）のインデント整合性維持に注意が必要です。

**テンプレート展開の実行方式**:

本ツールキットでは、テンプレート展開は **Claude Code が scaffold スキル実行時に直接コードを生成** する方式で行います。
`${DOMAIN_FIELDS_*}` プレースホルダーは、sed や envsubst による機械的な文字列置換ではなく、
**Claude Code がフィールド定義に基づいてコードブロックを生成し、テンプレートの該当箇所に挿入する際の
位置と生成内容の指示** として機能します。

この方式により:
- 複数行コードブロックのインデント整合性は Claude Code が保証する
- 型に応じた処理分岐（`.(string)`, `.(float64)` 等）も Claude Code が生成する
- Go の構文正確性は `go vet ./...` で検証する

**Step 4 着手前の確認事項**:
- [ ] scaffold スキル (SKILL.md) の Phase 2 に上記の展開方式を明記（T4-4 で実施）
- [ ] HTTP ドメインでの PoC テスト: 既存テンプレートの HTTP フィールドをプレースホルダーに置き換え、scaffold スキルで展開して `go vet` がパスすることを確認

---

### T4-1: PluginEvent のドメイン非依存化

| 項目 | 内容 |
|------|------|
| 要件ID | A1-2 |
| 優先度 | P1 High |
| 依存関係 | T1-1（parser 接続）、T3-5（config 非依存化）が完了していること |
| 後続タスク | T4-2 とセットで実装、T4-4（scaffold スキル更新） |

**注**: T4-1 と T4-2 は同一コミットとして実装すること。PluginEvent と LogEntry のフィールドは parseLine() でマッピングされているため、片方だけの変更では `go vet` がパスしない。

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L211-L296（A1-2 セクション）、Fields/Extract 方針・テンプレート展開機構 | 設計方針と選択肢 |
| `.claude/templates/plugin/plugin.go.tmpl` | L58-L74（PluginEvent）、L307-L327（Fields）、L332-L395（Extract） | 現在の HTTP 固有構造 |
| `/Users/takaos/lab/falco-plugin-openclaw/cmd/plugin-sdk/plugin.go` | PluginEvent 構造体、Fields()、Extract() | 参照実装（AI ドメイン） |

#### 変更内容

**編集対象**: `.claude/templates/plugin/plugin.go.tmpl`

1. **PluginEvent 構造体を 2 層構造に変更**:
   - 共通フィールド（固定）: `Timestamp`, `LogPath`, `Raw`, `Headers`
   - ドメイン固有フィールド: テンプレート変数で動的生成（具体的な型付きフィールド方式を採用）

2. **Fields() を動的生成に変更**:
   - 共通フィールド（4 個）は固定
   - ドメイン固有フィールドはテンプレート展開時に追加

3. **Extract() を動的生成に変更**:
   - 共通フィールドの case 文は固定
   - ドメイン固有フィールドの case 文はテンプレート展開時に追加

4. **parseLine() のマッピングを更新**（T1-1 で追加した部分を非依存化）
   - T1-1 で追加した HTTP 固有マッピング（`RemoteAddr: entry.RemoteAddr,` 等 11 行）を `${DOMAIN_FIELDS_MAPPING}` プレースホルダーに置換（R5-003 対応）
   - T1-1 で追加した `TimeLocal: entry.TimeLocal.Format("${TIME_FORMAT}")` は、LogEntry の TimeLocal → Timestamp 統合（T4-2）に伴い削除される。代わりに共通フィールドの `Timestamp: entry.Timestamp,` が使用される

5. **Info() の Description を動的化**: ハードコーディングされた `"${PLUGIN_NAME} log monitoring plugin for Falco"` を `${PLUGIN_DESCRIPTION}` テンプレート変数に置き換え（E7 テンプレート変数仕様参照）

#### テンプレート展開機構（R-306 対応）

現在のテンプレートシステムは `${VARIABLE}` 形式の単純な文字列置換のため、
複数行のコードブロックを動的生成するには以下のプレースホルダーを導入する:

| プレースホルダー | 用途 | 展開先テンプレート |
|----------------|------|-----------------|
| `${DOMAIN_FIELDS_STRUCT}` | 構造体のドメイン固有フィールド | `plugin.go.tmpl`（PluginEvent）、`parser.go.tmpl`（LogEntry） |
| `${DOMAIN_FIELDS_DEFS}` | Fields() のフィールド定義配列 | `plugin.go.tmpl` |
| `${DOMAIN_FIELDS_EXTRACT}` | Extract() の switch/case 文 | `plugin.go.tmpl` |
| `${DOMAIN_FIELDS_MAPPING}` | parseLine() のフィールドマッピング | `plugin.go.tmpl` |
| `${DOMAIN_FIELDS_PARSE_JSON}` | parseJSON() のドメイン固有フィールド設定 | `parser.go.tmpl` |

scaffold スキル（T4-4）が WF-Phase 0 で収集したフィールド定義から対応するコードブロックを
生成し、テンプレート展開時に plugin.go.tmpl および parser.go.tmpl のプレースホルダーに挿入する。

#### 注意事項

- **推奨方式**: 具体的な型付きフィールド生成（要件定義書の設計選択表参照）
- openclaw が採用している方式と同じ
- 汎用マップ方式は代替案として残すが、推奨しない

#### 完了条件

- [ ] PluginEvent が共通 + ドメイン固有の 2 層構造になっている
- [ ] Fields() がドメイン固有フィールドを動的生成している
- [ ] Extract() がドメイン固有フィールドの case 文を動的生成している
- [ ] HTTP 固有フィールドがハードコーディングされていない
- [ ] Info() の Description が `${PLUGIN_DESCRIPTION}` で動的化されている

---

### T4-2: LogEntry のドメイン非依存化

| 項目 | 内容 |
|------|------|
| 要件ID | A2-3 |
| 優先度 | P1 High |
| 依存関係 | T4-1 とセットで実装 |
| 後続タスク | T4-4（scaffold スキル更新） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L497-L581（A2-3 セクション）、TimeLocal 統合と parser_test 更新 | 設計方針と注意事項 |
| `.claude/templates/plugin/parser.go.tmpl` | L41-L59（LogEntry）、L96-L105（format switch） | 現在の HTTP 固有構造 |
| `.claude/templates/plugin/parser_test.go.tmpl` | 全体（248行） | 更新が必要なテスト |
| `/Users/takaos/lab/falco-plugin-openclaw/pkg/parser/parser.go` | LogEntry 周辺 | 参照実装 |

#### 変更内容

1. **編集**: `.claude/templates/plugin/parser.go.tmpl` — LogEntry を 2 層構造に変更
   - 共通: `Timestamp`, `Raw`, `Headers`, `SecurityThreat`
   - ドメイン固有: `${DOMAIN_FIELDS_STRUCT}` プレースホルダーで型付きフィールドを生成（A1-2 PluginEvent と同じ方式）
   - `TimeLocal` を `Timestamp` に統合

2. **編集**: `.claude/templates/plugin/parser.go.tmpl` — フォーマット固有パーサー関数のテンプレート展開対応
   - `parseJSON()`: ドメイン固有フィールド設定を `${DOMAIN_FIELDS_PARSE_JSON}` プレースホルダーに置き換え
   - `parseCombined()` / `parseCommon()`: テンプレートにはスタブ（TODO コメントのみ）を残し、scaffold スキルが WF-Phase 2 でログフォーマットに応じて関数本体を生成する方式とする。正規表現変数（`combinedPattern`, `commonPattern`）もテンプレートから削除し、HTTP ドメイン選択時のみ scaffold が生成する
   - `parseCustom()`: ユーザー定義のため変更不要（TODO のまま維持）

3. **編集**: `.claude/templates/plugin/parser_test.go.tmpl` — ドメイン非依存テストに更新
   - HTTP 固有のテスト（TestParseCombined 等）をテンプレート化
   - `TimeLocal` を参照するテストアサーション（`entry.TimeLocal` 等）を `entry.Timestamp` に更新（R5-003 対応）

#### 完了条件

- [ ] LogEntry が共通 + `${DOMAIN_FIELDS_STRUCT}` による型付きフィールドの 2 層構造になっている
- [ ] `TimeLocal` が削除され `Timestamp` に統合されている（parser_test.go.tmpl の参照も更新済み）
- [ ] `parseJSON()` に `${DOMAIN_FIELDS_PARSE_JSON}` が使用されている
- [ ] `parseCombined()` / `parseCommon()` がテンプレート展開対応になっている
- [ ] `parser_test.go.tmpl` が更新されている

---

### T4-3: フォーマット自動検出モード追加

| 項目 | 内容 |
|------|------|
| 要件ID | A2-2 |
| 優先度 | P2 Medium（ドメイン非依存化との技術的関連性から Step 4 に配置） |
| 依存関係 | T2-2（JSON パーサー実装が前提） |
| 後続タスク | T4-5（parser スキル更新） |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L438-L495（A2-2 セクション） | After コード例 |
| `.claude/templates/plugin/parser.go.tmpl` | L96-L105（format switch） | 変更箇所 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/parser.go.tmpl`

1. format switch に `"auto"` case を追加
2. `parseAuto()` メソッドを追加（先頭文字 `{` で JSON/テキスト判定）。テキスト判定時のフォールバック先は、Parser 構造体に `textParseFunc` フィールドを追加して保持する（`"auto"` 選択時に scaffold がドメインに応じたテキストパーサーを設定）

#### 注意事項

- **openclaw との設計差異**: openclaw の参照実装では `parseAuto()` は独立メソッドではなく、
  `Parse()` メソッド内（L104-136）にインライン実装されている。
  本タスクでは独立メソッド方式を採用する（テスト容易性が高いため）。
  openclaw を参照実装としてコピーする際に混乱しないよう注意すること。

#### 完了条件

- [ ] `"auto"` case が format switch に追加されている
- [ ] `parseAuto()` が JSON とテキストを自動判定する

---

### T4-4: plugin-scaffold スキル更新

| 項目 | 内容 |
|------|------|
| 要件ID | B1 |
| 優先度 | P1 High |
| 依存関係 | T4-1, T4-2 の完了 |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1142-L1201（B1 セクション） | 変更内容とテンプレート対応表 |
| `.claude/skills/plugin-scaffold/SKILL.md` | 全体（239行） | 現在のスキル定義 |

#### 変更内容

**編集対象**: `.claude/skills/plugin-scaffold/SKILL.md`（239行）

1. 生成ファイル一覧を 14 → 18 ファイルに更新（Step 4 完了時点で存在する scaffold 担当テンプレート分。残り 2 テンプレート `CLAUDE.md.tmpl`（T5-4）, `CHANGELOG.md.tmpl`（T5-5）は Step 5 で作成後に T5-8 で追加し 20 に更新。テスト担当 2 テンプレート `plugin_test.go.tmpl`, `e2e_pattern_test.go.tmpl` は test スキル担当）
2. WF-Phase 0 にドメイン固有フィールド収集手順を追加
3. テンプレート→生成ファイル→担当スキル対応表を追加
4. `${PLUGIN_DESCRIPTION}`, `${LOG_SOURCE}` の収集手順を追加
5. WF-Phase 1 のテンプレート展開時に、収集したフィールド定義から 5 つのプレースホルダーのコードブロックを生成する手順を追加:
   - `${DOMAIN_FIELDS_STRUCT}`: 構造体フィールド定義 → `plugin.go.tmpl`（PluginEvent）と `parser.go.tmpl`（LogEntry）の **2 ファイル**に展開
   - `${DOMAIN_FIELDS_DEFS}`: Fields() のフィールド定義配列 → `plugin.go.tmpl`
   - `${DOMAIN_FIELDS_EXTRACT}`: Extract() の switch/case 文 → `plugin.go.tmpl`
   - `${DOMAIN_FIELDS_MAPPING}`: parseLine() のフィールドマッピング → `plugin.go.tmpl`
   - `${DOMAIN_FIELDS_PARSE_JSON}`: parseJSON() のドメイン固有フィールド設定 → `parser.go.tmpl`
6. WF-Phase 0 のフィールド収集時に JSON キー名も収集する手順を追加（`${DOMAIN_FIELDS_PARSE_JSON}` 生成に必要）

#### 注意事項

- フィールド収集時に Go の型として Extract() で使用可能なのは `string` と `uint64` のみ（Falco SDK の制約。R5-012 参照）。ユーザーが他の型を指定した場合は `string` として格納し、変換コードを生成する旨をガイドに記載する

#### 完了条件

- [ ] 生成ファイル一覧が 18 ファイルになっている（Step 4 完了時点の scaffold 担当分。T5-4/T5-5 完了後に 20 に更新）
- [ ] フィールド収集手順がある
- [ ] 新規テンプレート変数の収集手順がある
- [ ] 5 つのプレースホルダーのコード生成手順がある
- [ ] JSON キー名の収集手順がある

---

### T4-5: plugin-parser スキル更新

| 項目 | 内容 |
|------|------|
| 要件ID | B2 |
| 優先度 | P1 High |
| 依存関係 | T2-2, T4-3 の完了 |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1202-L1209（B2 セクション） | 変更内容 |
| `.claude/skills/plugin-parser/SKILL.md` | 全体（212行） | 現在のスキル定義 |

#### 変更内容

**編集対象**: `.claude/skills/plugin-parser/SKILL.md`（212行）

1. JSON パーサーのデフォルト実装手順を追加（A2-1 反映）
2. `"auto"` フォーマット検出モードの説明を追加（A2-2 反映）
3. 入力サイズ超過時の挙動を「スキップ」→「切り詰めて続行」に変更（A3-1 反映）

#### 完了条件

- [ ] JSON パーサーの手順がある
- [ ] `"auto"` モードの説明がある
- [ ] 入力サイズ超過の説明が「切り詰め」になっている

---

### Step 4 統合検証（受け入れテスト）

要件定義書 L1418-L1435（E8）に定義された受け入れテストを実施:

```bash
# AT-1: HTTP プラグイン生成
# ログ形式=combined, フィールド=HTTP 標準
go vet ./... && go test ./... && make build

# AT-2: AI プラグイン生成
# ログ形式=json, フィールド定義:
#   session_id: string, "ai.session_id"  — セッション識別子
#   type:       string, "ai.type"        — リクエストタイプ
#   tool:       string, "ai.tool"        — 使用ツール名
#   args:       string, "ai.args"        — ツール引数
go vet ./... && go test ./... && make build

# AT-3: IoT プラグイン生成
# ログ形式=custom, フィールド定義:
#   device_id:   string,  "iot.device_id"   — デバイス識別子
#   sensor_type: string,  "iot.sensor_type" — センサー種別
#   value:       string,  "iot.value"       — 計測値（R5-012: Falco SDK は string/uint64 のみ対応のため string で格納）
go vet ./... && go test ./... && make build

# AT-4: macOS ビルド
# AT-1 を macOS arm64 で実行
make build  # → .dylib 生成

# AT-5: E2E テスト
make e2e    # Level 1 + Level 2 全通過
```

---

## Step 5: ドキュメント・仕組み化【P2 Medium】

**目標**: ドキュメント生成の自動化、残りの改善項目の完了、ワークフロー全体の統合

**要件定義書の該当セクション**: セクション 7「Step 5」(L1512-L1529)

---

### T5-1: ~/パス展開 + パストラバーサル防止ロジック追加

| 項目 | 内容 |
|------|------|
| 要件ID | A1-3 |
| 優先度 | P2 Medium |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L299-L328（A1-3 セクション） | After コード例 |
| `.claude/templates/plugin/plugin.go.tmpl` | Open() 内（L120付近）のログパスイテレーション | 変更箇所 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/plugin.go.tmpl`

Open() 内のログパスイテレーションにパストラバーサル防止と `~/` 展開を追加:

```go
for _, logPath := range p.config.LogPaths {
    // パストラバーサル防止（E6 セキュリティ要件、R-304 対応）
    if strings.Contains(logPath, "..") {
        return nil, fmt.Errorf("path traversal not allowed: %s", logPath)
    }

    // ~/パス展開
    if strings.HasPrefix(logPath, "~/") {
        if home, err := os.UserHomeDir(); err == nil {
            logPath = filepath.Join(home, logPath[2:])
        }
    }
    // ... 既存のファイルオープン処理 ...
}
```

import に `"os"` と `"path/filepath"` を追加（未追加の場合）。

#### 完了条件

- [ ] `~/` で始まるパスがホームディレクトリに展開される
- [ ] `../` を含むパスが拒否される（パストラバーサル防止）

---

### T5-2: Extract() 冗長チェック削除

| 項目 | 内容 |
|------|------|
| 要件ID | A1-4 |
| 優先度 | P3 Low |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L329-L362（A1-4 セクション） | Before/After コード例 |
| `.claude/templates/plugin/plugin.go.tmpl` | L332-L343（Extract 冒頭） | 変更箇所 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/plugin.go.tmpl`

不要な `evt.EventData()` 呼び出しと nil チェックを削除し、直接 `gob.NewDecoder(evt.Reader())` でデコードする。

#### 注意事項

- T4-1 で Extract() がドメイン非依存化のために書き換え済みの場合、`evt.EventData()` が既に削除されている可能性がある。その場合は本タスクはスキップする

#### 完了条件

- [ ] `evt.EventData()` 呼び出しが削除されている
- [ ] 直接 `gob.NewDecoder(evt.Reader())` でデコードしている

---

### T5-3: URL デコード重複排除 ※ T1-1 に統合済み (R5-001)

| 項目 | 内容 |
|------|------|
| 要件ID | A3-2 |
| 優先度 | ~~P2 Medium~~ → **T1-1 に統合（Step 1 で実施）** |
| 依存関係 | なし |
| 後続タスク | なし |

**注**: R5-001 により T1-1 の追加作業として Step 1 に前倒しされました。parser 接続（T1-1）後に
`detectSecurityPatterns()` と `DetectSecurityThreat()` の両方で URL デコードが実行されると
最大6段階のデコードが発生するため、parser 接続と同時に重複排除を行います。
変更内容・完了条件は T1-1 の項目 8 を参照してください。

---

### T5-4: CLAUDE.md テンプレート作成

| 項目 | 内容 |
|------|------|
| 要件ID | A8-1 |
| 優先度 | P2 Medium |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L998-L1028（A8-1 セクション） | テンプレート概要 |
| `/Users/takaos/lab/falco-plugin-openclaw/CLAUDE.md` | 全体（119行） | **最重要参照**: 実際の CLAUDE.md |

#### 変更内容

**新規作成**: `.claude/templates/plugin/CLAUDE.md.tmpl`

テンプレート変数: `${PLUGIN_NAME}`, `${LOG_SOURCE}`, `${EVENT_SOURCE}`

内容:
- Project Overview
- Build & Development Commands
- Architecture
- Critical Constraints（P002, P004, P008, P010 等）

#### 完了条件

- [ ] テンプレートが新規作成されている
- [ ] テンプレート変数が正しく使われている
- [ ] Critical Constraints に P コードの知見が含まれている

---

### T5-5: CHANGELOG.md テンプレート作成

| 項目 | 内容 |
|------|------|
| 要件ID | A8-2 |
| 優先度 | P2 Medium |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1030-L1036（A8-2 セクション） | 概要 |
| `/Users/takaos/lab/falco-plugin-openclaw/CHANGELOG.md` | 先頭 50 行 | 参照: Keep a Changelog 形式 |

#### 変更内容

**新規作成**: `.claude/templates/plugin/CHANGELOG.md.tmpl`

Keep a Changelog 準拠のスケルトン。

#### 完了条件

- [ ] テンプレートが新規作成されている
- [ ] Keep a Changelog フォーマットに準拠している

---

### T5-6: README.md.tmpl 更新

| 項目 | 内容 |
|------|------|
| 要件ID | A8-3 |
| 優先度 | P2 Medium |
| 依存関係 | T1-2, T2-3, T3-2, T3-4, T1-3 の完了 |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1038-L1052（A8-3 セクション） | 影響を受ける改善 ID 一覧 |
| `.claude/templates/plugin/README.md.tmpl` | 全体（118行） | 現在の README テンプレート |
| `/Users/takaos/lab/falco-plugin-openclaw/README_ja.md` | 参考 | 参照実装 |

#### 変更内容

**編集対象**: `.claude/templates/plugin/README.md.tmpl`（118行）

追加するセクション:
1. OS 自動検出によるビルド説明（A4-1 反映）
2. `make e2e`, `make e2e-pattern`, `make e2e-pipeline` の説明（A4-3 反映）
3. 3 環境設定ファイルの説明（A6-1 反映）
4. 3 層テストアーキテクチャの概要（A7-1/A7-2 反映）

#### 完了条件

- [ ] macOS/Linux 両対応のビルド説明がある
- [ ] E2E テストコマンドの説明がある
- [ ] 3 環境 Falco 設定の説明がある
- [ ] 3 層テストの概要がある

---

### T5-7: dev-kit-feedback スキル新規作成

| 項目 | 内容 |
|------|------|
| 要件ID | C1 |
| 優先度 | P2 Medium |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1273-L1290（C1 セクション） | 処理フローと目的 |
| `.claude/skills/plugin-scaffold/SKILL.md` | 全体 | 既存スキルのフォーマット参照 |

#### 変更内容

**新規作成**: `.claude/skills/dev-kit-feedback/SKILL.md`

処理フロー:
1. 指定パスのプラグインコードを読み込む
2. dev-kit テンプレートとの差分を検出
3. 差分を分類（テンプレート改善 / スキル改善 / 新規パターン）
4. 改善提案レポートを出力
5. PROBLEM_PATTERNS.md への追加候補を提示

#### 完了条件

- [ ] スキルファイルが新規作成されている
- [ ] 処理フロー（5 ステップ）が記述されている
- [ ] `/dev-kit-feedback [plugin-path]` で実行可能な形式になっている

---

### T5-8: plugin-dev-workflow エージェント更新

| 項目 | 内容 |
|------|------|
| 要件ID | B6 |
| 優先度 | P2 Medium |
| 依存関係 | Step 1-4 の全タスク、および Step 5 の他のタスク（T5-1, T5-2, T5-4〜T5-7, T5-9。T5-3 は T1-1 に統合済みのため除外）の完了が前提 |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1257-L1269（B6 セクション） | 各 Phase の変更内容 |
| `.claude/agents/plugin-dev-workflow.md` | 全体（265行） | 現在のワークフロー定義 |
| 本文書の全タスク | 各 T*-* の完了条件 | 反映すべき変更一覧 |

#### 変更内容

**編集対象**: `.claude/agents/plugin-dev-workflow.md`（265行）

| 変更内容 | 対象 Phase |
|---------|-----------|
| ドメイン固有フィールド収集 | WF-Phase 0 |
| 新規テンプレートファイル一覧更新 | WF-Phase 1 |
| parser 統合を自動実行 | WF-Phase 2 |
| Level 2 テスト生成追加 | WF-Phase 4 |
| macOS ネイティブビルド追加 | WF-Phase 5 |
| 品質ゲートに Level 2 テスト通過を追加 | WF-Phase 4→5 |
| 完了報告に E2E テスト結果サマリー追加 | WF-Phase 6 |

#### 完了条件

- [ ] 全 Phase が v2 の変更を反映している
- [ ] 品質ゲートに Level 2 テスト通過が含まれている
- [ ] 新規テンプレート（8 ファイル: scaffold 担当 6 + test 担当 2）が WF-Phase 1 の生成一覧に正しく反映されている

---

### T5-9: plugin-rules スキル更新

| 項目 | 内容 |
|------|------|
| 要件ID | B3 |
| 優先度 | P2 Medium |
| 依存関係 | なし |
| 後続タスク | なし |

#### コンテキスト復元

| 読むべきファイル | 箇所 | 目的 |
|----------------|------|------|
| `docs/requirements/dev-kit-v2-requirements.md` | L1210-L1216（B3 セクション） | 変更内容 |
| `.claude/skills/plugin-rules/SKILL.md` | 全体（294行） | 現在のスキル定義 |

#### 変更内容

**編集対象**: `.claude/skills/plugin-rules/SKILL.md`（294行）

1. ドメイン非依存のルール構造ガイドを追加（HTTP 以外の例を含む）
2. priority の使い分け基準を追加（CRITICAL / WARNING / NOTICE）
3. **`plugin_rules.yaml.tmpl` のドメイン対応ガイドを追加**（R-303 対応）:
   - 現在のテンプレートは HTTP/Web セキュリティルール（SQLi, XSS, PathTraversal, CMDi, Suspicious UA）が
     ハードコーディングされている
   - AI ログや IoT プラグインでは不適切なルールが生成される
   - スキルにドメインに応じたルールカスタマイズ手順を追加:
     - HTTP プラグイン: テンプレートのルールをそのまま使用
     - AI プラグイン: prompt injection, data exfiltration 等のルールに置換
     - IoT プラグイン: anomaly detection, threshold violation 等のルールに置換
   - カスタマイズ手順の具体例を記載

#### 完了条件

- [ ] ドメイン非依存のルールガイドがある
- [ ] priority 基準の説明がある
- [ ] `plugin_rules.yaml.tmpl` のドメイン別カスタマイズ手順がある

---

### Step 5 統合検証

生成されたプラグインディレクトリで以下を実行:

```bash
# コード変更（T5-1 パス展開、T5-2 Extract 冗長削除）の検証（T5-3 は T1-1 に統合済み、Step 1 で検証済み）
go vet ./...                     # 静的解析パス
go test ./... -v                 # 全テスト通過

# ドキュメント生成の確認
ls CLAUDE.md CHANGELOG.md        # ドキュメントが生成されている
grep "macOS" README.md           # macOS ビルドの説明がある
grep "e2e" README.md             # E2E テストの説明がある
grep "falco-local" README.md     # 3 環境設定の説明がある
```

---

## 全体の依存関係グラフ

```
Step 1:
  T1-1 (parser接続) ─────────────────────────→ T1-3, T4-1
  T1-2 (OS自動検出) ─────────────────────────→ T3-1, T3-6, T5-6
  T1-3 (Level 2テスト) ←T1-1 ───────────────→ T1-4, T3-2, T5-6
  T1-4 (testスキル) ←T1-3

Step 2:
  T2-1 (truncate)
  T2-2 (JSONパーサー) ──────────────────────→ T4-3, T4-5
  T2-3 (Level 1テスト) ←T2-4 ────────────→ T3-2, T5-6
  T2-4 (パターン拡張) ──────────────────────→ T2-3
  T2-5 (PROBLEM_PATTERNS)

Step 3:
  T3-1 (build-release) ←T1-2 ─────────────→ T3-6
  T3-2 (E2Eターゲット) ←T1-3,T2-3 ────────→ T5-6
  T3-3 (CI分離)
  T3-4 (3環境設定) ────────────────────────→ T5-6
  T3-5 (config非依存化) ───────────────────→ T4-1
  T3-6 (buildスキル) ←T1-2,T3-1

Step 4:
  T4-1 (PluginEvent非依存化) ←T1-1,T3-5 ──→ T4-4
  T4-2 (LogEntry非依存化) ←→T4-1 (セット) ─→ T4-4
  T4-3 (auto検出) ←T2-2 ──────────────────→ T4-5
  T4-4 (scaffoldスキル) ←T4-1,T4-2
  T4-5 (parserスキル) ←T2-2,T4-3

Step 5:
  T5-1 (パス展開)
  T5-2 (Extract冗長削除)
  T5-3 (URLデコード重複) ※ T1-1 に統合済み
  T5-4 (CLAUDE.mdテンプレート)
  T5-5 (CHANGELOG.mdテンプレート)
  T5-6 (README更新) ←T1-2,T2-3,T3-2,T3-4,T1-3
  T5-7 (feedbackスキル)
  T5-8 (エージェント更新) ←── 全Step完了
  T5-9 (rulesスキル)
```

---

## 進捗管理

各タスクのステータスを以下で管理する:

| タスクID | ステータス | 完了日 | コミット |
|---------|-----------|--------|---------|
| T1-1 | 完了 | 2026-03-16 | `bc35ef8` |
| T1-2 | 完了 | 2026-03-16 | `bc35ef8` |
| T1-3 | 完了 | 2026-03-16 | `bc35ef8` |
| T1-4 | 完了 | 2026-03-16 | `bc35ef8` |
| T2-1 | 完了 | 2026-03-16 | `8e61f00` |
| T2-2 | 完了 | 2026-03-16 | `8e61f00` |
| T2-3 | 完了 | 2026-03-16 | `8e61f00` |
| T2-4 | 完了 | 2026-03-16 | `8e61f00` |
| T2-5 | 完了 | 2026-03-16 | `8e61f00` |
| T3-1 | 完了 | 2026-03-16 | `dab05fc` |
| T3-2 | 完了 | 2026-03-16 | `dab05fc` |
| T3-3 | 完了 | 2026-03-16 | `dab05fc` |
| T3-4 | 完了 | 2026-03-16 | `dab05fc` |
| T3-5 | 完了 | 2026-03-16 | `dab05fc` |
| T3-6 | 完了 | 2026-03-16 | `dab05fc` |
| T4-1 | 完了 | 2026-03-16 | `781c1bf` |
| T4-2 | 完了 | 2026-03-16 | `781c1bf` |
| T4-3 | 完了 | 2026-03-16 | `781c1bf` |
| T4-4 | 完了 | 2026-03-16 | `781c1bf` |
| T4-5 | 完了 | 2026-03-16 | `781c1bf` |
| T5-1 | 完了 | 2026-03-16 | `316ab2c` |
| T5-2 | 完了 | 2026-03-16 | `316ab2c` |
| T5-3 | T1-1に統合 | 2026-03-16 | `bc35ef8` |
| T5-4 | 完了 | 2026-03-16 | `316ab2c` |
| T5-5 | 完了 | 2026-03-16 | `316ab2c` |
| T5-6 | 完了 | 2026-03-16 | `316ab2c` |
| T5-7 | 完了 | 2026-03-16 | `316ab2c` |
| T5-8 | 完了 | 2026-03-16 | `316ab2c` |
| T5-9 | 完了 | 2026-03-16 | `316ab2c` |

---

## 改訂履歴

| 日付 | バージョン | 変更内容 |
|------|-----------|---------|
| 2026-03-07 | 1.0 | 初版作成。要件定義書 v4.0 の全改善項目を 29 タスクに分解 |
| 2026-03-07 | 2.0 | レビュー R-301〜R-318 + S-301 の指摘を反映。scaffold ファイル数修正(R-301)、T3-5 config→detector 配線追加(R-302)、T5-9 plugin_rules.yaml.tmpl 対応拡張(R-303)、T5-1 パストラバーサル防止追加(R-304)、依存グラフ T3-5→T4-1/T4-2 追加(R-305)、T4-1 テンプレート展開機構定義(R-306)、Step 1 統合検証に E4 互換性確認追加(R-307)、T1-1 完全フィールドマッピング(R-308)、T4-3 優先度注記(R-311)、Step 1 手動展開検証経路(R-315)、T2-5 挿入位置明確化(R-316)、AT-2/AT-3 フィールド定義追加(R-318)、T4-3 openclaw parseAuto 注記(S-301) |
| 2026-03-10 | 2.1 | 再レビュー R2-001〜R2-010 の指摘を反映。基盤文書バージョン v4.0→v5.0 修正(R2-001)、T4-1 依存関係に T3-5 追加(R2-003)、T1-2 後続タスクから T3-2 削除(R2-004)、T3-5 コンテキスト復元に parser.go/regex_simple.go 追加(R2-007)、T5-1 行番号追加(R2-008)、T4-4 展開機構の完了条件追加(R2-009)、Step 2 統合検証コマンド修正(R2-010) |
| 2026-03-10 | 2.2 | 修正レビュー R3-001〜R3-018 反映。T1-3 行番号参照 L793→L821 修正(R3-005)、T3-5 後続タスクから T4-2 削除(R3-004/R3-017)、T5-6 依存関係に T2-3 追加(R3-006)、Step 2/5 統合検証に go vet 追加・前提条件明記(R3-007/R3-018)、T4-1 に Info() Description 動的化追加(R3-003)、依存グラフを Step 別に整理し直し(R3-011) |
| 2026-03-10 | 2.3 | R3-019 反映。T4-2 にパーサー関数テンプレート展開の変更内容・完了条件追加、T4-4 プレースホルダー 4→5 に更新（`${DOMAIN_FIELDS_PARSE_JSON}` 追加）、T2-2 にプレースホルダー配置の記述追加 |
| 2026-03-10 | 2.4 | 第3回レビュー R4-001〜R4-006 反映。T4-1 プレースホルダーテーブルに DOMAIN_FIELDS_PARSE_JSON 追加(R4-001)、行番号参照 32 箇所の系統的修正(R4-003)、T4-1 テキスト修正(R4-004)、T1-1 後続タスクに T4-1 追加(R4-005)、T5-8 依存関係明確化(R4-006) |
| 2026-03-15 | 2.5 | 第4回レビュー R5-001〜R5-013 反映。T5-3をT1-1に統合しURLデコード二重実行を解消(R5-001)、Step 4前提にテンプレート展開メカニズム明確化(R5-002)、T4-1/T4-2にTimeLocal→Timestamp移行パス追記(R5-003)、後続タスク6箇所を依存グラフと整合(R5-005〜R5-010)、基盤文書バージョンv5.0→v5.4(R5-011/R5-013)、AT-3 value型をfloat64→string(R5-012) |
| 2026-03-15 | 2.6 | 実装リハーサルレビュー RH-001〜RH-024 反映。T1-3 型キャスト手順追加(RH-001)、T3-5 parser_test.go.tmpl 更新追加(RH-002)、T1-1 行番号修正(RH-003)、T2-2 HTTP フィールド方針追加(RH-004)、T2-3 ファイル配置・命名規約追加(RH-005/RH-006)、T4-2 パーサー方式明確化(RH-007)、T4-4 ファイル数 20→18 修正(RH-008)、依存関係・注意事項追加(RH-009〜RH-024) |

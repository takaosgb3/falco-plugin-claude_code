---
name: plugin-rules
description: Falcoルールの作成・検証・ベストプラクティスチェックを支援する。ユーザーが「ルールを作成」「Falcoルール」「検出ルール」「_rules.yaml」「condition/output/priority」「source指定」「icontains」「ルールのベストプラクティス」「アラート設定」「セキュリティルール」と言った場合にトリガーする。HTTP/AI/IoTドメイン別のルールカスタマイズガイドとpriority基準（CRITICAL/WARNING/NOTICE）を含む。P003(source必須)/P005(evt.type不使用)/P011(YAMLコメント)/P019(1イベント1ルール)の知見を適用。パーサー実装にはplugin-parser、テスト実行にはplugin-testを使用すること。
argument-hint: "[action] [rule-type]"
---

# Falcoルール作成支援

$ARGUMENTS についてFalcoルールの作成・検証を支援します。

## 引数

- **action**: 実行アクション
  - `create`: ルールの新規作成
  - `validate`: 既存ルールの検証
  - `add-pattern`: 検出パターンの追加
- **rule-type**: ルールタイプ
  - `detection`: 検出ルール
  - `list`: リスト定義
  - `macro`: マクロ定義
  - `all`: 全タイプ（デフォルト）

## 実行手順

### 1. Phase 1: フィールド定義の読み込み

プラグインの `Fields()` 定義を確認し、ルールで使用可能なフィールドを把握する。

```bash
# プラグインソースからフィールド定義を抽出
grep -A 2 'Name:.*"<plugin>\.' cmd/plugin-sdk/plugin.go
```

典型的なフィールド:

| フィールド | 型 | 説明 |
|-----------|------|------|
| `<plugin>.method` | string | HTTPメソッド |
| `<plugin>.path` | string | リクエストパス |
| `<plugin>.query_string` | string | クエリストリング |
| `<plugin>.request_uri` | string | 完全なリクエストURI（計算フィールド） |
| `<plugin>.status` | uint64 | ステータスコード |
| `<plugin>.client_ip` | string | クライアントIP |
| `<plugin>.user_agent` | string | User-Agent |
| `<plugin>.bytes_sent` | uint64 | レスポンスサイズ |
| `<plugin>.headers[key]` | string | リクエストヘッダー |

### 2. Phase 2: ルール生成

#### 2.1 リスト定義の生成

```yaml
- list: ${EVENT_SOURCE}_safe_methods
  items: ['"GET"', '"HEAD"', '"OPTIONS"']

- list: ${EVENT_SOURCE}_sqli_patterns
  items: [
    '"'' or "', '"'' and "', '"union select"', '"union all select"',
    '"%27"', '"%2527"', '"1=1"', '"sleep("'
  ]

- list: ${EVENT_SOURCE}_xss_patterns
  items: [
    '"<script"', '"javascript:"', '"onerror="', '"onload="',
    '"%3cscript"', '"%3Cscript"'
  ]

- list: ${EVENT_SOURCE}_path_traversal_patterns
  items: [
    '"../"', '"..%2f"', '"..%252f"',
    '"/etc/passwd"', '"/etc/shadow"'
  ]

- list: ${EVENT_SOURCE}_cmdi_patterns
  items: [
    '";ls"', '"$(cat"', '"`cat"', '"| cat"',
    '"%3Bls"', '"%7Ccat"'
  ]

- list: ${EVENT_SOURCE}_suspicious_agents
  items: [
    '"sqlmap"', '"nikto"', '"nmap"', '"masscan"',
    '"dirbuster"', '"gobuster"', '"wpscan"'
  ]
```

#### 2.2 マクロ定義の生成

```yaml
- macro: ${EVENT_SOURCE}_is_write_method
  condition: >
    ${EVENT_SOURCE}.method in ("POST", "PUT", "DELETE", "PATCH")
```

#### 2.3 検出ルール例

##### SQL Injection

```yaml
- rule: "[${PLUGIN_NAME_UPPER} SQLi] SQL Injection Attempt"
  desc: "Detects SQL injection attempts in request URI"
  condition: >
    ${EVENT_SOURCE}.request_uri icontains "' or "
    or ${EVENT_SOURCE}.request_uri icontains "union select"
    or ${EVENT_SOURCE}.request_uri icontains "%27"
    or ${EVENT_SOURCE}.request_uri icontains "%2527"
    or ${EVENT_SOURCE}.request_uri icontains "1=1"
    or ${EVENT_SOURCE}.request_uri icontains "sleep("
  output: >
    [${PLUGIN_NAME_UPPER} SQLi] SQL Injection detected
    (ip=%${EVENT_SOURCE}.client_ip
    method=%${EVENT_SOURCE}.method
    uri=%${EVENT_SOURCE}.request_uri
    status=%${EVENT_SOURCE}.status
    ua=%${EVENT_SOURCE}.user_agent)
  priority: WARNING
  source: ${EVENT_SOURCE}
  tags: [security, sqli, owasp, ${PLUGIN_NAME}]
```

##### XSS

```yaml
- rule: "[${PLUGIN_NAME_UPPER} XSS] Cross-Site Scripting Attempt"
  desc: "Detects XSS attempts in request URI"
  condition: >
    ${EVENT_SOURCE}.request_uri icontains "<script"
    or ${EVENT_SOURCE}.request_uri icontains "javascript:"
    or ${EVENT_SOURCE}.request_uri icontains "onerror="
    or ${EVENT_SOURCE}.request_uri icontains "%3cscript"
  output: >
    [${PLUGIN_NAME_UPPER} XSS] XSS detected
    (ip=%${EVENT_SOURCE}.client_ip
    method=%${EVENT_SOURCE}.method
    uri=%${EVENT_SOURCE}.request_uri
    status=%${EVENT_SOURCE}.status)
  priority: WARNING
  source: ${EVENT_SOURCE}
  tags: [security, xss, owasp, ${PLUGIN_NAME}]
```

##### Path Traversal

```yaml
- rule: "[${PLUGIN_NAME_UPPER} PathTraversal] Path Traversal Attempt"
  desc: "Detects path traversal attempts"
  condition: >
    ${EVENT_SOURCE}.request_uri icontains "../"
    or ${EVENT_SOURCE}.request_uri icontains "..%2f"
    or ${EVENT_SOURCE}.request_uri icontains "/etc/passwd"
  output: >
    [${PLUGIN_NAME_UPPER} PathTraversal] Path traversal detected
    (ip=%${EVENT_SOURCE}.client_ip
    uri=%${EVENT_SOURCE}.request_uri
    status=%${EVENT_SOURCE}.status)
  priority: WARNING
  source: ${EVENT_SOURCE}
  tags: [security, path_traversal, owasp, ${PLUGIN_NAME}]
```

##### Command Injection

```yaml
- rule: "[${PLUGIN_NAME_UPPER} CMDi] Command Injection Attempt"
  desc: "Detects command injection attempts"
  condition: >
    ${EVENT_SOURCE}.request_uri icontains ";ls"
    or ${EVENT_SOURCE}.request_uri icontains "$(cat"
    or ${EVENT_SOURCE}.request_uri icontains "|cat"
  output: >
    [${PLUGIN_NAME_UPPER} CMDi] Command injection detected
    (ip=%${EVENT_SOURCE}.client_ip
    uri=%${EVENT_SOURCE}.request_uri
    status=%${EVENT_SOURCE}.status)
  priority: WARNING
  source: ${EVENT_SOURCE}
  tags: [security, cmdi, owasp, ${PLUGIN_NAME}]
```

##### Suspicious User-Agent

```yaml
- rule: "[${PLUGIN_NAME_UPPER} UA] Suspicious User-Agent Detected"
  desc: "Detects suspicious scanning tools"
  condition: >
    ${EVENT_SOURCE}.user_agent icontains "sqlmap"
    or ${EVENT_SOURCE}.user_agent icontains "nikto"
    or ${EVENT_SOURCE}.user_agent icontains "nmap"
    or ${EVENT_SOURCE}.user_agent icontains "dirbuster"
  output: >
    [${PLUGIN_NAME_UPPER} UA] Suspicious user agent detected
    (ip=%${EVENT_SOURCE}.client_ip
    ua=%${EVENT_SOURCE}.user_agent
    uri=%${EVENT_SOURCE}.request_uri)
  priority: NOTICE
  source: ${EVENT_SOURCE}
  tags: [security, suspicious_agent, ${PLUGIN_NAME}]
```

#### 2.4 ドメイン別カスタマイズ例（`plugin_rules.yaml.tmpl`）

HTTPプラグイン以外では、上記のデフォルトルールをドメインに合わせて置換する:

```yaml
# AI プラグインの例
- rule: "[${PLUGIN_NAME_UPPER} PromptInj] Prompt Injection Attempt"
  desc: "Detects prompt injection in AI model input"
  condition: >
    ${EVENT_SOURCE}.input icontains "ignore previous instructions"
    or ${EVENT_SOURCE}.input icontains "system prompt"
  output: >
    [${PLUGIN_NAME_UPPER} PromptInj] Prompt injection detected
    (user=%${EVENT_SOURCE}.user_id input=%${EVENT_SOURCE}.input)
  priority: CRITICAL
  source: ${EVENT_SOURCE}
  tags: [security, prompt_injection, ai, ${PLUGIN_NAME}]

# IoT プラグインの例
- rule: "[${PLUGIN_NAME_UPPER} Threshold] Sensor Threshold Violation"
  desc: "Detects sensor value exceeding safe threshold"
  condition: >
    ${EVENT_SOURCE}.value > 100
    or ${EVENT_SOURCE}.value < 0
  output: >
    [${PLUGIN_NAME_UPPER} Threshold] Threshold violated
    (device=%${EVENT_SOURCE}.device_id value=%${EVENT_SOURCE}.value)
  priority: WARNING
  source: ${EVENT_SOURCE}
  tags: [security, threshold, iot, ${PLUGIN_NAME}]
```

### 3. Phase 3: ルール検証

#### 3.1 構文検証

```bash
# falcoがインストールされている場合
if command -v falco &> /dev/null; then
  falco -V rules/${PLUGIN_NAME}_rules.yaml
else
  echo "INFO: falco未インストール。YAML構文チェックで代替。"
  yamllint rules/${PLUGIN_NAME}_rules.yaml 2>/dev/null || echo "yamllintも未インストール。CI/CDで検証推奨。"
fi
```

#### 3.2 ベストプラクティスチェック

以下の項目を自動チェック:

```bash
# P003: 全ルールに source: <plugin_name> が含まれているか
grep -c "source:" rules/${PLUGIN_NAME}_rules.yaml
grep -c "^- rule:" rules/${PLUGIN_NAME}_rules.yaml
# 両者が一致すること

# P005: evt.type を使用していないか
grep "evt.type" rules/${PLUGIN_NAME}_rules.yaml && echo "ERROR: evt.type はプラグインルールでは使用不可" || echo "OK"

# icontains が使用されているか
grep -c "icontains" rules/${PLUGIN_NAME}_rules.yaml

# URLエンコードパターンが含まれているか
grep -c "%27\|%3c\|%2f" rules/${PLUGIN_NAME}_rules.yaml

# P012: Output内のheaders参照が小文字か
grep "headers\[" rules/${PLUGIN_NAME}_rules.yaml | grep -v "headers\[[a-z-]*\]" && echo "WARNING: headers参照は小文字で" || echo "OK"

# P011: YAML複数行条件内にコメントがないか
# condition: > ブロック内の # を検出
```

チェック項目一覧:

| ID | チェック項目 | 対策パターン |
|----|------------|-------------|
| P003 | 全ルールに `source: ${EVENT_SOURCE}` | 必須 |
| P005 | `evt.type` を使用していない | プラグインルールでは不可 |
| P006 | URLエンコードパターン含む | `%27`, `%3c`, `%2f` 等 |
| P011 | YAML複数行条件内にコメントなし | パースエラー防止 |
| P012 | Output headers参照が小文字 | `headers[x-test-id]` |
| P015 | クロスルール干渉なし | `icontains` の部分文字列マッチ注意 |
| P016 | URLエンコード精度 | 多段エンコード対応 |

## コンテキスト補完情報

### 参照すべきドキュメント

- `rules/nginx_rules.yaml`: ルールの参照実装
- Serena memory: `falcorule_best_practices`
- `PROBLEM_PATTERNS.md`: 過去のルール関連問題

### 過去の失敗パターン（要注意）

1. **P003: source指定必須** — `source: <plugin_name>` がないとルールエンジンが登録しない（Issue #643）
2. **P005: evt.type不使用** — プラグインルールでは `evt.type` は使用不可
3. **P011: YAML複数行コメント禁止** — `condition: >` ブロック内の `#` はパースエラーになる
4. **P012: Headers小文字必須** — Output内の `%<plugin>.headers[key]` は小文字で指定
5. **P015: クロスルール干渉** — `icontains` は部分文字列マッチなので、意図しないルール発火に注意

### ドメイン別ルールカスタマイズ

`plugin_rules.yaml.tmpl` はHTTP/Webセキュリティルールがデフォルトで含まれています。
異なるドメインのプラグインでは、ドメインに応じたルールに置換してください:

- **HTTP プラグイン**: テンプレートのルール（SQLi, XSS, PathTraversal, CMDi, Suspicious UA）をそのまま使用
- **AI プラグイン**: prompt injection, data exfiltration, unauthorized model change 等に置換
- **IoT プラグイン**: anomaly detection, threshold violation, unauthorized access 等に置換

ルール構造の共通パターン:
- `source: ${EVENT_SOURCE}` は必須（P003）
- `condition:` で `${EVENT_SOURCE}.field_name` を使用
- `output:` にアラートの詳細情報を含める

### Priority ガイドライン

| Priority | 用途 | 例 |
|----------|------|-----|
| CRITICAL | 即時対応が必要な脅威 | Command Injection, Data Exfiltration |
| WARNING | セキュリティ上の懸念 | SQLi, XSS, Path Traversal |
| NOTICE | 監視・記録対象 | Suspicious User-Agent, Unusual Access Pattern |

### テンプレート変数の使い方

| コンテキスト | 使用する変数 | 例 |
|-------------|-------------|-----|
| ルールの`source:` | `${EVENT_SOURCE}` | `source: apache` |
| フィールドプレフィックス | `${EVENT_SOURCE}` | `apache.method` |
| ルール名プレフィックス | `${PLUGIN_NAME_UPPER}` | `[APACHE SQLi]` |
| ファイル名 | `${PLUGIN_NAME}` | `apache_rules.yaml` |

## 成功基準

| ID | 基準 | 検証方法 |
|----|------|----------|
| SC-020 | 全ルールに`source: <plugin_name>`が含まれている | grep確認 |
| SC-021 | `falco -V`が成功する | コマンド実行（利用可能な場合） |
| SC-022 | `evt.type`を使用していない | grep確認 |
| SC-023 | セキュリティパターンで`icontains`を使用している | grep確認 |
| SC-024 | URLエンコードパターン(%27, %3c等)が含まれている | grep確認 |

## 重要な注意事項

- 全ルールに `source: ${EVENT_SOURCE}` を必ず指定する（P003）
- `evt.type` はプラグインルールでは使用不可（P005）
- セキュリティ検出条件では `icontains` を推奨（大文字小文字を無視）
- URLエンコードパターン（`%27`, `%3c`, `%2f` 等）と多段エンコード（`%2527` 等）を含める
- Output内のヘッダー参照は小文字で指定する（`headers[x-test-id]`）
- YAML複数行条件（`condition: >`）内に `#` コメントを入れない

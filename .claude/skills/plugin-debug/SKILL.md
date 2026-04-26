---
name: plugin-debug
description: Falcoプラグイン開発中のエラー・障害のトラブルシューティングを支援する。ユーザーが「エラーが出る」「ビルドが通らない」「テストが失敗する」「Falcoがプラグインをロードしない」「パニック」「nil pointer」「動かない」「なぜ動かない」「debug」「トラブルシュート」「原因を調べて」「修正して」「このエラーの意味は」「make buildが失敗」「go testが落ちる」「ルールが発火しない」「アラートが出ない」と言った場合にトリガーする。P001-P021の21パターンの知見を自動適用して診断する。ビルド手順そのものにはplugin-build、テスト作成にはplugin-test、ルール構文にはplugin-rulesを使用すること。このスキルは「何かがうまくいかない」状況でのみ使用し、正常な開発フローには他のスキルを使うこと。
argument-hint: "[error-message-or-symptom]"
---

# Falco プラグイン デバッグ支援

$ARGUMENTS についてエラーの診断と修正を支援します。

## 診断フロー

### 1. エラー情報の収集

まず現在の状況を把握する。ユーザーからエラーメッセージが提供されていない場合は、以下を実行して情報を集める。

```bash
# ビルドエラーの確認
go vet ./... 2>&1 | head -50

# テスト失敗の確認
go test ./... -v 2>&1 | tail -50

# Makefile の存在確認
ls Makefile go.mod 2>/dev/null
```

### 2. エラーカテゴリの判定

エラーメッセージまたは症状から、以下のカテゴリに分類する。

| カテゴリ | 典型的なエラー | 関連Pコード |
|---------|-------------|-----------|
| ビルドエラー | `go build` / `make build` 失敗 | P001, P002, P013 |
| テスト失敗 | `go test` FAIL | P004, P010, P021 |
| ランタイムエラー | Falco起動時・実行時エラー | P003, P005, P007, P008, P009 |
| macOS固有 | macOSでのみ発生する問題 | P017, P018 |
| ルール不発火 | アラートが出ない | P003, P005, P006, P008, P009, P015, P016, P019 |
| ルール構文エラー | YAML パースエラー | P011 |
| データ不整合 | フィールド値が空・不正 | P004, P010, P012, P014, P020 |

### 3. カテゴリ別の診断手順

---

#### ビルドエラー

**症状**: `make build` や `go build` が失敗する

**診断チェックリスト**:

```bash
# 1. Go環境の確認
go version
echo "GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=$CGO_ENABLED"

# 2. 依存関係の確認
go mod tidy
go mod verify

# 3. 静的解析
go vet ./...
```

**P002チェック（-buildmode=c-shared忘れ）**:
Makefile に `-buildmode=c-shared` が含まれているか確認。このフラグがないと Falco にロードできないバイナリが生成される。
```bash
grep "buildmode=c-shared" Makefile
```

**P001チェック（macOSバイナリ）**:
macOS でビルドした .dylib を Linux にデプロイしていないか確認。
```bash
file lib*-plugin-* 2>/dev/null
# macOS: "Mach-O 64-bit dynamically linked shared library"
# Linux: "ELF 64-bit LSB shared object" ← Falcoが必要とする形式
```

**P013チェック（GLIBC互換性）**:
CI/CDのGLIBCバージョンとデプロイ先が一致しているか確認。

**よくあるビルドエラーと対処**:

| エラーメッセージ | 原因 | 対処 |
|---------------|------|------|
| `undefined: parser.New` | import パス不一致 | `go.mod` のモジュールパスと import パスを確認 |
| `cannot use ... as type uint64` | 型の不一致 | `int` → `uint64` キャストを追加 |
| `CGO_ENABLED=0` でのビルド | CGO未有効化 | `CGO_ENABLED=1` を設定 |
| `cc: command not found` | Cコンパイラ未インストール | `gcc` / `build-essential` をインストール |

---

#### テスト失敗

**症状**: `go test` で FAIL が出る

**診断チェックリスト**:

```bash
# 1. 失敗テストの特定
go test ./... -v 2>&1 | grep -E "^--- FAIL"

# 2. 特定テストの詳細実行
go test ./pkg/parser/... -v -run "TestXxx" -count=1

# 3. レース条件の確認
go test ./... -race -count=1
```

**P004チェック（nil map panic）**:
`panic: assignment to entry in nil map` → `Headers` マップが未初期化。
parser.go の `Parse()` 内で `entry.Headers == nil` のチェックと `make(map[string]string)` 初期化を確認。

**P021チェック（fsnotifyタイミング）**:
テストが断続的に失敗する場合、fsnotify の非同期配信が原因の可能性。
`time.Sleep` による待機が十分か確認（最低 200ms、書き込み後の検証には 500ms 推奨）。

**P010チェック（Fields/Extract不一致）**:
`Fields()` で定義したフィールドが `Extract()` の switch 文に存在するか確認。
```bash
# Fields() のフィールド数と Extract() の case 数を比較
grep -c "Name:" cmd/plugin-sdk/plugin.go
grep -c "^	case " cmd/plugin-sdk/plugin.go
```

**よくあるテスト失敗と対処**:

| エラーメッセージ | 原因 | 対処 |
|---------------|------|------|
| `panic: assignment to entry in nil map` | P004: Headers 未初期化 | `make(map[string]string)` で初期化 |
| `Timed out waiting for event` | P021: fsnotify タイミング | `time.Sleep` を増やす |
| `GOB decode error` | PluginEvent 構造体変更 | エンコード/デコードの構造体を一致させる |
| `expected X, got Y` | フィールドマッピングミス | parseLine() のマッピングを確認 |

---

#### ランタイムエラー（Falco実行時）

**症状**: Falco がプラグインをロードしない、起動時にエラーが出る

**P008チェック（load_plugins欠落）**:
```bash
grep "load_plugins" falco.yaml falco-local.yaml 2>/dev/null
# 出力がない → load_plugins ディレクティブが欠落している
```

**P007チェック（rules_filesパス）**:
```bash
grep "rules_files" falco.yaml
# ディレクトリ指定ではなく個別ファイルパスになっているか確認
```

**プラグインロードの確認**:
```bash
# Falco がプラグインを認識しているか
falco --list-plugins 2>/dev/null

# バイナリパスの確認
grep "library_path" falco.yaml
ls -la $(grep "library_path" falco.yaml | awk '{print $2}') 2>/dev/null
```

---

#### ルール不発火

**症状**: ログを流してもアラートが出ない

**診断ステップ**（上から順にチェック）:

1. **P008: load_plugins** → プラグイン自体がロードされていない
2. **P003: source 指定** → ルールに `source:` がない、または間違っている
   ```bash
   grep "source:" rules/*_rules.yaml
   ```
3. **P005: evt.type 使用** → プラグインイベントには evt.type が存在しない
   ```bash
   grep "evt.type" rules/*_rules.yaml
   # 出力がある → このルールは発火しない
   ```
4. **P009: レート制限** → 連続発火がサイレントに抑制されている
   ```bash
   grep -E "rate:|max_burst:" falco.yaml
   # rate: 0, max_burst: 0 でなければ抑制される可能性
   ```
5. **P019/P015: クロスルール干渉** → 先行ルールがマッチしてしまっている
   - ルールファイル内の順序を確認。より具体的なルールが上位にあるか
   - 1イベント1ルール制約のため、広い条件のルールが先にマッチすると後続は発火しない

---

#### macOS 固有の問題

**P017チェック（outputs拒否）**:
```bash
grep "outputs:" falco-local.yaml 2>/dev/null
# macOS 用設定に outputs: が含まれている → 削除が必要
```

**P018チェック（-Uフラグ）**:
macOS で Falco のアラートが表示されない場合、`-U` フラグを追加:
```bash
falco -c falco-local.yaml --disable-source syscall -U
```

---

#### データ不整合

**P014チェック（SeekEnd）**:
起動時に既存ログが大量に再処理される → `Open()` で `file.Seek(0, io.SeekEnd)` を確認。

**P012チェック（headers小文字）**:
ヘッダーフィールドの値が取得できない → parser 側で `strings.ToLower()` で正規化しているか確認。

**P020チェック（truncation vs 全文）**:
10KB超の入力で検出結果とフィールド値が不一致 → これは意図的な設計。検出は先頭10KBのみ、Extract()は全文返却。

---

## 4. 修正の適用

診断結果に基づき、以下の優先順で修正を適用する:

1. **即座に修正可能な問題**: import パス、型キャスト、設定値の修正
2. **テンプレート確認が必要な問題**: `.claude/templates/plugin/` の対応テンプレートと比較して正しい実装を確認
3. **設計変更が必要な問題**: ユーザーに報告し、方針を確認してから修正

修正後は必ず検証:
```bash
go vet ./...       # 静的解析
go test ./... -v   # テスト実行
make build         # ビルド確認
```

## 参照ドキュメント

- `PROBLEM_PATTERNS.md`: P001-P021 の詳細な問題パターンと対策
- `.claude/templates/plugin/`: テンプレートファイル群（正しい実装の参照元）
- `.claude/skills/plugin-build/SKILL.md`: ビルド手順の詳細
- `.claude/skills/plugin-test/SKILL.md`: テスト手順の詳細

## 成功基準

| ID | 基準 | 検証方法 |
|----|------|----------|
| DB-001 | エラーの根本原因が特定されている | 診断レポート |
| DB-002 | 関連する P コードが参照されている | P コードの引用 |
| DB-003 | 修正提案が具体的である | コード差分の提示 |
| DB-004 | 修正後に検証コマンドが実行されている | go vet / go test / make build の結果 |

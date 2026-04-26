---
name: plugin-build
description: Falcoプラグインのビルド・検証・パッケージングを支援する。ユーザーが「ビルドしたい」「make build」「コンパイル」「バイナリ生成」「.so/.dylib」「リリースパッケージ」「make package」「ELF検証」「make verify」「macOSビルド」「Linuxビルド」「クロスコンパイル」「CGO_ENABLED」「-buildmode=c-shared」「GLIBC互換性」と言った場合にトリガーする。macOS(.dylib)/Linux(.so)の自動検出ビルド、build-releaseサイズ最適化、ELFバイナリ検証を含む。P001(macOSバイナリ)/P002(c-shared必須)/P013(ビルド環境)/P017(macOS outputs)/P018(-Uフラグ)の知見を適用。テスト実行にはplugin-test、ルール検証にはplugin-rulesを使用すること。
argument-hint: "[action] [target]"
---

# ビルド・デプロイ支援

$ARGUMENTS についてプラグインのビルド・検証を支援します。

## 引数

- **action**: 実行アクション
  - `build`: ビルド実行
  - `verify`: バイナリ検証
  - `package`: リリースパッケージ作成
- **target**: ビルドターゲット
  - `linux-amd64`: Linux x86_64（デフォルト）
  - `linux-arm64`: Linux ARM64
  - `all`: 全ターゲット

## 実行手順

### 1. build: ビルド実行

#### 1.1 環境判定

```bash
OS=$(uname -s)
ARCH=$(uname -m)
echo "Build environment: ${OS} ${ARCH}"
```

#### 1.2 Linux環境でのビルド

```bash
# 必須ビルドフラグ（⚠️ -buildmode=c-shared は絶対に省略しない）
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -buildmode=c-shared \
  -o lib${PLUGIN_NAME}-plugin-linux-amd64.so \
  ./cmd/plugin-sdk/

echo "Build completed: lib${PLUGIN_NAME}-plugin-linux-amd64.so"
```

**重要なビルドフラグ**:

| フラグ | 必須 | 説明 |
|--------|------|------|
| `CGO_ENABLED=1` | 必須 | CGOを有効化（Falco Plugin SDKで必須） |
| `GOOS=linux` | 必須 | Linux用バイナリ |
| `GOARCH=amd64` | 必須 | x86_64アーキテクチャ |
| `-buildmode=c-shared` | 必須 | 共有ライブラリとしてビルド（P002） |

#### 1.3 macOS ネイティブビルド（.dylib）

macOS 環境では `.dylib` 形式のネイティブバイナリをビルドできる。Makefile の OS 自動検出により自動的に適切なバイナリが生成される。

```bash
# macOS ネイティブビルド（.dylib が生成される）
make build
# → lib${PLUGIN_NAME}-plugin-darwin-arm64.dylib (Apple Silicon)
# → lib${PLUGIN_NAME}-plugin-darwin-amd64.dylib (Intel)

# ローカルテスト用 Falco 設定
# P017: macOS の Falco は outputs: セクションを拒否するため falco-local.yaml を使用
# P018: stdout バッファリング無効化のため -U フラグが必要
falco -c falco-local.yaml --disable-source syscall -U
```

**注意**: macOS ビルドは開発・テスト専用。Linux デプロイ用のバイナリ（.so）は Linux 環境または CI/CD でビルドすること。

#### 1.4 リリースビルド

```bash
# サイズ最適化リリースビルド（-trimpath -ldflags="-s -w"）
make build-release

# リリースパッケージ作成（ビルド + 検証 + チェックサム）
make package
```

#### 1.5 macOS 環境での静的解析

Linux クロスコンパイルは不可だが、静的解析とテストは実行可能。

```bash
# 静的解析
go vet ./...

# テスト実行
go test ./... -v -race

# Level 1 + Level 2 E2E テスト
make e2e
```

### 2. verify: バイナリ検証

#### 2.1 ELF検証

```bash
# バイナリの種別確認（P001対策）
file lib${PLUGIN_NAME}-plugin-linux-amd64.so
# 期待出力: "ELF 64-bit LSB shared object, x86-64"

# ❌ NG例: "Mach-O 64-bit" → macOSバイナリ
# ❌ NG例: "ELF 64-bit LSB executable" → 実行可能ファイル（-buildmode=c-shared 忘れ）
```

#### 2.2 ファイルサイズ確認

```bash
ls -lh lib${PLUGIN_NAME}-plugin-linux-amd64.so
# 通常 5〜20MB程度
```

#### 2.3 GLIBC互換性チェック（P013対策）

```bash
# Linux環境の場合
if command -v objdump &> /dev/null; then
  objdump -T lib${PLUGIN_NAME}-plugin-linux-amd64.so | grep GLIBC | sort -t'_' -k2 -V | tail -5
  echo "INFO: GLIBC 2.17以上が必要（一般的なLinuxディストリビューション）"
fi
```

### 3. package: リリースパッケージ作成

#### 3.1 SHA256チェックサム生成

```bash
sha256sum lib${PLUGIN_NAME}-plugin-linux-amd64.so > checksums.sha256
cat checksums.sha256
```

#### 3.2 リリースアセット一覧

```
リリースに含めるアセット:
├── lib${PLUGIN_NAME}-plugin-linux-amd64.so  ← プラグインバイナリ
├── ${PLUGIN_NAME}_rules.yaml                ← Falcoルール
└── checksums.sha256                         ← チェックサム
```

#### 3.3 Makefile テンプレート

生成される Makefile:

```makefile
PLUGIN_NAME := ${PLUGIN_NAME}
BINARY := lib$(PLUGIN_NAME)-plugin-linux-amd64.so
SRC_DIR := ./cmd/plugin-sdk

.PHONY: build test lint clean verify package

build:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		go build -buildmode=c-shared \
		-o $(BINARY) $(SRC_DIR)/

test:
	go test ./... -v

lint:
	golangci-lint run

clean:
	rm -f $(BINARY) *.h coverage.out checksums.sha256

verify: build
	@echo "Verifying binary..."
	@file $(BINARY) | grep -q "ELF 64-bit LSB shared object" \
		&& echo "OK: Valid ELF shared object" \
		|| (echo "ERROR: Not a valid ELF shared object"; exit 1)

package: build verify
	sha256sum $(BINARY) > checksums.sha256
	sha256sum rules/$(PLUGIN_NAME)_rules.yaml >> checksums.sha256
	@echo "Package ready for release"
	@echo "Assets:"
	@echo "  - $(BINARY)"
	@echo "  - rules/$(PLUGIN_NAME)_rules.yaml"
	@echo "  - checksums.sha256"
```

#### 3.4 GitHub Actions ワークフローテンプレート

CI/CDワークフロー（`.github/workflows/ci.yml`）:

```yaml
name: CI Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    # Option 1: セルフホストランナー（既存の開発環境向け）
    # runs-on: [self-hosted, linux, x64, local]
    # Option 2: GitHub-hosted ランナー（新規リポジトリ向け）
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Test
        run: go test ./... -v
      - name: Lint
        run: |
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
          golangci-lint run

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build
        run: make build
      - name: Verify
        run: make verify
```

## コンテキスト補完情報

### 参照すべきドキュメント

- `CLAUDE.md`: 「ビルドとリリースの鉄則」セクション
- `Makefile`: ビルドタスクの参照実装
- `.github/workflows/ci-pipeline.yml`: CI/CDの参照実装
- `PROBLEM_PATTERNS.md`: ビルド関連の失敗パターン

### 過去の失敗パターン（要注意）

1. **P001: macOSバイナリリリース** — macOSでビルドしたバイナリをLinux用として配布しない。`file`コマンドでELF確認必須
2. **P002: -buildmode=c-shared忘れ** — 必ず共有ライブラリとしてビルド。省略すると実行可能ファイルになりFalcoがロードできない
3. **P013: GLIBC互換性** — ビルド環境のGLIBCバージョンが実行環境と互換性があること
4. **LC-001: macOSクロスコンパイル不可** — macOSでは `go vet` と `go test` のみ実行
5. **LC-002: ランナー設定** — CI/CDテンプレートはself-hosted/GitHub-hostedの両オプションを提供

### ビルド環境チェックリスト

```
Linux環境:
  ✅ CGO_ENABLED=1
  ✅ GOOS=linux GOARCH=amd64
  ✅ -buildmode=c-shared
  ✅ file コマンドで ELF 64-bit 確認
  ✅ SHA256 チェックサム生成

macOS環境:
  ✅ go vet ./...
  ✅ go test ./...
  ❌ go build -buildmode=c-shared（不可）
  ❌ ELFバイナリ生成（不可）
```

## 成功基準

| ID | 基準 | 検証方法 |
|----|------|----------|
| SC-040 | バイナリが「ELF 64-bit LSB shared object」 | `file`コマンド（Linux環境のみ） |
| SC-041 | SHA256チェックサムが生成されている | ファイル存在確認 |
| SC-042 | Makefileが正しく動作する | `make build && make test`（Linux） / `make -n build`（macOS） |
| SC-043 | GitHub ActionsワークフローのYAML構文が正しく、適切なランナー設定 | YAML構文チェック、ランナー設定確認 |

## 重要な注意事項

- **絶対にワークフローでビルドされたバイナリを使用すること**（ローカルビルドは検証用のみ）
- macOSでは本番用バイナリは生成できない。Linux環境（CI/CDまたはLinuxマシン）でビルドすること
- `-buildmode=c-shared` を絶対に省略しない（P002）
- `file` コマンドで必ず「ELF 64-bit LSB shared object」を確認（P001）
- 新規リポジトリでは GitHub-hosted ランナー（`ubuntu-latest`）を使用可能
- 開発リポジトリ（falco-nginx-plugin-claude）ではセルフホストランナー必須

# Rehearsal Round 4: Phase 5-6（build / release / 運用）

## リハーサル想定シナリオ
実装者が初回リリースを `make package` → release 公開 → ユーザに macOS 導入してもらう全フローを再現する。

## 抽出論点

### RH4-01: SBOM / cosign 署名の具体実装が未明示【手順実行可能性】
**詰まる場面**: §21.3 SC-002「SBOM（CycloneDX JSON or SPDX、Syft / `go-licenses` で自動生成可）」と SC-003「cosign keyless 署名（GitHub Actions OIDC）」が「推奨」だが、v0.1 で先行導入したい場合の **具体コマンド** が無い。release.yml に組み込むサンプルもない。

**修正対象**: 要件 §21.3 末尾に v0.1 で導入できる最小実装サンプルを追加：
- SBOM 生成: `syft scan dir:. -o cyclonedx-json=sbom.cdx.json`
- cosign 署名: `cosign sign-blob --yes lib*.so > lib*.so.sig`
- GitHub Actions step 例

---

### RH4-02: doctor CLI の `--max-age` の duration 形式が未明示【情報十分性】
**詰まる場面**: §22.4 OPS-005 で `--max-age <duration>` とあるが:
- `--max-age 900`（秒？）
- `--max-age 15m`（Go time.Duration 文字列？）
- `--max-age PT15M`（ISO 8601？）

どれか実装者が決めるしかない。

**修正対象**: 要件 §22.4 OPS-005 に「duration は Go `time.ParseDuration` 形式（例: `--max-age 15m` / `--max-age 1h30m` / `--max-age 30s`）」を明示。

---

### RH4-03: self-check rule の同梱方法が未明示【手順実行可能性】
**詰まる場面**: §22.4 OPS-004「Self-check rule を rule pack に同梱」だが、性質が異なる（events.jsonl に書き込みがない＝検出系ルール条件は不発火、自前で時刻ベース監視）ため通常の検出ルールと同居しにくい。実装者は:
- `rules/claude_code_rules.yaml` 末尾に追加するか
- `rules/claude_code_health.yaml` として別ファイルにするか

決めかねる。Falco の rules_files ロード順序にも影響する（P007 個別パス指定）。

**修正対象**: 要件 §29 付録 B 末尾と §22.4 OPS-004 で「**別ファイル `rules/claude_code_health.yaml`** として切り出す」方針を v0.1 で採用。

---

### RH4-04: チェックサム生成コマンドが未明示【手順実行可能性】
**詰まる場面**: §21.2 B-007「checksum を必ず生成」だが、生成コマンドが未明示。`sha256sum`（Linux）と `shasum -a 256`（macOS）が異なるため、release.yml で OS 横断で書く場合の例が必要。

**修正対象**: 要件 §21.2 B-007 にコマンド例（`sha256sum lib*.so > checksums.sha256` / `shasum -a 256 lib*.dylib >> checksums.sha256`）を追加。

---

### RH4-05: macOS 個人導入で `falco-local.yaml` 入手元が未明示【手順実行可能性】
**詰まる場面**: §22.1 個人 macOS 導入の手順 4 で `falco -c falco-local.yaml --disable-source syscall -U` だが、`falco-local.yaml` の入手元が示されない。release artifact §21.1 にあるが、`wget` / `curl` / `git clone` のどれを使うか未明示。

**修正対象**: 要件 §22.1 の手順 1 と 4 の間に「release から falco-local.yaml + rules を取得」のステップを追加。

---

## 観点別件数

| 観点 | 件数 |
|---|---:|
| 1. **手順実行可能性** | **4（RH4-01, RH4-03, RH4-04, RH4-05）** |
| 2. 情報十分性 | 1（RH4-02） |
| 3. 設定値整合性 | 0 |
| 4. テスト実行可能性 | 0 |

合計 5 件。Round 4 で適用予定。

## Round 4 適用済み

| # | 章 | 修正 | 状態 |
|---|----|------|------|
| RH4-01 | 要件 §21.3.1 新設 | SBOM 生成（anchore/sbom-action 例）と cosign keyless 署名（GitHub Actions step + ユーザ検証コマンド）の最小実装例を追加 | 適用 |
| RH4-02 | 要件 §22.4 OPS-005 | duration 形式を Go time.ParseDuration 互換と明示。整数指定は受け付けない方針 | 適用 |
| RH4-03 | 要件 §22.4 OPS-004 | self-check rule を別ファイル `rules/claude_code_health.yaml` に切り出す方針を明示 | 適用 |
| RH4-04 | 要件 §21.2 B-007 | チェックサム生成・検証コマンドを Linux / macOS 別に明記 | 適用 |
| RH4-05 | 要件 §22.1 個人 macOS 導入 | 手順 0 として release artifacts ダウンロード + shasum 検証を先行ステップに追加 | 適用 |

合計 5 件適用、Round 4 抽出 5 件全件解消。


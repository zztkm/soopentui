# soopentui

Solod (So) 向けの [OpenTUI](https://github.com/anomalyco/opentui) バインディングです。
ネイティブ core（`libopentui.a`）を静的リンクして使う前提の薄い C ABI ラッパーです、現在開発中です。

```bash
go get github.com/zztkm/soopentui@latest
```

`go get` で届くのは So ソース・C ヘッダ・ビルド用ツールです。**リンク用の `libopentui.a` は同梱しません**（OS/arch 依存のため）。別途ビルドしてください。

## 前提

| ツール | 用途 |
|--------|------|
| Go 1.22+ | モジュール取得 / `cmd/*` の実行 |
| [Zig 0.15.2](https://ziglang.org/download/) | OpenTUI ネイティブ core とアプリのリンク |
| Git | OpenTUI の clone とパッチ適用 |
| [Solod (`so`)](https://github.com/solod-dev/solod) | So のトランスパイル (tip 版を利用します) |
| Xcode Command Line Tools（macOS） | SDK / システム framework |

```bash
go install solod.dev/cmd/so@main
go get solod.dev@main   # 標準ライブラリも tip 推奨（@latest は古いことがある）
```

## 利用の流れ

### 1. モジュールを取得

```bash
go get github.com/zztkm/soopentui@v0.1.0
```

アプリ側:

```go
import "github.com/zztkm/soopentui"
```

### 2. 静的ライブラリをビルド

```bash
go run github.com/zztkm/soopentui/cmd/opentui-static@v0.1.0
```

カレントディレクトリに `_build/opentui/.../libopentui.a` ができます（モジュールキャッシュは読み取り専用のため）。

### 3. トランスパイルとリンク

`so translate` 後、`zig cc` で生成 C と `libopentui.a` をリンクします。C ABI ヘッダはモジュールの `include/` にあります:

```bash
SOOPENTUI_DIR="$(go list -m -f '{{.Dir}}' github.com/zztkm/soopentui)"
# zig cc ... -I"$SOOPENTUI_DIR/include" ... path/to/libopentui.a
```

手元のクローンではサンプル一式をビルドできます:

```bash
go run ./cmd/hello-tui
./examples/hello-tui/hello-tui
```

| フラグ | 意味 |
|--------|------|
| `-o path` | 出力バイナリ |
| `-run` | ビルド後に実行 |
| `-skip-lib` | `libopentui.a` が無いときに自動ビルドしない |

## ローカル開発

このリポジトリを直接いじるときは、例アプリで一時的に replace を足します:

```go
replace github.com/zztkm/soopentui => ../..
```

公開済みバージョンだけを使う場合は replace を外し、`require` でタグを指定します。

## 既知の注意（macOS + Zig 0.15.2）

新しい Xcode / macOS SDK では Zig 0.15.2 の build runner が失敗することがあります。  
`cmd/opentui-static` は `DEVELOPER_DIR=/dev/null` を自動設定します。

```bash
OPENTUI_KEEP_DEVELOPER_DIR=1 go run ./cmd/opentui-static
```

`so build` の `LDFLAGS` は `-o` の後に付くため、framework リンクでは使えません。hello-tui は translate + 明示リンクです。

起動時に端末 capability 問い合わせの応答がシェルに漏れることがあります（stdin を読んでいないため）。表示自体は問題ありません。

## ライセンス

- soopentui: MIT（[LICENSE](LICENSE)）
- THIRD_PARTY_NOTICES: [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)

`libopentui.a` をリンクしたバイナリを配布する場合は、OpenTUI の MIT 表記に従ってください。OpenTUI 配下の依存は上流を参照してください。


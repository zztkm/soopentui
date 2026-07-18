# solod-vs-go

Solod (So) と Go、および OpenTUI 連携を試すための作業用リポジトリです。

## 前提

| ツール | 用途 |
|--------|------|
| Go 1.22+ | `go run` で OpenTUI の clone / パッチ / ビルド |
| [Zig 0.15.2](https://ziglang.org/download/) | OpenTUI ネイティブ core とアプリのリンク（`.zig-version` と一致） |
| Git | OpenTUI の clone とパッチ適用 |
| [Solod (`so`)](https://github.com/solod-dev/solod) | So アプリのトランスパイル |
| Xcode Command Line Tools（macOS） | SDK / システム framework |

```bash
# so CLI（tip / main）
go install solod.dev/cmd/so@main
```

So 標準ライブラリも tip を使う（`@latest` は古いタグのままのことがある）:

```bash
go get solod.dev@main
```

OpenTUI 本体は手動 clone 不要です（`go run` が `_build/opentui` に取得します）。

## OpenTUI を静的ライブラリとしてビルドする

```bash
go run ./cmd/opentui-static
```

1. 必要なら OpenTUI を `_build/opentui` に clone  
2. `patches/opentui-static-linkage.patch` を適用  
3. `zig build -Dlinkage=static` で `libopentui.a` を生成  

| フラグ | 意味 |
|--------|------|
| `--force` | `_build/opentui` を消して再 clone してから patch / build |
| `-skip-build` | clone / patch のみ |
| `-optimize mode` | Zig 最適化（既定: `ReleaseFast`） |

成果物例:

`_build/opentui/packages/core/src/zig/lib/aarch64-macos-static/libopentui.a`

## Hello TUI（Solod + OpenTUI）

画面に文言を出して 2 秒後に終了する最小アプリです。

```bash
go install solod.dev/cmd/so@main
go run ./cmd/hello-tui
./examples/hello-tui/hello-tui
```

`libopentui.a` が無ければ、内部で `go run ./cmd/opentui-static` も実行します。

| フラグ | 意味 |
|--------|------|
| `-o path` | 出力バイナリ（既定: `examples/hello-tui/hello-tui`） |
| `-run` | ビルド後に実行 |
| `-skip-lib` | 静的ライブラリが無いときに自動ビルドしない |

構成:

| パス | 役割 |
|------|------|
| `cmd/hello-tui/` | `go run` 用ビルドオーケストレータ |
| `include/opentui.h` | MVP 用 C ABI 宣言 |
| `soopentui.go` | So 向け第三者バインディング（`github.com/zztkm/soopentui`） |
| `examples/hello-tui/` | サンプルアプリ |

ビルドは `so translate` → `zig cc` で静的リンクします（単一バイナリ、`libopentui.dylib` 依存なし）。

## C スモーク

```bash
go run ./cmd/opentui-static
./examples/smoke-static/build.sh
```

## 既知の注意（macOS + Zig 0.15.2）

新しい Xcode / macOS SDK では Zig 0.15.2 の build runner が失敗することがあります。  
`go run ./cmd/opentui-static` は `DEVELOPER_DIR=/dev/null` を自動設定します。

無効化:

```bash
OPENTUI_KEEP_DEVELOPER_DIR=1 go run ./cmd/opentui-static
```

また `so build` の `LDFLAGS` は `-o` の後に付くため、`zig cc` + framework リンクでは使えません。hello-tui は translate + 明示リンクにしています。

## ディレクトリ

```text
cmd/opentui-static/     # OpenTUI: clone → patch → build
cmd/hello-tui/          # hello-tui: translate → link
patches/                # OpenTUI 向けパッチ
include/opentui.h       # MVP C ヘッダ
soopentui.go            # So 向け OpenTUI バインディング（モジュールルート）
examples/hello-tui/     # 最小 TUI
examples/smoke-static/  # C 静的リンクスモーク
_build/                 # 作業ディレクトリ（gitignore）
```

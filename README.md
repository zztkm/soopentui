# soopentui

Solod (So) 向けの [OpenTUI](https://github.com/anomalyco/opentui) バインディングです。
ネイティブ core（`libopentui.a`）を静的リンクして使う前提の薄い C ABI ラッパーです、現在開発中です。

```bash
go get github.com/zztkm/soopentui@main
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

まだリリースタグを打っていないため、利用は `main` 前提です。

### 1. モジュールを取得

```bash
go get github.com/zztkm/soopentui@main
```

```go
import "github.com/zztkm/soopentui"
```

### 2. ビルド

```bash
go run github.com/zztkm/soopentui/cmd/build@main .
```

`cmd/build` が `include/` の解決・`so translate`・`libopentui.a` とのリンクまで行います。  
`libopentui.a` が無ければ内部で `opentui-static` を実行し、カレントに `_build/opentui/.../libopentui.a` を作ります。

| フラグ | 意味 |
|--------|------|
| `-o path` | 出力バイナリ（省略時は `./<パッケージ名>`） |
| `-run` | ビルド後に実行 |
| `-skip-lib` | `libopentui.a` が無いときに自動ビルドしない |

静的ライブラリだけ先に用意する場合:

```bash
go run github.com/zztkm/soopentui/cmd/opentui-static@main
```

手元のクローンではサンプルも同じ導線です:

```bash
go run ./cmd/build -o examples/hello-tui/hello-tui ./examples/hello-tui
# または
go run ./cmd/hello-tui
./examples/hello-tui/hello-tui
```

## 公開 API（描画・端末制御）

薄い C ABI ラッパーです。キー／マウス入力のデコードやイベントループは含みません（stdin はアプリ側で読んでください）。`EnableMouse` は端末のマウス報告を有効化するだけです。

| 関数 | 役割 |
|------|------|
| `CreateRenderer` / `Destroy` | レンダラ生成・破棄 |
| `SetupTerminal` / `RestoreTerminal` / `ClearTerminal` | 端末モード |
| `SetClearOnShutdown` / `SetBackgroundColor` / `SetTerminalTitle` | 終了時クリア・背景・タイトル |
| `Resize` | リサイズ |
| `SetCursorPosition` / `SetCursorColor` / `SetCursorStyle` | カーソル |
| `EnableMouse` / `DisableMouse` | マウス報告の ON/OFF |
| `NextBuffer` / `CurrentBuffer` / `BufferWidth` / `BufferHeight` | バッファ |
| `Clear` / `FillRect` / `DrawText` / `DrawTextAttr` | 描画 |
| `Render` | 表示 |

## ローカル開発

このリポジトリを直接いじるときは、例アプリで一時的に replace を足します:

```go
replace github.com/zztkm/soopentui => ../..
```

公開利用時は replace を外し、`require` で `@main`（擬似バージョン）を指定します。タグが出たらそちらに切り替えてください。

## 既知の注意（macOS + Zig 0.15.2）

新しい Xcode / macOS SDK では Zig 0.15.2 の build runner が失敗することがあります。  
`cmd/opentui-static` は `DEVELOPER_DIR=/dev/null` を自動設定します。

```bash
OPENTUI_KEEP_DEVELOPER_DIR=1 go run ./cmd/opentui-static
```

`so build` の `LDFLAGS` は `-o` の後に付くため、framework リンクでは使えません。`cmd/build` は translate + 明示リンクです。

起動時に端末 capability 問い合わせの応答がシェルに漏れることがあります（stdin を読んでいないため）。表示自体は問題ありません。

## ライセンス

- soopentui: MIT（[LICENSE](LICENSE)）
- THIRD_PARTY_NOTICES: [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)

`libopentui.a` をリンクしたバイナリを配布する場合は、OpenTUI の MIT 表記に従ってください。OpenTUI 配下の依存は上流を参照してください。


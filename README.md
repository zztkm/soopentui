# solod-vs-go

Solod (So) と Go、および OpenTUI 連携を試すための作業用リポジトリです。

## 前提

| ツール | 用途 |
|--------|------|
| Go 1.22+ | `go run` で clone / パッチ / ビルド |
| [Zig 0.15.2](https://ziglang.org/download/) | OpenTUI ネイティブ core のビルド（clone した OpenTUI の `.zig-version` と一致させる） |
| Git | OpenTUI の clone とパッチ適用 |
| Xcode Command Line Tools（macOS） | SDK / システム framework |

OpenTUI の取得・ビルドはすべて `_build/` 以下で行います。手動 clone は不要です。

## OpenTUI を静的ライブラリとしてビルドする

OpenTUI 上流は動的ライブラリ（`dlopen` 用）前提です。  
このリポジトリでは **静的リンク用の `-Dlinkage=static`** をパッチで足し、`libopentui.a` をビルドします。

リポジトリルートで:

```bash
go run ./cmd/opentui-static
```

これが行うこと:

1. 必要なら `https://github.com/anomalyco/opentui.git` を `_build/opentui` に clone
2. `patches/opentui-static-linkage.patch` を適用（未適用なら）
3. `zig build -Dlinkage=static -Doptimize=ReleaseFast` を実行
4. 成果物パスを表示

成功時の出力例:

```text
clone: https://github.com/anomalyco/opentui.git -> .../_build/opentui
patch: applied opentui-static-linkage.patch
...
OK: static OpenTUI library ready (13.0M)
.../_build/opentui/packages/core/src/zig/lib/aarch64-macos-static/libopentui.a
```

### オプション

```bash
go run ./cmd/opentui-static -h
```

| フラグ | 意味 |
|--------|------|
| `--force` / `-force` | `_build/opentui` を削除して再 clone してから patch / build |
| `-opentui path` | OpenTUI の場所（既定: `_build/opentui`） |
| `-skip-patch` | パッチ適用をスキップ |
| `-skip-build` | clone / patch のみ（ビルドしない） |
| `-optimize mode` | Zig の最適化（既定: `ReleaseFast`） |

最初からやり直す:

```bash
go run ./cmd/opentui-static --force
```

clone とパッチだけ:

```bash
go run ./cmd/opentui-static -skip-build
```

### 成果物の場所

| プラットフォーム | パス |
|------------------|------|
| Apple Silicon macOS | `_build/opentui/packages/core/src/zig/lib/aarch64-macos-static/libopentui.a` |
| Intel macOS | `_build/opentui/packages/core/src/zig/lib/x86_64-macos-static/libopentui.a` |
| Linux arm64 | `_build/opentui/packages/core/src/zig/lib/aarch64-linux-static/libopentui.a` |
| Linux x86_64 | `_build/opentui/packages/core/src/zig/lib/x86_64-linux-static/libopentui.a` |

## 静的リンクのスモークテスト（C）

```bash
go run ./cmd/opentui-static
./examples/smoke-static/build.sh
```

ポイント:

- 最終リンクは **`zig cc`** を使う（Apple `ld` は Zig が固めた archive 内の C++ オブジェクトを拒むことがある）
- 成功すると `libopentui.dylib` への依存はなく、システム framework のみになる（macOS）

## 既知の注意（macOS + Zig 0.15.2）

新しい Xcode / macOS SDK では、Zig 0.15.2 の build runner がリンクに失敗することがあります。  
`go run ./cmd/opentui-static` は macOS 上で `DEVELOPER_DIR=/dev/null` を自動設定します。

この回避を無効化したい場合:

```bash
OPENTUI_KEEP_DEVELOPER_DIR=1 go run ./cmd/opentui-static
```

Homebrew のパッチ済み Zig 0.15.2 を使う方法もあります。

## ディレクトリ

```text
cmd/opentui-static/     # clone → patch → build
patches/                # OpenTUI 向けパッチ
examples/smoke-static/  # 静的リンクの C スモーク
_build/                 # 作業ディレクトリ（gitignore）
  opentui/              # 自動 clone
solod/                  # 任意の参照用 clone（gitignore）
```

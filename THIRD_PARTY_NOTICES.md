# Third-Party Notices

このリポジトリ（`soopentui`）自体は [MIT License](LICENSE) です。

`cmd/opentui-static` で取得・ビルドする OpenTUI ネイティブ core（`libopentui.a`）を静的リンクしたバイナリには、次の第三者コンポーネントが含まれます。  
配布物には、各ライセンスが求める著作権表示・許諾文の保持が必要です（とくに MIT）。

バージョンはビルド時に取得した OpenTUI / Zig 依存の時点のものに依存します。正確な条文は各上流リポジトリを参照してください。

## OpenTUI

- Project: [anomalyco/opentui](https://github.com/anomalyco/opentui)
- License: MIT
- Copyright: Copyright (c) 2025 opentui

```
MIT License

Copyright (c) 2025 opentui

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Yoga (via OpenTUI)

- Project: [facebook/yoga](https://github.com/facebook/yoga)（OpenTUI の `build.zig.zon` 依存）
- License: MIT
- Copyright: Copyright (c) Facebook, Inc. and its affiliates.

```
MIT License

Copyright (c) Facebook, Inc. and its affiliates.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## uucode (via OpenTUI)

- Project: [jacobsandlund/uucode](https://github.com/jacobsandlund/uucode)
- License: MIT（加えて Unicode データ等の別表記が上流にあります。詳細は上流の `LICENSE.md` / `licenses/` を参照）
- Copyright: Copyright (c) 2026 Jacob Sandlund

```
MIT License

Copyright (c) 2026 Jacob Sandlund

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## miniaudio (via OpenTUI)

- Project: [mackron/miniaudio](https://github.com/mackron/miniaudio)（OpenTUI vendor）
- License: Public Domain (Unlicense) または MIT No Attribution（利用者が選択）
- Copyright (MIT-0): Copyright 2025 David Reid

miniaudio は attribution 必須ではありません（Public Domain / MIT-0）。完全な条文は上流ヘッダ末尾を参照してください。

## Solod について

`soopentui` は Solod のソースやバイナリを再配布しません。  
So アプリ側が `so` / `solod.dev` を使う場合のライセンスは、そのアプリの配布者が Solod 上流に従って判断してください（本リポジトリの第三者表記対象外）。

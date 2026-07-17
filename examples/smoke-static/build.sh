#!/usr/bin/env bash
# Link smoke.c against statically built libopentui.a into a single binary.
#
# Prefer zig cc over Apple clang: Zig's static archive embeds C/C++ objects
# (Yoga, miniaudio) that Apple ld currently rejects as "not 8-byte aligned".
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
DIR="$(cd "$(dirname "$0")" && pwd)"

case "$(uname -m)" in
  arm64|aarch64) ARCH=aarch64 ;;
  x86_64) ARCH=x86_64 ;;
  *) echo "error: unsupported arch $(uname -m)" >&2; exit 1 ;;
esac
case "$(uname -s)" in
  Darwin) OS=macos ;;
  Linux) OS=linux ;;
  *) echo "error: unsupported OS $(uname -s)" >&2; exit 1 ;;
esac
LIB="$ROOT/_build/opentui/packages/core/src/zig/lib/${ARCH}-${OS}-static/libopentui.a"

if [[ ! -f "$LIB" ]]; then
  echo "building static OpenTUI via go run..."
  (cd "$ROOT" && go run ./cmd/opentui-static)
fi

SDKROOT="${SDKROOT:-$(xcrun --show-sdk-path)}"
OUT="$DIR/smoke"

if [[ "$(uname -s)" == "Darwin" ]]; then
  if [[ -z "${OPENTUI_KEEP_DEVELOPER_DIR:-}" ]]; then
    export DEVELOPER_DIR=/dev/null
  fi
  zig cc -O2 --sysroot "$SDKROOT" \
    -F"$SDKROOT/System/Library/Frameworks" \
    "$DIR/smoke.c" "$LIB" \
    -lc++ -lpthread \
    -framework CoreFoundation -framework CoreAudio -framework AudioToolbox \
    -o "$OUT"
else
  zig cc -O2 "$DIR/smoke.c" "$LIB" -lc++ -ldl -lpthread -lm -o "$OUT"
fi

echo "built: $OUT ($(ls -lh "$OUT" | awk '{print $5}'))"
echo "deps:"
if [[ "$(uname -s)" == "Darwin" ]]; then
  (
    unset DEVELOPER_DIR
    /usr/bin/otool -L "$OUT" | sed 's/^/  /'
  )
elif command -v ldd >/dev/null; then
  ldd "$OUT" | sed 's/^/  /'
fi

"$OUT"
echo "OK: static smoke passed"

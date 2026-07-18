#!/usr/bin/env bash
# Thin wrapper around: go run ./cmd/hello-tui
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
exec go run ./cmd/hello-tui "$@"

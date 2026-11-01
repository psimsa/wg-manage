#!/bin/sh
# Builds the wg-manage web (wasm + PWA) assets into web/dist/.
# Usage: scripts/build-wasm.sh
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/web/dist"
GOROOT="$(go env GOROOT)"

mkdir -p "$DIST"

echo "==> building wg-manage.wasm"
GOOS=js GOARCH=wasm go build -o "$DIST/wg-manage.wasm" ./wasm/

echo "==> copying runtime + PWA shell"
cp "$GOROOT/lib/wasm/wasm_exec.js" "$DIST/wasm_exec.js"
cp "$ROOT/web/index.html" "$ROOT/web/app.js" "$ROOT/web/sw.js" "$ROOT/web/manifest.webmanifest" "$DIST/"

echo "==> done: $DIST"
ls -la "$DIST"

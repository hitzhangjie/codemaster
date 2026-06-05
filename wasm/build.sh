#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
GOROOT="$(go env GOROOT)"

echo "Building main.wasm..."
GOOS=js GOARCH=wasm go build -o "${ROOT}/main.wasm" "${ROOT}/main.go"

echo "Copying wasm_exec.js..."
cp "${GOROOT}/lib/wasm/wasm_exec.js" "${ROOT}/wasm_exec.js"

echo "Done. Serve this directory with your HTTP server and open index.html"

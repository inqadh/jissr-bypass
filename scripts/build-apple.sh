#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUTPUT="${1:-$ROOT/dist/apple/JissrBypass.xcframework}"

mkdir -p "$(dirname "$OUTPUT")"
cd "$ROOT"
go test ./...
go tool gomobile init
go tool gomobile bind \
  -target=ios,iossimulator,macos \
  -iosversion=15.0 \
  -macosxversion=12.0 \
  -o "$OUTPUT" \
  ./mobile

echo "Apple artifact: $OUTPUT"
shasum -a 256 "$OUTPUT/Info.plist"

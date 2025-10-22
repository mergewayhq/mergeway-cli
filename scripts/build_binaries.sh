#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
DIST_DIR="$ROOT_DIR/dist"
mkdir -p "$DIST_DIR"

VERSION_FILE="$ROOT_DIR/internal/version/version.txt"
VERSION=$(cat "$VERSION_FILE")
COMMIT=$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PKG=github.com/mergewayhq/mergeway-cli
LDFLAGS="-X $PKG/internal/version.Number=$VERSION -X $PKG/internal/version.Commit=$COMMIT -X $PKG/internal/version.BuildDate=$BUILD_DATE"

PLATFORMS=(
  "linux:amd64"
  "linux:arm64"
  "darwin:amd64"
  "darwin:arm64"
)

for platform in "${PLATFORMS[@]}"; do
  IFS=":" read -r GOOS GOARCH <<<"$platform"
  OUTPUT="$DIST_DIR/mw_${GOOS}_${GOARCH}"
  echo "Building $OUTPUT"
  GOMODCACHE="$ROOT_DIR/.cache/go" GOCACHE="$ROOT_DIR/.cache/gobuild" \
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags "$LDFLAGS" -o "$OUTPUT" "$ROOT_DIR"
  chmod +x "$OUTPUT"
done

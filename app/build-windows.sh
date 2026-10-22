#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
out=../dist
db=internal/infrastructure/database/mrav.sqlite

version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)

mkdir -p "$out"
cp ../data/mrav.sqlite "$db"
trap 'rm -f "$db"' EXIT

CC=${CC:-x86_64-w64-mingw32-gcc} GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	go build -trimpath -tags embeddb,no_emoji -ldflags "-s -w -H windowsgui -X main.version=$version" -o "$out/mrav.exe" ./cmd/mrav

if command -v upx >/dev/null; then
	upx --best --lzma --force -q -o "$out/mrav-upx.exe" "$out/mrav.exe" >/dev/null
fi
ls -l "$out"/mrav*.exe

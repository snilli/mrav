#!/usr/bin/env bash
# Builds mrav.exe for Windows from macOS or Linux: one file, with the database built in, that needs nothing beside
# it. Needs mingw-w64 (brew install mingw-w64). With upx on the path it also writes mrav-upx.exe, a packed copy
# about a third of the size; some virus scanners distrust packed programs, so the plain one is the one to trust.
set -euo pipefail

cd "$(dirname "$0")"
out=../dist
db=internal/infrastructure/database/mrav.sqlite # what the embeddb build embeds; the copy is not in git

version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)

mkdir -p "$out"
cp ../data/mrav.sqlite "$db"
trap 'rm -f "$db"' EXIT

# no_emoji leaves out Fyne's 4 MB colour emoji font, which nothing here draws.
CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	go build -trimpath -tags embeddb,no_emoji -ldflags "-s -w -H windowsgui -X main.version=$version" -o "$out/mrav.exe" ./cmd/mrav

if command -v upx >/dev/null; then
	upx --best --lzma --force -q -o "$out/mrav-upx.exe" "$out/mrav.exe" >/dev/null
fi
ls -l "$out"/mrav*.exe

#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
out=../dist
db=internal/infrastructure/database/mrav.sqlite

version=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}

syso=cmd/mrav/zz_icon_windows_amd64.syso
rc=$(mktemp)

mkdir -p "$out"
cp ../data/mrav.sqlite "$db"
trap 'rm -f "$db" "$syso" "$rc"' EXIT

printf '1 ICON "%s"\nGLFW_ICON ICON "%s"\n' "$PWD/icon.ico" "$PWD/icon.ico" > "$rc"
${WINDRES:-x86_64-w64-mingw32-windres} -O coff --input-format rc -i "$rc" -o "$syso"

CC=${CC:-x86_64-w64-mingw32-gcc} GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	go build -trimpath -tags embeddb,no_emoji -ldflags "-s -w -H windowsgui -X main.version=$version" -o "$out/mrav.exe" ./cmd/mrav

if command -v upx >/dev/null; then
	upx --best --lzma --force -q -o "$out/mrav-upx.exe" "$out/mrav.exe" >/dev/null
fi
ls -l "$out"/mrav*.exe

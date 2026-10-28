#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
out=../dist
db=internal/infrastructure/database/mrav.sqlite

version=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}

mkdir -p "$out"
cp ../data/mrav.sqlite "$db"
work=$(mktemp -d)
trap 'rm -f "$db"; rm -rf "$work"' EXIT

app="$out/mrav.app"
rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"

go build -trimpath -tags embeddb,no_emoji -ldflags "-s -w -X main.version=$version" -o "$app/Contents/MacOS/mrav" ./cmd/mrav

iconset="$work/mrav.iconset"
mkdir "$iconset"
for size in 16 32 128 256 512; do
	sips -z "$size" "$size" icon.png --out "$iconset/icon_${size}x${size}.png" >/dev/null
	sips -z "$((size * 2))" "$((size * 2))" icon.png --out "$iconset/icon_${size}x${size}@2x.png" >/dev/null
done
iconutil -c icns "$iconset" -o "$app/Contents/Resources/mrav.icns"

cat > "$app/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key><string>mrav</string>
	<key>CFBundleDisplayName</key><string>mrav</string>
	<key>CFBundleIdentifier</key><string>com.github.snilli.mrav</string>
	<key>CFBundleExecutable</key><string>mrav</string>
	<key>CFBundleIconFile</key><string>mrav</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>$version</string>
	<key>CFBundleVersion</key><string>$version</string>
	<key>LSMinimumSystemVersion</key><string>11.0</string>
	<key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
PLIST

codesign --force --deep --sign - --entitlements debugger.entitlements "$app"
du -sh "$app"

#!/bin/sh
if command -v codesign >/dev/null 2>&1; then
	codesign --force --sign - --entitlements "$(dirname "$0")/debugger.entitlements" "$1"
fi
exec "$@"

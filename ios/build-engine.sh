#!/usr/bin/env bash
# Builds Frameworks/Adkit.xcframework (the shared Go transfer engine) for iOS.
# Needs macOS with Xcode, Go 1.22+ and network access for golang.org/x/mobile.
set -euo pipefail
cd "$(dirname "$0")/.."
export GOFLAGS=-mod=mod
command -v gomobile >/dev/null || go install golang.org/x/mobile/cmd/gomobile@latest
export PATH="$PATH:$(go env GOPATH)/bin"
gomobile init
go get golang.org/x/mobile/bind
rm -rf ios/Frameworks/Adkit.xcframework
gomobile bind -target=ios,iossimulator -iosversion=15.0 -o ios/Frameworks/Adkit.xcframework ./mobile/adkit
git checkout -- go.mod 2>/dev/null || true   # keep the desktop build on Go 1.20
rm -f go.sum
echo "OK: ios/Frameworks/Adkit.xcframework"

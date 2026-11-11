#!/usr/bin/env bash
# Builds every AllDrop release file into ./dist
# Needs Go 1.20.x (the last Go release that still supports Windows 7/8 and macOS 10.13+),
# Python 3 with Pillow, and zip.
set -euo pipefail
cd "$(dirname "$0")"

VERSION="${VERSION:-1.1.2}"
GO="${GO:-go}"
DIST=dist
export CGO_ENABLED=0 GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off

case "$($GO version)" in
  *go1.20.*) ;;
  *) echo "WARNING: use Go 1.20.x for Windows 7/8 + macOS 10.13 support (found: $($GO version))" >&2 ;;
esac

rm -rf "$DIST" && mkdir -p "$DIST"
LDFLAGS="-s -w"

echo "== Windows (7, 8, 8.1, 10, 11)"
GOOS=windows GOARCH=amd64 $GO build -trimpath -ldflags "$LDFLAGS -H windowsgui" -o "$DIST/AllDrop-Windows-64bit.exe" .
GOOS=windows GOARCH=386   $GO build -trimpath -ldflags "$LDFLAGS -H windowsgui" -o "$DIST/AllDrop-Windows-32bit.exe" .
GOOS=windows GOARCH=arm64 $GO build -trimpath -ldflags "$LDFLAGS -H windowsgui" -o "$DIST/AllDrop-Windows-ARM64.exe" .

echo "== macOS (10.13 High Sierra and newer, Intel + Apple Silicon)"
mkdir -p build/mac
GOOS=darwin GOARCH=amd64 $GO build -trimpath -ldflags "$LDFLAGS" -o build/mac/alldrop-amd64 .
GOOS=darwin GOARCH=arm64 $GO build -trimpath -ldflags "$LDFLAGS" -o build/mac/alldrop-arm64 .
python3 scripts/macpkg.py build/mac/alldrop-amd64 build/mac/alldrop-arm64 assets/mac-1024.png build/mac "$VERSION"
if command -v codesign >/dev/null 2>&1; then
  codesign --force --deep --sign - build/mac/AllDrop.app
fi
cp packaging/macos-first-run.txt "build/mac/macOS - ilk acilis (first run).txt"
(cd build/mac && rm -f ../../$DIST/AllDrop-macOS.zip && zip -qry ../../$DIST/AllDrop-macOS.zip AllDrop.app "macOS - ilk acilis (first run).txt")

echo "== Linux"
GOOS=linux GOARCH=amd64 $GO build -trimpath -ldflags "$LDFLAGS" -o "$DIST/AllDrop-Linux-x64" .
GOOS=linux GOARCH=arm64 $GO build -trimpath -ldflags "$LDFLAGS" -o "$DIST/AllDrop-Linux-ARM64" .

echo "== Android engine (packed into the APK as native libraries)"
JNI=android/app/src/main/jniLibs
mkdir -p $JNI/arm64-v8a $JNI/armeabi-v7a $JNI/x86_64 $JNI/x86
GOOS=linux GOARCH=arm64         $GO build -trimpath -ldflags "$LDFLAGS" -o $JNI/arm64-v8a/liballdrop.so .
GOOS=linux GOARCH=arm GOARM=7   $GO build -trimpath -ldflags "$LDFLAGS" -o $JNI/armeabi-v7a/liballdrop.so .
GOOS=linux GOARCH=amd64         $GO build -trimpath -ldflags "$LDFLAGS" -o $JNI/x86_64/liballdrop.so .
GOOS=linux GOARCH=386           $GO build -trimpath -ldflags "$LDFLAGS" -o $JNI/x86/liballdrop.so .

(cd "$DIST" && sha256sum * > SHA256SUMS.txt 2>/dev/null || shasum -a 256 * > SHA256SUMS.txt)
ls -la "$DIST"

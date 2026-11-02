#!/usr/bin/env python3
"""Build AllDrop.app from two Go binaries without needing a Mac.

usage: macpkg.py <amd64-binary> <arm64-binary> <icon-1024.png> <out-dir> <version>

Creates <out-dir>/AllDrop.app with a universal (Intel + Apple Silicon)
executable, an .icns icon and an Info.plist. On a Mac (CI), run
`codesign --force --deep -s - AllDrop.app` afterwards for a full ad-hoc
bundle signature; the arm64 slice is already linker-signed by Go.
"""
import io
import os
import shutil
import struct
import sys

from PIL import Image


def fat_binary(paths, out):
    slices = []
    for p in paths:
        data = open(p, "rb").read()
        magic = struct.unpack("<I", data[:4])[0]
        if magic != 0xFEEDFACF:
            raise SystemExit(f"{p} is not a 64-bit Mach-O")
        cputype, cpusub = struct.unpack("<ii", data[4:12])
        slices.append((cputype, cpusub & 0x00FFFFFF, data))
    align = 14
    offset = 1 << align
    header = struct.pack(">II", 0xCAFEBABE, len(slices))
    entries, blobs = b"", []
    for cputype, cpusub, data in slices:
        entries += struct.pack(">iiIII", cputype, cpusub, offset, len(data), align)
        blobs.append((offset, data))
        offset += len(data)
        offset = (offset + (1 << align) - 1) & ~((1 << align) - 1)
    with open(out, "wb") as f:
        f.write(header + entries)
        for off, data in blobs:
            f.seek(off)
            f.write(data)
    os.chmod(out, 0o755)


def icns(png1024, out):
    src = Image.open(png1024).convert("RGBA")
    types = [("icp4", 16), ("icp5", 32), ("icp6", 64), ("ic07", 128), ("ic08", 256),
             ("ic09", 512), ("ic10", 1024), ("ic11", 32), ("ic12", 64), ("ic13", 256), ("ic14", 512)]
    body = b""
    for t, s in types:
        buf = io.BytesIO()
        src.resize((s, s), Image.LANCZOS).save(buf, "PNG")
        png = buf.getvalue()
        body += t.encode() + struct.pack(">I", len(png) + 8) + png
    with open(out, "wb") as f:
        f.write(b"icns" + struct.pack(">I", len(body) + 8) + body)


PLIST = """<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDevelopmentRegion</key><string>tr</string>
  <key>CFBundleExecutable</key><string>AllDrop</string>
  <key>CFBundleIconFile</key><string>AppIcon</string>
  <key>CFBundleIdentifier</key><string>app.alldrop.desktop</string>
  <key>CFBundleInfoDictionaryVersion</key><string>6.0</string>
  <key>CFBundleName</key><string>AllDrop</string>
  <key>CFBundleDisplayName</key><string>AllDrop</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>{v}</string>
  <key>CFBundleVersion</key><string>{v}</string>
  <key>LSMinimumSystemVersion</key><string>10.13</string>
  <key>LSUIElement</key><true/>
  <key>NSHighResolutionCapable</key><true/>
  <key>NSLocalNetworkUsageDescription</key><string>AllDrop finds nearby devices and sends files directly over your local network.</string>
  <key>NSHumanReadableCopyright</key><string>AllDrop</string>
</dict>
</plist>
"""


def main():
    amd, arm, icon, outdir, ver = sys.argv[1:6]
    app = os.path.join(outdir, "AllDrop.app")
    shutil.rmtree(app, ignore_errors=True)
    os.makedirs(os.path.join(app, "Contents", "MacOS"))
    os.makedirs(os.path.join(app, "Contents", "Resources"))
    fat_binary([amd, arm], os.path.join(app, "Contents", "MacOS", "AllDrop"))
    icns(icon, os.path.join(app, "Contents", "Resources", "AppIcon.icns"))
    with open(os.path.join(app, "Contents", "Info.plist"), "w") as f:
        f.write(PLIST.replace("{v}", ver))
    with open(os.path.join(app, "Contents", "PkgInfo"), "w") as f:
        f.write("APPL????")
    print(app)


if __name__ == "__main__":
    main()

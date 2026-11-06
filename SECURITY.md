# AllDrop Security

AllDrop moves files **directly between devices on the same local network**. No
account, no cloud, no server. This document describes the threat model, the
defenses in the code, and the automated attack tests that verify them.

## Threat model

AllDrop assumes the local network is **hostile**: other hosts may send crafted
UDP/TLS/HTTP traffic, spoof device names, replay packets, or try to read files
they were not offered. The owner's own machine is trusted; other users/apps on
the same machine are treated as untrusted on mobile.

| Asset | Protected against |
|---|---|
| File contents in transit | Eavesdropping / MITM on the LAN |
| The receive folder | Path traversal, overwrite, disk-fill |
| A transfer | Being accepted by the wrong device / without consent |
| Device identity | Impersonation / spoofing |
| The local web UI | CSRF, DNS rebinding, XSS, cross-origin reads |

## Defenses (by layer)

### Device-to-device transport — `engine/tlsid.go`
- Every install generates its own **P-256 key pair + self-signed certificate**.
- A device's **ID is the SHA-256 of its certificate** (`certID`), so an identity
  cannot be forged without the private key.
- App-to-app traffic uses **mutually-authenticated TLS 1.2+**. Self-signed certs
  are pinned: `InsecureSkipVerify` + a `VerifyConnection` callback that checks the
  peer's cert hash against the **expected** ID. HTTP/2 is disabled on both ends.
- UDP discovery is only a hint — a peer is never trusted until its identity is
  **proven over TLS** (`verifyAndAdd` → `/peer/info` → cert-hash match).

### Consent & the 6-digit code — `engine/transfer.go`
- Each incoming transfer shows a **crypto-random 6-digit code** (`crypto/rand`)
  on the **receiver**. The sender must type it; nothing is written before that.
- Codes are compared with **`subtle.ConstantTimeCompare`** (no timing oracle).
- **5 wrong tries** (including malformed input) decline the request; a declined
  request refuses even the correct code afterwards. Pending offers expire in 3 min.
- Per-source-IP cap of 8 pending offers (anti-flood).

### Safe file handling — `engine/util.go`, `engine/transfer.go`
- `sanitizeRel` / `sanitizeName` strip control chars and `<>:"/\|?*`, reject
  `.`/`..`, neutralize Windows reserved names, cap length and folder depth (32).
- Received bytes are **length-checked against the declared size**; an oversize
  body is aborted and the partial file removed.
- Files are written `O_EXCL` with collision-free unique names, always **inside**
  the receive directory.

### Local web UI — `engine/http.go`
- Owner access requires a **loopback client AND a literal `localhost`/`127.0.0.1`
  Host header** → defeats DNS rebinding.
- State-changing requests require `POST` + an `X-AD: 1` header and a same-origin
  `Origin` → defeats cross-site request forgery.
- Browser guests authenticate with a secret key from the QR code (constant-time
  compared); cookies are `HttpOnly` + `SameSite`.
- On mobile, the local UI is additionally gated by a per-launch **owner token**
  so other apps on the phone cannot drive the engine.
- Security headers: `X-Content-Type-Options`, `X-Frame-Options: DENY`,
  `Referrer-Policy: no-referrer`. All request bodies are `io.LimitReader`-bounded.
- The web UI escapes **all** untrusted strings (device names, file names, paths)
  through a single `esc()` before inserting them into the DOM → no XSS.

### Mobile
- **iOS:** App Transport Security uses `NSAllowsLocalNetworking` only (no
  arbitrary cleartext). The WKWebView only loads `127.0.0.1`; the native
  "open file" bridge refuses any path outside the app's Documents folder.
- **Android:** cleartext is **restricted to loopback** via
  `network_security_config.xml`; the WebView has `allowFileAccess=false` and its
  JS bridge is only reachable from the local engine page.

### Supply chain
- The only third-party dependency, [`go-qrcode`](https://github.com/skip2/go-qrcode)
  (MIT), is **vendored** in `third_party/` and `replace`d in `go.mod`. Release
  builds run with `GOPROXY=off` — no code is fetched at build time.

## Automated attack tests

`engine/security_test.go` runs real attacks against the defenses:

```
go test ./engine/ -run Attack -v
```

Covered: path-traversal / filename-injection into the receive folder, oversize
upload rejection + cleanup, 6-digit code brute-force lockout, TLS identity
uniqueness, `validFiles` bounds, Content-Disposition header injection,
CSRF / foreign-Origin / DNS-rebind rejection on the web server, anonymous
(no-client-cert) peer rejection, and a **full two-device encrypted transfer**
proving the wrong code is refused over the wire and the file only arrives after
the correct one. `go vet ./...` is clean.

## Hardening notes / known trade-offs

- **Desktop owner access** is granted to any loopback client (no owner token),
  which is the standard model for a localhost app and keeps "open in browser"
  working. On a shared desktop, other local users' processes could reach the UI;
  single-user desktops are the assumed trust boundary.
- **Legacy desktop builds** are compiled with Go 1.20.14 (the last Go supporting
  Windows 7/8 and macOS 10.13). The app disables HTTP/2, which neutralizes the
  notable post-1.20 `net/http` CVEs. For modern-only OSes a parallel build with
  the latest Go toolchain is recommended to pick up current stdlib fixes.

## Reporting a vulnerability

Open a private security advisory on the GitHub repository, or an issue for
non-sensitive reports.

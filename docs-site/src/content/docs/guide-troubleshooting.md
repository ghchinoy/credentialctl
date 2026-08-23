---
title: Troubleshooting & CGO
description: Common installation issues, CGO configurations, and supported media formats.
---

This guide addresses common setup hurdles, native library linking, and format compatibility.

## Dynamic Library Lookup on macOS

### The Error
```text
dyld: Library not loaded: @rpath/libcredentio_c.dylib
  Reason: no LC_RPATH's found
```

### The Cause
macOS System Integrity Protection (SIP) prevents child processes and subshells from inheriting `DYLD_LIBRARY_PATH`. When Go compiles binaries without embedded runtime paths (`rpath`), the dynamic linker fails to locate `libcredentio_c.dylib`.

### The Solution
Ensure your build environment enables CGO and uses embedded rpaths:
```bash
CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/credentialctl main.go
```
The Go bindings in `credentio-contributions` embed `-Wl,-rpath,${SRCDIR}/lib` in `#cgo darwin LDFLAGS`.

---

## CGO Compilation Requirements

Because `credentialctl` links directly against the native C-ABI wrapper:
- Always set `CGO_ENABLED=1` when running `go build`, `go test`, or `go run`.
- Verify that clang or gcc is available in your shell environment (`xcode-select --install` on macOS).

---

## Supported Media Formats

`credentialctl` discovers and parses the following media formats based on file signatures and extensions:

| Format Category | Supported Extensions | IANA MIME Type |
|---|---|---|
| JPEG | `.jpg`, `.jpeg` | `image/jpeg` |
| PNG | `.png` | `image/png` |
| WebP | `.webp` | `image/webp` |
| AVIF | `.avif` | `image/avif` |
| High Efficiency | `.heic`, `.heif` | `image/heic` |
| TIFF | `.tif`, `.tiff` | `image/tiff` |
| MP4 Video | `.mp4` | `video/mp4` |
| QuickTime | `.mov` | `video/quicktime` |
| MPEG Audio | `.mp3` | `audio/mpeg` |
| MP4 Audio | `.m4a` | `audio/mp4` |
| WAV Audio | `.wav` | `audio/wav` |
| C2PA Stores | `.c2pa` | `application/x-c2pa-manifest-store` |

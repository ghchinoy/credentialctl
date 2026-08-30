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
Compile using `make build` (or `make fetch-credentio-lib`), which automatically injects required CGO include flags and `@loader_path` rpaths:
```bash
make build
```
`credentialctl` links against the prebuilt dynamic library in `third_party/credentio/lib/` and embeds `@loader_path`, allowing the compiled binary to locate `libcredentio_c.dylib` directly beside the executable.

---

## Benign Linker Warnings During Build

When building with `make build`, the linker may output informational warnings:

```text
ld: warning: duplicate -rpath '@loader_path' ignored
ld: warning: duplicate -rpath '.../third_party/credentio/lib' ignored
ld: warning: ignoring duplicate libraries: '-lcredentio_c'
ld: warning: search path '.../go@v0.1.5/lib' not found
ld: warning: search path '.../go@v0.1.5/../native' not found
```

These warnings are harmless:
- **Duplicate flags:** The Makefile explicitly prepends required Credentio flags, and the linker safely deduplicates them.
- **Module cache paths:** The upstream Go module defines development search paths (`${SRCDIR}/lib`) that are inert when consuming prebuilt releases from the Go module cache. The library is resolved from `third_party/credentio/lib/`.

---

## CGO Compilation Requirements

Because `credentialctl` links directly against the native C-ABI wrapper:
- Always set `CGO_ENABLED=1` when running `go build`, `go test`, or `go run` (handled automatically by `Makefile` targets).
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

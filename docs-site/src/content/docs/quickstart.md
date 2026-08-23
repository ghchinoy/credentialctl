---
title: Quick Start
description: Get up and running with credentialctl in minutes.
---

`credentialctl` provides both automated command-line validation and an interactive terminal interface for inspecting Coalition for Content Provenance and Authenticity (C2PA) credentials.

## Prerequisites

Before building `credentialctl`, ensure your system meets the following requirements:
- **Go:** Version 1.26 or newer.
- **CGO:** Enabled (`CGO_ENABLED=1`).
- **Native Library:** The `libcredentio_c.dylib` (macOS) or `libcredentio_c.so` (Linux) shared library staged in `../credentio-contributions/go/lib`.

## Installation

Clone the repository and compile the binary:

```bash
git clone https://github.com/ghchinoy/credentialctl.git
cd credentialctl
make build
```

The executable is generated at `bin/credentialctl`.

## Your First Validation

Validate any JPEG, PNG, WebP, or MP4 file on disk:

```bash
./bin/credentialctl validate photo.jpg
```

Output:
```text
  C2PA VALIDATION REPORT  

  Asset:           photo.jpg (1.20 MB, image/jpeg)
  Path:            /path/to/photo.jpg
  Status:          [✓ SIGNED]
  Generator:       Google C2PA Core Generator Library
  Signer:          Google C2PA Media Services
  Format/Spec:     image/jpeg (C2PA 2.2)
  Assertions:      3 attached
  Validation:      10 reported
  Wall Time:       4.61 ms
```

## Launching the Interactive TUI

To launch the full-screen terminal interface on your current folder or a specific directory:

```bash
./bin/credentialctl tui /path/to/media
```

Use arrow keys to navigate files, press `Enter` to drill into any asset's manifest, and press `Tab` to filter signed vs unsigned media.

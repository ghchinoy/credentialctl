# credentialctl

A command-line utility and interactive terminal interface to validate and inspect Coalition for Content Provenance and Authenticity (C2PA) Content Credentials in media files.

`credentialctl` validates manifests, digital signatures, certificate chains, and assertions in images, videos, and audio. It is powered by Google Credentio native C-ABI bindings.

---

## Table of Contents
- [Quick Installation](#quick-installation)
- [Immediate Usage](#immediate-usage)
- [Local Development Setup](#local-development-setup)
- [Publish and Release Process](#publish-and-release-process)
- [Contributing](#contributing)
- [Features and TUI Controls](#features-and-tui-controls)
- [Command Reference and Exit Codes](#command-reference-and-exit-codes)
- [Documentation](#documentation)
- [License](#license)

---

## Quick Installation

### Install via Homebrew Tap (macOS and Linux)

Install `credentialctl` using the official Homebrew tap:

```bash
# Add the tap and install
brew tap ghchinoy/tap
brew install credentialctl
```

Or install directly in a single command:

```bash
brew install ghchinoy/tap/credentialctl
```

Verify your installation and runtime engine version:

```bash
credentialctl version --json
```

### Build from Source

You can also compile `credentialctl` directly from source:

```bash
git clone https://github.com/ghchinoy/credentialctl.git
cd credentialctl
make build
```

The build process automatically downloads the prebuilt Credentio C-ABI shared library via `scripts/fetch-credentio-lib.sh` and outputs the executable binary to `bin/credentialctl`.

---

## Immediate Usage

Validate any media file with a single command:

```bash
credentialctl validate sample.jpg
```

Output:
```text
  C2PA VALIDATION REPORT  

  Asset:           sample.jpg (1.20 MB, image/jpeg)
  Path:            /path/to/sample.jpg
  Status:          [✓ SIGNED]
  Generator:       Google C2PA Core Generator Library
  Signer:          Google C2PA Media Services
  Format/Spec:     image/jpeg (C2PA 2.2)
  Assertions:      3 attached
  Validation:      10 reported
  Wall Time:       4.61 ms
```

### Structured JSON Output for Agents and Pipelines

Output structured JSON data for automated scripts and coding agents:

```bash
credentialctl validate sample.jpg --json | jq .
```

### Raw Native Engine crJSON

Extract the raw C2PA manifest JSON payload directly from the underlying engine:

```bash
credentialctl validate sample.jpg --raw
```

### Deep Manifest Inspection

Inspect detailed assertions, signing certificates, and validation check items:

```bash
credentialctl inspect artwork.png
```

### Batch Directory Validation

Scan and validate all media assets in a directory:

```bash
credentialctl folder ./media
```

### Interactive TUI Launcher

Launch the full-screen terminal interface to browse folders and inspect credentials:

```bash
credentialctl tui ./media
```

### Version and Runtime Engine Observability

Display CLI build metadata and runtime C-ABI engine information:

```bash
credentialctl version
```

---

## Local Development Setup

### Prerequisites
- Go 1.22 or newer
- CGO enabled (`CGO_ENABLED=1`) with a local C toolchain (Clang or GCC)
- Prebuilt `libcredentio_c` shared library (staged automatically via `make fetch-credentio-lib`)

### Development Commands

Stage native dependencies, build binaries, and run the test suite:

```bash
# Fetch prebuilt C-ABI shared libraries and headers
make fetch-credentio-lib

# Build the executable binary with embedded dynamic RPATHs
make build

# Run unit, integration, and E2E test suites
make test

# Launch the interactive interface directly from source
make run
```

---

## Publish and Release Process

`credentialctl` uses GoReleaser and GitHub Actions to automate multi-platform binary compilation, dynamic library packaging, and Homebrew tap formula distribution.

### Local Snapshot Release Testing

Maintainers can test packaging and archive generation locally:

```bash
# Package local release snapshot archives in dist/
make release-snapshot
```

### Automated Production Releases

Pushing a semantic version tag (such as `v0.1.5`) triggers `.github/workflows/release.yaml` to build release artifacts, generate SHA-256 checksums, publish GitHub Releases, and update the formula in `ghchinoy/homebrew-tap`.

For full release runbooks, secret configuration, and rollback procedures, consult the [Release Guide](docs/RELEASING.md).

---

## Contributing

Contributions are welcome. Please open an issue first to discuss substantial architectural modifications or new subcommand designs.

### Pre-Submission Checklist

Before submitting a pull request, ensure all changes pass verification:

1. Run the test suite: `make test`
2. Compile and verify binary linkage: `make build && otool -L bin/credentialctl`
3. Test release snapshot packaging: `make release-snapshot`
4. Confirm documentation adheres to house style rules (zero em dashes in prose, active voice)

---

## Features and TUI Controls

`credentialctl` provides both automated CLI commands and an interactive Bubble Tea terminal user interface.

### Interactive TUI Navigation

Launch the full-screen interface on any target directory:

```bash
credentialctl tui /path/to/media
```

| Keybinding | Action |
|---|---|
| `↑` / `↓` or `k` / `j` | Navigate media table rows |
| `Enter` | Open deep File Inspector for selected asset |
| `Tab` / `f` | Cycle status filters: All, Signed Only, Unsigned Only, Invalid Only |
| `r` | Rescan directory and update validation counters |
| `1` - `5` | Switch Inspector tabs: Overview, Assertions, Signature, Validation, Raw JSON |
| `PgUp` / `PgDn` | Scroll viewport content |
| `Esc` / `Backspace` | Return from Inspector to Folder Review |
| `q` / `Ctrl+C` | Quit application |

---

## Command Reference and Exit Codes

Automation scripts and agent workers can inspect process exit codes to determine asset authenticity:

| Exit Code | Meaning | Description |
|---|---|---|
| `0` | `SIGNED` | Valid C2PA content credentials present and verified |
| `1` | `UNSIGNED` | No C2PA credentials found in asset |
| `2` | `INVALID` | Corrupted, invalid, or failing credential manifest |
| `3` | `ERROR` | Input error, unsupported media format, or file not found |

---

## Documentation

- [User Guide](docs/user_guide.md): Forensic CLI workflows, TUI inspection tabs, JSON schemas, and troubleshooting.
- [Release Guide](docs/RELEASING.md): Release lifecycle, GitHub Actions automation, Homebrew tap updates, and rollback procedures.

---

## License

Apache 2.0. Copyright 2026 Google LLC.

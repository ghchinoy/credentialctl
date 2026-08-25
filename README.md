# credentialctl

A command-line tool and interactive terminal interface to validate and inspect C2PA content credentials in media files.

`credentialctl` validates manifests, digital signatures, certificate chains, and assertions in images, videos, and audio. It is powered by Google Credentio native C-ABI bindings.

---

## Table of Contents
- [Quick Installation](#quick-installation)
- [Immediate Usage](#immediate-usage)
- [Local Development Setup](#local-development-setup)
- [Publish and Release Process](#publish-and-release-process)
- [Contributing](#contributing)
- [Features and TUI Controls](#features-and-tui-controls)
- [Documentation](#documentation)
- [License](#license)

---

## Quick Installation

Build the binary directly from source (downloads the prebuilt Credentio C-ABI library automatically):

```bash
git clone https://github.com/ghchinoy/credentialctl.git
cd credentialctl
make build
```

The compiled executable is placed at `bin/credentialctl`.

---

## Immediate Usage

Validate any media file with a single command:

```bash
./bin/credentialctl validate sample.jpg
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

To output structured JSON for agent and pipeline automation:

```bash
./bin/credentialctl validate sample.jpg --json
```

To output raw native engine crJSON directly:

```bash
./bin/credentialctl validate sample.jpg --raw
```

---

## Local Development Setup

### Prerequisites
- Go 1.22 or newer
- CGO enabled (`CGO_ENABLED=1`)
- Prebuilt `libcredentio_c` (downloaded automatically via `make fetch-credentio-lib` from GitHub Releases)

### Development Commands

Clone the repository and run the test suite:

```bash
git clone https://github.com/ghchinoy/credentialctl.git
cd credentialctl
make test
```

Launch the interactive interface directly from source:

```bash
make run
```

---

## Publish and Release Process

Maintainers package release binaries using GoReleaser:

```bash
# Build local snapshot archives
make release-snapshot
```

Artifacts and checksums are generated in the `dist/` directory.

---

## Contributing

Pull requests and issue reports are welcome. For major architectural changes or new command additions, please open an issue first to discuss the design.

Ensure all unit tests pass before submitting changes:

```bash
make test
```

---

## Features and TUI Controls

`credentialctl` provides both automated CLI commands and an interactive Bubble Tea terminal user interface.

### Interactive TUI Navigation

Launch the full-screen interface on any folder:

```bash
./bin/credentialctl tui /path/to/media
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

## Documentation

For comprehensive command documentation, JSON schema definitions, and workflow recipes, consult the [User Guide](docs/user_guide.md).

---

## License

Apache 2.0. Copyright 2026 Google LLC.

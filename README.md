# CredentialCTL (`credentialctl`)

`credentialctl` is an **Agent-Aware** CLI and interactive [Bubble Tea](https://github.com/charmbracelet/bubbletea) terminal user interface for validating and deeply inspecting [C2PA](https://c2pa.org/) (Coalition for Content Provenance and Authenticity) manifests in media assets.

Built on top of Google Credentio high-performance native C-ABI bindings via [`credentio-contributions`](https://github.com/ghchinoy/credentio-contributions).

---

## Features

- 📁 **Folder Review Mode**: Rapidly scan directories for supported media assets (`.jpg`, `.jpeg`, `.png`, `.webp`, `.avif`, `.heic`, `.mp4`, `.mov`, etc.), batch validating credentials in milliseconds.
- 🔍 **File Inspector Mode**: Tabbed drill-down into asset metadata, assertion lists (`c2pa.actions`, data hashes, thumbnails), cryptographic signatures & certificate authorities, validation status codes, and raw `crJSON`.
- 🤖 **Agent-Aware CLI Design**:
  - Logical Cobra command groups (`validation`, `inspection`, `interactive`).
  - Comprehensive documentation with `Short`, `Long`, and executable `Example` blocks.
  - Machine-parsable `--json` outputs for automation pipelines and agent workflows.
  - Proactive, actionable hints in error messages.
  - Semantic Lipgloss styling (`Accent`, `Pass`, `Warn`, `Fail`, `Muted`, `ID`).
- ⚡ **Interactive Terminal UI**: Full-screen responsive Bubble Tea interface with live filter cycling (`All`, `Signed Only`, `Unsigned Only`, `Invalid Only`), progress spinners, and keyboard-driven navigation.

---

## Installation & Prerequisites

### Requirements
- **Go**: Version `1.26+`
- **CGO**: `CGO_ENABLED=1`
- **Native Library**: Built `libcredentio_c.dylib` (macOS) or `libcredentio_c.so` (Linux) staged in `../credentio-contributions/go/lib`.

### Building
```bash
# Clone and build
git clone https://github.com/ghchinoy/credentialctl.git
cd credentialctl
make build
```

Binary will be produced at `bin/credentialctl`.

---

## CLI Usage

### 1. Interactive TUI
```bash
# Launch interactive TUI for current folder
credentialctl

# Open TUI on a specific media folder
credentialctl tui /path/to/media

# Open TUI recursively scanning subdirectories
credentialctl tui /path/to/media -r
```

**TUI Keybindings:**
- `↑` / `↓` or `k` / `j`: Move selection up and down.
- `Enter`: Inspect selected file in the File Inspector.
- `Tab` / `f`: Cycle status filters (`All` → `Signed Only` → `Unsigned Only` → `Invalid Only`).
- `r`: Rescan directory.
- `1` - `5` or `Left` / `Right`: Switch tabs in File Inspector (`Overview`, `Assertions`, `Signature`, `Validation`, `Raw JSON`).
- `PgUp` / `PgDn`: Scroll inspection views.
- `Esc` / `Backspace`: Return to Folder Review view.
- `q` / `Ctrl+C`: Quit application.

### 2. Validate a Single Asset
```bash
# Human-readable formatted card
credentialctl validate image.png

# Machine-parsable JSON output
credentialctl validate image.png --json
```

**Exit Codes:**
- `0`: Asset is authentic and signed (`SIGNED`).
- `1`: Asset contains no C2PA credentials (`UNSIGNED`).
- `2`: Asset contains an invalid or failing signature/manifest (`INVALID`).

### 3. Scan a Directory
```bash
# Scan directory non-recursively
credentialctl folder ./photos

# Scan directory recursively
credentialctl folder ./photos --recursive

# Output batch scan summary as JSON
credentialctl folder ./photos --json
```

### 4. Deep Inspection
```bash
# Full manifest breakdown in terminal
credentialctl inspect photo.jpg

# Output formatted or raw crJSON
credentialctl inspect photo.jpg --raw
credentialctl inspect photo.jpg --json

# Open directly into the interactive TUI inspector
credentialctl inspect photo.jpg --tui
```

---

## License
Apache 2.0 - Copyright 2026 Google LLC.

---
title: Architecture & Design
description: Technical architecture, engine layers, and semantic styling in credentialctl.
---

`credentialctl` is organized into three decoupled layers: the Command-Line Interface (CLI), the Interactive Terminal User Interface (TUI), and the Core Engine.

## System Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│                      credentialctl                          │
├──────────────────────────────┬──────────────────────────────┤
│       Cobra CLI Layer        │     Bubble Tea TUI Layer     │
│  - validate [file] (--json)  │  - Folder Review Model       │
│  - folder [dir] (--json, -r) │  - File Inspector Model      │
│  - inspect [file] (--json)   │  - App Model Router          │
├──────────────────────────────┴──────────────────────────────┤
│                     Core Engine Layer                       │
│  - Media Scanner (MIME Discovery)                           │
│  - Thread-Safe Validator Service                            │
├─────────────────────────────────────────────────────────────┤
│                 Credentio C-ABI Bindings                    │
│  - github.com/ghchinoy/credentio-contributions/go           │
│  - libcredentio_c.dylib / libcredentio_c.so (BoringSSL)     │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Command-Line Interface (`cmd/`)
Built with [Cobra](https://github.com/spf13/cobra) following Agent-Aware design guidelines:
- **Logical Groups:** Organizes commands into `validation`, `inspection`, and `interactive` categories.
- **Tri-Part Documentation:** Every command provides `Short`, `Long`, and practical `Example` documentation blocks.
- **Automation Ready:** Every data command accepts `--json` for machine consumption.

### 2. Interactive Terminal UI (`internal/ui/`)
Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss):
- **Folder Review View:** Renders responsive tables, live progress spinners, and status filter bars.
- **File Inspector View:** Provides tabbed navigation and scrollable viewports for analyzing manifests.
- **App Model:** Manages top-level state routing between folder navigation and file inspection while handling terminal resize events.

### 3. Core Engine (`internal/engine/`)
- **Scanner:** Discovers supported media files (`.jpg`, `.jpeg`, `.png`, `.webp`, `.avif`, `.heic`, `.mp4`, `.mov`, `.mp3`, `.wav`, `.c2pa`) and maps MIME types.
- **Validator Service:** Protects native C-ABI pointers with mutex synchronization and provides batch processing routines.

## Semantic Color Palette

The user interface uses a consistent semantic color palette rather than arbitrary colors:

| Semantic Role | Hex Value | Visual Meaning |
|---|---|---|
| `Accent` | `#7D56F4` | Header titles, active tabs, and primary action text |
| `Pass` | `#04B575` | Authentic signatures (`✓ SIGNED`) and valid hash checks |
| `Warn` | `#FFB300` | Unsigned assets (`⊘ UNSIGNED`) and validation warnings |
| `Fail` | `#FF4D4D` | Corrupted manifests (`✕ INVALID`) and validation failures |
| `Muted` | `#626262` | Table dividers, borders, and execution timestamps |
| `ID` | `#00D7D7` | Manifest URNs, Instance IDs, and MIME types |

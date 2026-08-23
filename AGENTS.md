# AGENTS.md: CredentialCTL Developer & Agent Guide

This document outlines the architecture, Agent-Aware design standards, and common development workflows for agents and developers contributing to `credentialctl`.

---

## 1. Architecture Overview

`credentialctl` is a terminal user interface and command-line utility for C2PA content credential validation and inspection:

- **`cmd/`**: Cobra command definitions.
  - `root.go`: Base command, global flags, and automatic TUI fallback.
  - `validate.go`: Single-file validation with human or machine (`--json`) output.
  - `folder.go`: Batch directory scanning with summary tables and JSON serialization.
  - `inspect.go`: Deep C2PA manifest, assertion, and cryptographic signature analysis.
  - `tui.go`: Full-screen interactive terminal interface launcher.
- **`internal/engine/`**: Core validation logic and media discovery.
  - `models.go`: File item data models and aggregate folder statistics.
  - `scanner.go`: Recursive directory traversal and MIME type detection for supported formats.
  - `validator.go`: Thread-safe wrapper around Google Credentio C-ABI bindings.
- **`internal/ui/`**: Interactive Bubble Tea interface.
  - `theme/theme.go`: Semantic Lipgloss color palette and status badges.
  - `folderview/folderview.go`: Media asset table with real-time scan spinner and filter cycling.
  - `inspectview/inspectview.go`: Multi-tab manifest inspector with scrollable viewports.
  - `app.go`: Root model managing view switching and window resize events.

---

## 2. Agent-Aware CLI Design Mandates

All commands and interfaces in `credentialctl` must follow these conventions:

1. **Logical Grouping:** Group subcommands using Cobra `GroupID` (`validation`, `inspection`, `interactive`).
2. **Tri-Part Documentation:** Every command must provide comprehensive `Short`, `Long`, and executable `Example` documentation.
3. **Structured JSON:** All commands outputting data must support the `--json` flag for pipeline and agent consumption.
4. **Proactive Error Hints:** Error messages must provide actionable recommendations and hints (e.g. suggesting valid file paths or supported extensions).
5. **Semantic Lipgloss Styling:** Use the semantic palette (`Accent`, `Pass`, `Warn`, `Fail`, `Muted`, `ID`) rather than arbitrary colors.

---

## 3. Common Development & Build Commands

```bash
# Build the binary
make build

# Run full test suite
make test

# Run interactive TUI directly from source
make run

# Clean build artifacts
make clean
```

> **Note:** Building and testing requires `CGO_ENABLED=1` and the native library `libcredentio_c` staged in `../credentio-contributions/go/lib`.

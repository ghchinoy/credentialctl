# AGENTS.md: CredentialCTL Developer & Agent Guide

This document outlines the architecture, Agent-Aware design standards, documentation quality workflows, and common development tasks for agents and developers contributing to `credentialctl`.

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

# Snapshot release packaging
make release-snapshot
```

> **Note:** Building and testing requires `CGO_ENABLED=1` and the native library `libcredentio_c` staged in `../credentio-contributions/go/lib`.

---

## 4. Documentation Quality & Editorial Gating

All documentation in this repository must pass strict quality and readability audits:

### 1. Mark Allen README Quality Rubric
Whenever editing `README.md`, adhere to Mark Allen's 8-part checklist:
1. Clear Project Description (opening one-sentence summary).
2. Quick Installation Instructions (exact copy-pasteable build commands).
3. Immediate Usage Example (working command and sample output in the first screenful).
4. Local Development Setup (prerequisites and test commands).
5. Publish / Deploy Process (release build commands).
6. Encourage Contributions (clear stance and verification steps).
7. Use Markdown Well (consistent hierarchy, language-hinted code blocks).
8. Optional Extras (TUI keybindings table, TOC, links to docs).

Target quality score: **≥ 34 / 40 (Excellent band)**.

### 2. Technical Writing Editorial Standards
Technical prose in `README.md` and `docs/` must follow house style:
- **Zero em dashes in prose:** Use commas, colons, or parentheses.
- **Active voice with named actors:** Avoid passive voice and false agency.
- **No throat-clearing openers:** Cut phrases like "Here's what we found:" or "It is worth noting that".
- **Direct statements:** Avoid rhetorical Wh- starters and binary contrast frames ("Not X, it's Y").

### 3. Automated Docstats Audit
Validate prose using the local `docstats` analyzer:
```bash
uv run --directory ../docstats python -c "
import sys, os
sys.path.insert(0, os.path.abspath('../docstats'))
from metrics import _sync_analyze_document

files = ['README.md', 'docs/user_guide.md']
for f in files:
    with open(f, 'r', encoding='utf-8') as fp:
        raw = fp.read()
    res = _sync_analyze_document(raw, f)
    print(f'{f}: Grade={res.readability.flesch_kincaid_grade:.1f}, Style Score={res.ai_patterns.ai_tell_score:.1f}/10, EmDashes={res.ai_patterns.em_dash_count}')
"
```
**Acceptance Gate:** Target House-Style Score ≥ 9.0 / 10.0 (hard floor: 7.0), zero em dashes.

---

## 5. CGO Dynamic Linking & C2PA Schema Conventions

1. **Darwin RPATH Resolution:**
   When linking against `libcredentio_c.dylib`, ensure `-Wl,-rpath` flags are preserved. macOS System Integrity Protection (SIP) blocks `DYLD_LIBRARY_PATH` inheritance across subshells, requiring embedded rpaths for executable binaries and test runners.
2. **Schema Resilience:**
   `ParseCrJSON` implementations must accommodate both C2PA v1 and v2 structures (`claim` vs `claim.v2`, `certificateInfo` with `issuer.CN`, and categorized `validationResults` with `success`, `failure`, and `informational` lists).

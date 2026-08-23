---
title: Terminal User Interface
description: Overview of the interactive Bubble Tea terminal user interface.
---

`credentialctl` includes a full-screen, keyboard-driven terminal user interface (TUI) built using [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Launching the TUI

To launch the TUI on your current working directory:

```bash
credentialctl
# or explicitly:
credentialctl tui
```

To target a specific folder or media archive:

```bash
credentialctl tui /path/to/media
```

## Global Keybindings

The following controls work across the application:

| Key | Function |
|---|---|
| `q` / `Ctrl+C` | Quit the application immediately |
| `Esc` / `Backspace` | Return to the previous view (e.g. from Inspector back to Folder Review) |
| `Tab` / `f` | Cycle status filters in Folder Review |
| `r` | Rescan directory and refresh validation states |
| `Enter` | Open detailed File Inspector on highlighted asset |

## Two Dedicated Views

1. **Folder Review View:** Scan directories, monitor live progress counters, and filter assets by signature status.
2. **File Inspector View:** Drill down into individual file manifests with 5 specialized inspection tabs.

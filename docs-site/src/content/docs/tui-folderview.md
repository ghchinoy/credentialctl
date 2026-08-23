---
title: Folder Review View
description: Navigating, filtering, and scanning media assets in the Folder Review interface.
---

The Folder Review view provides an at-a-glance dashboard of all supported media files within the scanned directory.

## Interface Elements

### 1. Header Banner
Displays the application title and the absolute directory path currently being monitored.

### 2. Aggregate Status Counters
Shows real-time tally chips across all discovered assets:
- **Total:** Count of discovered media files.
- **Signed:** Count of authentic assets carrying valid C2PA credentials (`[✓ SIGNED]`).
- **Unsigned:** Count of assets with no credential stores (`[⊘ UNSIGNED]`).
- **Invalid:** Count of corrupted or failing assets (`[✕ INVALID]`).

### 3. Media Asset Table
Displays media rows with the following data points:
- **Status:** Semantic badge (`SIGNED`, `UNSIGNED`, `INVALID`).
- **Filename:** File basename or relative path.
- **Size:** Formatted file size in KB or MB.
- **Type:** IANA MIME media type.
- **Signer / Generator:** Name of the signing certificate authority or generator tool.
- **Wall Time:** Validation duration in milliseconds.

## Keybindings & Controls

| Key | Action |
|---|---|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Home` / `g` | Jump to first row |
| `End` / `G` | Jump to last row |
| `Tab` / `f` | Cycle status filters (`All` → `Signed Only` → `Unsigned Only` → `Invalid Only`) |
| `r` | Rescan directory |
| `Enter` | Inspect selected file |
| `?` | Toggle inline help |

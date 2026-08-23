---
title: File Inspector View
description: Forensic inspection tabs and navigation inside the File Inspector.
---

When you press `Enter` on any media row in the Folder Review view, `credentialctl` opens the File Inspector.

## The Five Inspector Tabs

Navigate between tabs using the number keys (`1` - `5`), `Tab`, or the `Left` / `Right` arrow keys.

### 1. Overview (`1`)
Displays high-level metadata:
- Full file path and size.
- Detected MIME media type.
- C2PA specification version (e.g. `C2PA 2.2`).
- Generator application and active manifest URN label.
- Core engine validation duration vs total wall clock time.

### 2. Assertions (`2`)
Lists all assertions attached to the active manifest:
- **Actions:** Editing provenance (e.g. `c2pa.created`, `c2pa.edited`, SynthID watermark applications).
- **Data Hashes:** Cryptographic hash exclusions and pad byte markers.
- **Ingredients:** Input source ingredients used during composite creation.

### 3. Signature & Trust (`3`)
Details cryptographic integrity:
- Signing certificate authority (Issuer Common Name).
- Signing algorithm (e.g. `ES256`, `Ed25519`).
- Timestamp Authority (TSA) signature time.
- Certificate serial numbers and validity periods.

### 4. Validation (`4`)
Enumerates all individual validation rules evaluated by Google Credentio:
- Categorized severity tags: `[INFO]`, `[WARN]`, `[ERROR]`.
- Status codes (e.g. `claimSignature.validated`, `assertion.dataHash.match`).
- Explanations and standard C2PA specification URLs.

### 5. Raw crJSON (`5`)
Provides a scrollable viewport rendering the complete JSON manifest payload formatted with syntax indentation.

## Inspector Controls

| Key | Action |
|---|---|
| `1` - `5` | Jump directly to specific tab |
| `Tab` / `→` / `l` | Switch to next tab |
| `Shift+Tab` / `←` / `h` | Switch to previous tab |
| `↑` / `↓` / `k` / `j` | Scroll viewport line by line |
| `PgUp` / `PgDn` | Scroll viewport page by page |
| `Esc` / `Backspace` | Return to Folder Review table |

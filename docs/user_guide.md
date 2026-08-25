# CredentialCTL User Guide

This guide provides end-to-end instructions for validating and inspecting Coalition for Content Provenance and Authenticity (C2PA) media credentials using `credentialctl`.

---

## 1. Overview and Architecture

`credentialctl` validates cryptographic manifests embedded in digital media. It operates in two primary modes:
1. **Agent-Aware Command-Line Interface:** Produces human-readable semantic summaries or machine-parsable JSON streams for verification pipelines.
2. **Interactive Terminal Interface:** Offers full-screen keyboard navigation to scan folders, filter assets by provenance status, and inspect assertion graphs.

The application links directly against Google Credentio native C-ABI bindings for validation performance.

---

## 2. Command Reference

### `validate`
Validates credentials for a single media asset file.

```bash
# Human-readable output
credentialctl validate photo.jpg

# JSON output
credentialctl validate video.mp4 --json

# Explicit MIME type specification
credentialctl validate asset.bin --media-type image/webp
```

#### Exit Codes
Automation scripts and agent workers can inspect process exit codes to determine asset authenticity:
- `0`: Valid C2PA credentials present (`SIGNED`).
- `1`: No C2PA credentials found in asset (`UNSIGNED`).
- `2`: Corrupted, invalid, or failing credential manifest (`INVALID`).
- `3`: Input error or file not found.

---

### `folder`
Scans a target directory and validates all contained media assets in batch.

```bash
# Non-recursive scan of current folder
credentialctl folder ./media

# Recursive traversal through subdirectories
credentialctl folder ./archive --recursive

# Batch JSON stream output
credentialctl folder ./incoming --json
```

---

### `inspect`
Performs deep structural analysis of embedded manifests, assertions, cryptographic signers, and validation rules.

```bash
# Formatted terminal inspection card
credentialctl inspect artwork.png

# Raw crJSON manifest dump
credentialctl inspect artwork.png --raw

# Structured Go ProvenanceReport model JSON
credentialctl inspect artwork.png --json

# Launch directly into the interactive inspector view
credentialctl inspect artwork.png --tui
```

---

### `tui`
Launches the full-screen interactive Bubble Tea terminal user interface.

```bash
# Open TUI for current working directory
credentialctl tui

# Open TUI on a specific folder
credentialctl tui /path/to/assets

# Open TUI recursively
credentialctl tui /path/to/assets -r
```

---

## 3. Interactive Terminal Interface Guide

The interactive interface organizes operations into two dedicated views.

### Folder Review View
Upon launching `credentialctl tui`, the application discovers media assets and validates them concurrently.
- **Summary Header:** Displays live progress spinners and aggregate counters for Total, Signed, Unsigned, and Invalid files.
- **Asset Table:** Presents file status badges, filenames, sizes, MIME types, signer authorities, and engine execution times.
- **Status Filter:** Press `Tab` or `f` to cycle view filters across `All Files`, `Signed Only`, `Unsigned Only`, and `Invalid Only`.
- **Drill-down:** Highlight any row and press `Enter` to open the File Inspector.

### File Inspector View
The File Inspector presents five navigable tabs for detailed forensic review:
1. **Overview:** Summarizes file attributes, claim generator version, active manifest label, and overall validity status.
2. **Assertions:** Lists attached assertions such as actions (`c2pa.created`, `c2pa.edited`), data hash exclusions, and input ingredients.
3. **Signature:** Displays cryptographic signing algorithms (e.g. ES256), signing timestamps, issuer authorities, and certificate serial numbers.
4. **Validation:** Details individual check statuses, validation codes, severity indicators (`INFO`, `WARN`, `ERROR`), and spec URI references.
5. **Raw crJSON:** Provides a scrollable viewport rendering the complete JSON payload.

Press `Esc` or `Backspace` at any time to return to the Folder Review view.

---

## 4. Agent-Aware Automation Integration

`credentialctl` adheres to Agent-Aware CLI design conventions:
- **Consistent JSON Serialization:** All data commands accept `--json`. Output payloads include engine timing, file metadata, and complete manifest structures.
- **Cobra Grouping:** Commands are grouped logically into `validation`, `inspection`, and `interactive` categories.
- **Actionable Error Messages:** Errors provide prescriptive hints indicating proper path syntax or supported file extensions.

Example pipeline validation command:

```bash
STATUS=$(credentialctl validate input.png --json | jq -r .badge)
if [ "$STATUS" = "signed" ]; then
  echo "Asset verified successfully."
fi
```

---

## 5. Supported Formats and Troubleshooting

### Supported Formats
`credentialctl` detects and parses the following media formats:
- **Images:** JPEG (`.jpg`, `.jpeg`), PNG (`.png`), WebP (`.webp`), AVIF (`.avif`), HEIC (`.heic`, `.heif`), TIFF (`.tif`, `.tiff`)
- **Video:** MP4 (`.mp4`), QuickTime (`.mov`)
- **Audio:** MP3 (`.mp3`), M4A (`.m4a`), WAV (`.wav`)
- **Standalone Manifests:** C2PA manifest stores (`.c2pa`)

### Dynamic Library Loading on macOS
If execution fails with `dyld: Library not loaded: @rpath/libcredentio_c.dylib`, ensure `credentialctl` was compiled with `CGO_ENABLED=1` and run `make fetch-credentio-lib` to stage the prebuilt native library into `third_party/credentio/lib/libcredentio_c.dylib`. Release archives bundle the dynamic library directly beside the executable using `@loader_path`.

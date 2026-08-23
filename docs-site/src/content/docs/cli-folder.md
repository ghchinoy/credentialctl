---
title: folder
description: Command reference for batch directory C2PA validation.
---

The `folder` command scans a directory for supported media files, validates each asset, and presents an aggregate summary table or batch JSON stream.

## Usage

```bash
credentialctl folder <directory_path> [flags]
```

Aliases: `scan`, `dir`.

## Examples

### Scan Directory
```bash
credentialctl folder ./photos
```

Output:
```text
  FOLDER VALIDATION SUMMARY: /path/to/photos  

╭────────────────────────────────────────────────────────────────────────────────────╮
│  Total Files: 14  •  Signed: 4  •  Unsigned: 10  •  Invalid: 0  •  Duration: 0.04s │
╰────────────────────────────────────────────────────────────────────────────────────╯

 STATUS        FILENAME                    SIZE      TYPE          SIGNER / GENERATOR      
 ──────────────────────────────────────────────────────────────────────────────────────────
   ⊘ UNSIGNED   banner.jpg                  154.5 KB  image/jpeg    —                       
   ✓ SIGNED     hero_render.png             1002.7 KB image/png     Google C2PA Media Se... 
   ⊘ UNSIGNED   diagram.webp                44.7 KB   image/webp    —                       
   ✓ SIGNED     diagram_render.png          898.3 KB  image/png     Google C2PA Media Se... 
```

### Recursive Directory Scan
```bash
credentialctl folder /path/to/media --recursive
```

### Batch JSON Output
```bash
credentialctl folder ./photos --json
```

Output:
```json
{
  "directory": "/path/to/photos",
  "total_files": 2,
  "signed_count": 1,
  "unsigned_count": 1,
  "invalid_count": 0,
  "error_count": 0,
  "duration_seconds": 0.009,
  "files": [
    {
      "path": "/path/to/photos/hero_render.png",
      "filename": "hero_render.png",
      "size_bytes": 1026813,
      "media_type": "image/png",
      "validated": true
    }
  ]
}
```

## Flags

| Flag | Description | Default |
|---|---|---|
| `--json` | Output structured JSON for batch results | `false` |
| `-r`, `--recursive` | Recursively scan subdirectories | `false` |
| `--skip-trust-checks` | Skip certificate trust anchor validation | `true` |

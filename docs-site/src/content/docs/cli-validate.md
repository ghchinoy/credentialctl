---
title: validate
description: Command reference for single-file C2PA validation.
---

The `validate` command checks the C2PA manifest, digital signatures, certificate chains, and assertions inside a single target media asset.

## Usage

```bash
credentialctl validate <file_path> [flags]
```

## Examples

### Human-Readable Output
```bash
credentialctl validate photo.jpg
```

Output:
```text
  C2PA VALIDATION REPORT  

  Asset:           photo.jpg (1.20 MB, image/jpeg)
  Path:            /path/to/photo.jpg
  Status:          [✓ SIGNED]
  Generator:       Google C2PA Core Generator Library
  Signer:          Google C2PA Media Services
  Format/Spec:     image/jpeg (C2PA 2.2)
  Assertions:      3 attached
  Validation:      10 reported
  Wall Time:       4.61 ms
```

### JSON Output
```bash
credentialctl validate photo.jpg --json
```

Output payload:
```json
{
  "asset_path": "/path/to/photo.jpg",
  "byte_size": 1258291,
  "media_type": "image/jpeg",
  "engine_id": "credentio",
  "has_credentials": true,
  "badge": "signed",
  "spec_version": "2.2",
  "elapsed_seconds": 0.00461,
  "core_seconds": 0.00386
}
```

### Raw crJSON Output
```bash
credentialctl validate photo.jpg --raw
```

## Flags

| Flag | Description | Default |
|---|---|---|
| `--json` | Output structured JSON for machine and agent parsing | `false` |
| `--raw` | Output raw native engine crJSON payload directly | `false` |
| `--media-type` | Explicitly specify IANA media type (e.g. `image/jpeg`) | auto-detected |
| `--skip-trust-checks` | Skip certificate trust anchor validation for local verification | `true` |

## Process Exit Codes

`credentialctl validate` exits with deterministic status codes to simplify script branching:
- `0`: Valid C2PA credentials verified (`SIGNED`).
- `1`: No C2PA credentials found in asset (`UNSIGNED`).
- `2`: Corrupted or invalid credentials (`INVALID`).
- `3`: Input error (file missing or unreadable).

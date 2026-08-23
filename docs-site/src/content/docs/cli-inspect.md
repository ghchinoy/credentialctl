---
title: inspect
description: Command reference for forensic C2PA manifest inspection.
---

The `inspect` command performs deep structural analysis of embedded C2PA manifests, assertions (actions, hashes, ingredients), cryptographic signatures, and validation status explanations.

## Usage

```bash
credentialctl inspect <file_path> [flags]
```

## Examples

### Human-Readable Inspection
```bash
credentialctl inspect image.png
```

Output:
```text
  C2PA MANIFEST INSPECTION: image.png  

  Path:              /path/to/image.png
  Size:              0.98 MB (1026813 bytes)
  Media Type:        image/png
  Status:            [✓ SIGNED]

MANIFEST METADATA
  Label:             urn:c2pa:cfd29b96-e160-a6ad-cbdb-496b06d580f5
  Claim Generator:   Google C2PA Core Generator Library 899773717

SIGNATURE & TRUST
  Issuer / Signer:   Google C2PA Media Services 1P ICA G3
  Algorithm:         ES256
  Signing Time:      Mon, 20 Apr 2026 22:48:01 +0000
  Cert Serial:       eec7bfb94c35dc6a264f917da23fb79f242b02

ASSERTIONS (3)
  1. c2pa.actions.v2   actions 
     Summary: c2pa.created, c2pa.edited
  2. c2pa.hash.data   hash 
  3. c2pa.ingredient.v3   ingredient 

VALIDATION STATUSES (10)
  1. [INFO] signingCredential.ocsp.skipped
     Spec: self#jumbf=/c2pa/urn:c2pa:cfd29b96-e160-a6ad-cbdb-496b06d580f5/c2pa.signature
  2. [INFO] timeStamp.validated
     ...
```

### Raw crJSON Output
```bash
credentialctl inspect image.png --raw
```

### Launch Interactive TUI Inspector
```bash
credentialctl inspect image.png --tui
```

## Flags

| Flag | Description | Default |
|---|---|---|
| `--json` | Output full Go ProvenanceReport model as JSON | `false` |
| `--raw` | Output unformatted or pretty-printed raw crJSON | `false` |
| `--tui` | Open directly in interactive terminal inspector UI | `false` |
| `--skip-trust-checks` | Skip certificate trust anchor validation | `true` |

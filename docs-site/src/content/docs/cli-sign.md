---
title: sign & keygen
description: Generate C2PA-compliant test PKIs and sign media assets in-process with credentialctl.
---

`credentialctl` provides built-in commands for generating local C2PA-compliant Public Key Infrastructure (`keygen`) and embedding signed C2PA manifests into media assets in-process (`sign`).

---

## 1. Generating a Local Test PKI (`credentialctl keygen`)

The C2PA specification enforces strict X.509 v3 certificate profiles:
- **Claim Signer Certificate:** Requires the `digitalSignature` key usage and the C2PA Claim Signing Extended Key Usage (`1.3.6.1.4.1.62558.2.1`).
- **TSA Certificate:** Requires a critical RFC 3161 `timeStamping` Extended Key Usage (`1.3.6.1.5.5.7.3.8`).

Generate a complete EC P-256 test PKI bundle with a single command (zero external OpenSSL scripts required):

```bash
credentialctl keygen --out-dir ./test_pki --org "My Media Studio"
```

This writes six PEM files into `./test_pki`:
- `root_ca_key.pem` & `root_ca_cert.pem`
- `claim_signer_key.pem` & `claim_signer_cert.pem`
- `tsa_key.pem` & `tsa_cert.pem`

Pass `--json` to output the generated file paths as machine-readable JSON.

---

## 2. Signing Media Assets (`credentialctl sign`)

Embed a C2PA manifest store into an image, video, audio clip, PDF, or Office document, sign it in-process, and immediately verify the signed output.

### One-Line Ephemeral Signing
If `--key` and `--cert` are omitted, `credentialctl sign` automatically generates an ephemeral C2PA test PKI for quick local testing:

```bash
credentialctl sign input.png -o signed.png
```

### Signing with Explicit Keys, AI Provenance, and Training Opt-Out

```bash
credentialctl sign artwork.png -o artwork_signed.png \
  --key ./test_pki/claim_signer_key.pem \
  --cert ./test_pki/claim_signer_cert.pem \
  --tsa-key ./test_pki/tsa_key.pem \
  --tsa-cert ./test_pki/tsa_cert.pem \
  --source-type ai \
  --ai-opt-out \
  --generator "StudioPipeline"
```

### Flags Reference

| Flag | Description |
| :--- | :--- |
| `-o`, `--output <path>` | Output path for the signed media asset (required; must differ from input path). |
| `--key <pem_path>` | Path to claim signer EC P-256 PKCS#8 private key PEM. |
| `--cert <pem_path>` | Path to claim signer X.509 certificate chain PEM (leaf first). |
| `--tsa-key <pem_path>` | Optional path to TSA EC P-256 private key PEM. |
| `--tsa-cert <pem_path>` | Optional path to TSA X.509 certificate PEM. |
| `--source-type <val>` | IPTC `digitalSourceType` URI or shorthand: `capture`, `ai`, `composite-ai`, `algorithmic`, `creation`. |
| `--ai-opt-out` | Attach a `c2pa.training-mining` assertion setting `c2pa.ai_generative_training` and `c2pa.data_mining` to `notAllowed`. |
| `--generator <name>` | Claim generator software name (default: `"credentialctl"`). |
| `--media-type <mime>` | Explicit IANA media type override. |
| `--json` | Emit structured JSON output with signing latency and post-sign verification report. |

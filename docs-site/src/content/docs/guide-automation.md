---
title: Agent-Aware Automation
description: Best practices for integrating credentialctl into automated agent workflows and CI/CD pipelines.
---

`credentialctl` is designed from the ground up to support autonomous agents, verification sidecars, and automated CI/CD pipelines.

## Machine-Readable JSON Output

Every CLI command that outputs data supports the `--json` flag.

### Single Asset Verification

```bash
credentialctl validate input.png --json
```

Parsing with `jq` in bash scripts:

```bash
# Extract validity badge
BADGE=$(credentialctl validate input.png --json | jq -r .badge)

if [ "$BADGE" != "signed" ]; then
  echo "Validation failed: Asset is $BADGE"
  exit 1
fi
```

### Batch Directory Verification

```bash
credentialctl folder ./incoming_media --json
```

Extracting files with errors or missing credentials:

```bash
# List all unsigned assets from batch results
credentialctl folder ./incoming_media --json | jq -r '.files[] | select(.report.has_credentials == false) | .path'
```

## Exit Code Automation

`credentialctl validate` maps validity directly to process exit codes:

```bash
credentialctl validate upload.jpg > /dev/null 2>&1
STATUS=$?

case $STATUS in
  0)
    echo "Authentic signed media."
    ;;
  1)
    echo "Unsigned media without credentials."
    ;;
  2)
    echo "Invalid or corrupted credentials."
    ;;
  *)
    echo "Error processing asset."
    ;;
esac
```

## Proactive Error Hints

When automation pipelines supply invalid file paths or directories, error messages return actionable hints:
```text
Error: file not found at '/path/to/missing.jpg'.
Hint: Check the filename spelling or provide an absolute path.
```

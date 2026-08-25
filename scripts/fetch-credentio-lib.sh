#!/usr/bin/env bash
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# fetch-credentio-lib.sh: downloads pre-compiled native libcredentio_c shared library
# from GitHub Releases for credentialctl.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

VERSION="${1:-0.1.2}"
VERSION="${VERSION#v}" # strip leading v if provided

REPO="ghchinoy/credentio-contributions"
BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"

# Detect OS and Arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${ARCH}" in
  x86_64|amd64)
    ARCH_NAME="amd64"
    ;;
  arm64|aarch64)
    ARCH_NAME="arm64"
    ;;
  *)
    echo "ERROR: Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

case "${OS}" in
  darwin)
    EXT="dylib"
    PLATFORM="darwin-${ARCH_NAME}"
    ;;
  linux)
    EXT="so"
    PLATFORM="linux-${ARCH_NAME}"
    ;;
  *)
    echo "ERROR: Unsupported OS: ${OS}" >&2
    exit 1
    ;;
esac

TARGET_LIB="libcredentio_c.${EXT}"
DOWNLOAD_URL="${BASE_URL}/libcredentio_c-${PLATFORM}.${EXT}"
OUTPUT_DIR="${REPO_DIR}/third_party/credentio/lib"
OUTPUT_FILE="${OUTPUT_DIR}/${TARGET_LIB}"

echo "=== Fetching prebuilt Credentio library for ${PLATFORM} (v${VERSION}) ==="
echo "Downloading: ${DOWNLOAD_URL}"

mkdir -p "${OUTPUT_DIR}"

TMP_FILE="$(mktemp /tmp/credentio-lib.XXXXXX)"
trap 'rm -f "${TMP_FILE}"' EXIT

if curl -fL --progress-bar -o "${TMP_FILE}" "${DOWNLOAD_URL}"; then
  echo "==> Download completed successfully."
else
  # Fallback to direct library filename
  FALLBACK_URL="${BASE_URL}/${TARGET_LIB}"
  echo "==> Trying release asset fallback: ${FALLBACK_URL}"
  if ! curl -fL --progress-bar -o "${TMP_FILE}" "${FALLBACK_URL}"; then
    echo "ERROR: Failed to download libcredentio_c from release v${VERSION}." >&2
    echo "Check release assets at https://github.com/${REPO}/releases/tag/v${VERSION}" >&2
    exit 1
  fi
fi

cp -f "${TMP_FILE}" "${OUTPUT_FILE}"
chmod 755 "${OUTPUT_FILE}"

echo "======================================================="
echo "SUCCESS: Staged ${TARGET_LIB} into ${OUTPUT_FILE}"
echo "======================================================="

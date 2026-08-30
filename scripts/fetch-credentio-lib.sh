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

VERSION="${1:-0.1.5}"
VERSION="${VERSION#v}" # strip leading v if provided

REPO="ghchinoy/credentio-contributions"
BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"
OUTPUT_DIR="${REPO_DIR}/third_party/credentio/lib"

mkdir -p "${OUTPUT_DIR}"

echo "=== Fetching prebuilt Credentio native libraries (v${VERSION}) ==="

fetch_asset() {
  local asset_name="$1"
  local target_name="$2"
  local download_url="${BASE_URL}/${asset_name}"
  local output_file="${OUTPUT_DIR}/${target_name}"
  local tmp_file
  tmp_file="$(mktemp /tmp/credentio-lib.XXXXXX)"

  echo "Downloading: ${download_url} -> ${target_name}"
  if curl -fL --progress-bar -o "${tmp_file}" "${download_url}"; then
    cp -f "${tmp_file}" "${output_file}"
    chmod 755 "${output_file}"
    rm -f "${tmp_file}"
    echo "==> Staged ${target_name} successfully."
  else
    rm -f "${tmp_file}"
    echo "WARNING: Failed to download ${asset_name} from v${VERSION}." >&2
    return 1
  fi
}

# Fetch Darwin arm64 asset
fetch_asset "libcredentio_c-darwin-arm64.dylib" "libcredentio_c.dylib" || true
if [ -f "${OUTPUT_DIR}/libcredentio_c.dylib" ]; then
  cp -f "${OUTPUT_DIR}/libcredentio_c.dylib" "${OUTPUT_DIR}/libcredentio_c-darwin-arm64.dylib"
fi

# Fetch Linux amd64 asset
fetch_asset "libcredentio_c-linux-amd64.so" "libcredentio_c.so" || true
if [ -f "${OUTPUT_DIR}/libcredentio_c.so" ]; then
  cp -f "${OUTPUT_DIR}/libcredentio_c.so" "${OUTPUT_DIR}/libcredentio_c-linux-amd64.so"
fi

# Verify host platform library is present
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  darwin)
    if [ ! -f "${OUTPUT_DIR}/libcredentio_c.dylib" ]; then
      echo "ERROR: libcredentio_c.dylib is required on Darwin but was not staged." >&2
      exit 1
    fi
    ;;
  linux)
    if [ ! -f "${OUTPUT_DIR}/libcredentio_c.so" ]; then
      echo "ERROR: libcredentio_c.so is required on Linux but was not staged." >&2
      exit 1
    fi
    ;;
esac

echo "======================================================="
echo "SUCCESS: Staged Credentio native libraries in ${OUTPUT_DIR}"
ls -lh "${OUTPUT_DIR}"
echo "======================================================="

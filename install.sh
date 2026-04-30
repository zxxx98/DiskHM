#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "This installer currently supports Linux only." >&2
  exit 1
fi

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run this installer as root, for example:" >&2
  echo "  curl -fsSL https://raw.githubusercontent.com/zxxx98/DiskHM/main/install.sh | sudo bash" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required." >&2
  exit 1
fi

if ! command -v tar >/dev/null 2>&1; then
  echo "tar is required." >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required to build DiskHM from source." >&2
  exit 1
fi

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

ARCHIVE_URL="${DISKHM_ARCHIVE_URL:-https://github.com/zxxx98/DiskHM/archive/refs/heads/main.tar.gz}"
ARCHIVE_PATH="${WORK_DIR}/diskhm.tar.gz"

curl -fsSL "${ARCHIVE_URL}" -o "${ARCHIVE_PATH}"
tar -xzf "${ARCHIVE_PATH}" -C "${WORK_DIR}"

SOURCE_DIR="$(find "${WORK_DIR}" -mindepth 1 -maxdepth 1 -type d -name 'DiskHM-*' | head -n 1)"
if [[ -z "${SOURCE_DIR}" ]]; then
  echo "Failed to locate extracted source tree." >&2
  exit 1
fi

export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"

cd "${SOURCE_DIR}"
go build -o diskhmd ./cmd/diskhmd
./scripts/install-local.sh

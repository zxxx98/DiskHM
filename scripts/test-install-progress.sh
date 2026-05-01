#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALLER_PATH="${REPO_ROOT}/install.sh"

assert_script_contains() {
  local needle="$1"
  if ! grep -Fq "${needle}" "${INSTALLER_PATH}"; then
    echo "expected installer to contain progress message: ${needle}" >&2
    exit 1
  fi
}

assert_script_contains "Downloading DiskHM source archive..."
assert_script_contains "Extracting DiskHM source archive..."
assert_script_contains "Building diskhmd with Go"
assert_script_contains "Installing DiskHM locally..."

echo "installer progress output verified"

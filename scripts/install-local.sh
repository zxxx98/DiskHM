#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

BINARY_SOURCE="${REPO_ROOT}/diskhmd"
BINARY_DEST="/usr/local/bin/diskhm"
SERVICE_SOURCE="${REPO_ROOT}/packaging/systemd/diskhm.service"
SERVICE_DEST="/etc/systemd/system/diskhm.service"
CONFIG_DIR="/etc/diskhm"
CONFIG_PATH="${CONFIG_DIR}/config.yaml"
DATA_DIR="/var/lib/diskhm"

if [[ ! -f "${BINARY_SOURCE}" ]]; then
  echo "missing binary: ${BINARY_SOURCE}" >&2
  echo "build it first with: C:\\Program Files\\Go\\bin\\go.exe build -o diskhmd ./cmd/diskhmd" >&2
  exit 1
fi

install -d -m 0755 "${CONFIG_DIR}"
install -d -m 0755 "${DATA_DIR}"
install -m 0755 "${BINARY_SOURCE}" "${BINARY_DEST}"
install -m 0644 "${SERVICE_SOURCE}" "${SERVICE_DEST}"

if [[ ! -f "${CONFIG_PATH}" ]]; then
  cat >"${CONFIG_PATH}" <<'EOF'
server:
  listen_addr: 0.0.0.0:9789
security:
  token_hash: bootstrap-token-hash-change-me
sleep:
  quiet_grace_seconds: 10
EOF
fi

systemctl daemon-reload
systemctl enable --now diskhm.service

echo "diskhm installed."
echo "Service commands:"
echo "  sudo diskhm start"
echo "  sudo diskhm stop"
echo "  sudo diskhm enable"
echo "  sudo diskhm disable"
echo "  sudo diskhm uninstall"
echo "Web UI:"
echo "  http://127.0.0.1:9789"
if command -v hostname >/dev/null 2>&1; then
  LAN_IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
  if [[ -n "${LAN_IP:-}" ]]; then
    echo "  http://${LAN_IP}:9789"
  fi
fi

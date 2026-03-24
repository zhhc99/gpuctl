#!/usr/bin/env bash
{
set -euo pipefail

err() { echo "Error: $*" >&2; exit 1; }

BIN_PATH="/usr/local/bin/gpuctl"
UNIT_PATH="/etc/systemd/system/gpuctl.service"
UNIT_NAME="gpuctl.service"
CONF_DIR="/etc/gpuctl"

[[ $EUID -eq 0 ]] || err "please run with sudo"

echo "Stopping and disabling service..."
systemctl disable --now "$UNIT_NAME" 2>/dev/null || true

echo "Removing service unit..."
rm -f "$UNIT_PATH"
systemctl daemon-reload
systemctl reset-failed 2>/dev/null || true

echo "Removing binary..."
rm -f "$BIN_PATH"

if [[ -d "$CONF_DIR" ]]; then
    read -r -p "Remove config directory $CONF_DIR? [y/N] (default N): " ans
    ans="${ans:-n}"
    if [[ "${ans,,}" == "y" || "${ans,,}" == "yes" ]]; then
        rm -rf "$CONF_DIR"
        echo "Removed $CONF_DIR."
    else
        echo "Keeping $CONF_DIR."
    fi
fi

echo "Done. gpuctl has been uninstalled."
}

#!/usr/bin/env bash
{
set -euo pipefail

err() { echo "Error: $*" >&2; exit 1; }

REPO="zhhc99/gpuctl"
BIN_PATH="/usr/local/bin/gpuctl"
UNIT_PATH="/etc/systemd/system/gpuctl.service"

[[ $EUID -eq 0 ]] || err "please run with sudo"
[[ "$(uname -s)" == "Linux" ]] || err "unsupported OS"

case "$(uname -m)" in
  x86_64)        ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64"  ;;
  *)             err "unsupported architecture: $(uname -m)" ;;
esac

URL="https://github.com/$REPO/releases/latest/download/gpuctl_Linux_${ARCH}.tar.gz"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading gpuctl for Linux/${ARCH}..."
curl -sSL "$URL" | tar -xz -C "$TMP"

echo "Installing binary to $BIN_PATH..."
install -m 755 "$TMP/gpuctl" "$BIN_PATH"

echo "Writing service unit to $UNIT_PATH..."
cat > "$UNIT_PATH" << UNIT
[Unit]
Description=gpuctl GPU controller
After=multi-user.target

[Service]
Type=simple
ExecStart=$BIN_PATH daemon
ExecReload=$BIN_PATH load
ExecStop=$BIN_PATH tune reset fan --all
Restart=on-failure

[Install]
WantedBy=multi-user.target
UNIT

echo "Enabling service..."
systemctl daemon-reload
systemctl enable --now gpuctl.service

echo "Done. Service is active."
}

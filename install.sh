#!/usr/bin/env bash
{
set -euo pipefail

err() { echo "Error: $*" >&2; exit 1; }

REPO="zhhc99/gpuctl"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

[[ $EUID -eq 0 ]] || err "please run with sudo"

[[ "$OS" == "linux" ]] || err "unsupported OS: $OS"
case "$ARCH" in
  x86_64)        ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)             err "unsupported architecture: $ARCH" ;;
esac

URL="https://github.com/$REPO/releases/latest/download/gpuctl_Linux_${ARCH}.tar.gz"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading gpuctl for Linux_${ARCH}..."
curl -sSL "$URL" | tar -xz -C "$TMP"

echo "Installing..."
sudo "$TMP/gpuctl" install
}

#!/usr/bin/env bash
#
# 安装 gpuctl 到 /usr/local/bin

{
set -euo pipefail

GPUCTL_TMP_DIR=""

err() {
  echo "❗ Error: $*" >&2
  exit 1
}

cleanup() {
  if [[ -n "${GPUCTL_TMP_DIR:-}" ]]; then
    rm -rf "${GPUCTL_TMP_DIR}"
  fi
}
trap cleanup EXIT

main() {
  if [[ $EUID -ne 0 ]]; then
    echo "🛡️  Need sudo powers to install..."
    exec sudo bash "$0" "$@"
  fi

  # os & arch check
  local os arch
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)

  [[ "${os}" == "linux" ]] || err "unsupported os: ${os}"

  case "${arch}" in
    x86_64) arch="x86_64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) err "unsupported architecture: ${arch}" ;;
  esac

  # find url for latest release
  local platform="Linux_${arch}"
  local repo="zhhc99/gpuctl"
  local url="https://github.com/${repo}/releases/latest/download/gpuctl_${platform}.tar.gz"

  # download and install
  GPUCTL_TMP_DIR=$(mktemp -d)
  echo "🚀 Downloading gpuctl for ${platform}..."

  trap 'rm -rf "${GPUCTL_TMP_DIR}"' EXIT
  if ! curl -sSL "${url}" | tar -xz -C "${GPUCTL_TMP_DIR}"; then
    err "failed to download. something must be wrong here... 🤔"
  fi

  echo "⚙️ Installing to /usr/local/bin..."
  sudo install -m 755 "${GPUCTL_TMP_DIR}/gpuctl" /usr/local/bin/gpuctl

  echo "🎉 Done. Try run 'gpuctl'!"
}

main "$@"
}
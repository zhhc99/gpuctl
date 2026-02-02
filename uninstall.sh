#!/usr/bin/env bash
#
# 卸载 gpuctl 及服务

set -euo pipefail

main() {
  if [[ $EUID -ne 0 ]]; then
    echo "🛡️  Need sudo powers to clean up..."
    exec sudo bash "$0" "$@"
  fi

  # disable all gpuctl@*.service
  local services
  services=$(systemctl list-units --type=service --all --no-legend "gpuctl@*" | awk '{print $1}') || true

  echo "🧹 Stopping and disabling active services..."
  if [[ -n "${services}" ]]; then
    for svc in ${services}; do
      echo "   - Terminating ${svc}"
      #  systemctl stop "${svc}" >/dev/null 2>&1 || true
      systemctl disable "${svc}" >/dev/null 2>&1 || true
    done
  else
    echo "No active gpuctl services found. Skipping..."
  fi

  # remove template gpuctl@.service
  local unit_file="/etc/systemd/system/gpuctl@.service"
  if [[ -f "${unit_file}" ]]; then
    echo "Removing systemd unit file..."
    rm -f "${unit_file}"
    systemctl daemon-reload
    systemctl reset-failed
  fi

  # remove binary
  local bin_path="/usr/local/bin/gpuctl"
  echo "🗑️ Removing binary from ${bin_path}..."
  if [[ -f "${bin_path}" ]]; then
    rm -f "${bin_path}"
  else
    echo "Binary not found. It might have already left the building... 🏃"
  fi

  echo "👋 Done. See you next time!"
}

main "$@"
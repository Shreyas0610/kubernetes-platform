#!/usr/bin/env bash
set -euo pipefail

CONFIG_PATH="${1:-./configs/kubeadm-join-worker.yaml}"

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run as root: sudo $0 ${CONFIG_PATH}" >&2
  exit 1
fi

if [[ ! -f "${CONFIG_PATH}" ]]; then
  echo "Config file not found: ${CONFIG_PATH}" >&2
  exit 1
fi

kubeadm join --config "${CONFIG_PATH}"

echo "Worker node joined."


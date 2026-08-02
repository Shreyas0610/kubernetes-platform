#!/usr/bin/env bash
set -euo pipefail

CONFIG_PATH="${1:-./configs/kubeadm-control-plane.yaml}"

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run as root: sudo $0 ${CONFIG_PATH}" >&2
  exit 1
fi

if [[ ! -f "${CONFIG_PATH}" ]]; then
  echo "Config file not found: ${CONFIG_PATH}" >&2
  exit 1
fi

kubeadm init --config "${CONFIG_PATH}" --upload-certs

echo
echo "Control plane initialized."
echo "Copy /etc/kubernetes/admin.conf to your admin workstation or run:"
echo "  mkdir -p \$HOME/.kube"
echo "  sudo cp /etc/kubernetes/admin.conf \$HOME/.kube/config"
echo "  sudo chown \$(id -u):\$(id -g) \$HOME/.kube/config"
echo
echo "Install a CNI before joining workers or scheduling workloads."


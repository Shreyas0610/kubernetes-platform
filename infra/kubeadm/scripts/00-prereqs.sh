#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run as root: sudo $0" >&2
  exit 1
fi

modprobe overlay
modprobe br_netfilter

cat >/etc/modules-load.d/kubernetes-platform.conf <<'EOF'
overlay
br_netfilter
EOF

cat >/etc/sysctl.d/99-kubernetes-platform.conf <<'EOF'
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward = 1
EOF

sysctl --system

swapoff -a
if grep -qE '^[^#].*\sswap\s' /etc/fstab; then
  cp /etc/fstab /etc/fstab.kubernetes-platform.bak
  sed -i.bak '/\sswap\s/s/^/#/' /etc/fstab
fi

systemctl enable --now systemd-timesyncd || true

echo "Node prerequisites complete."


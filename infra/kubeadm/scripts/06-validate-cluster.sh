#!/usr/bin/env bash
set -euo pipefail

kubectl version
kubectl get nodes -o wide
kubectl -n kube-system get pods -o wide
kubectl get --raw='/readyz?verbose'
kubectl get componentstatuses 2>/dev/null || true

echo
echo "Validation checks completed. If nodes are NotReady, install or inspect the CNI."


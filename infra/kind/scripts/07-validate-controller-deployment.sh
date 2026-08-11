#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SAMPLE="${REPO_ROOT}/platform/app-controller/config/samples/platform_v1alpha1_app.yaml"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "Missing required command: kubectl" >&2
  exit 1
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
  echo "kubectl cannot reach a Kubernetes cluster." >&2
  echo "Create the kind cluster first: ./infra/kind/scripts/00-create-cluster.sh" >&2
  exit 1
fi

echo "Validating controller Deployment."
kubectl get deployment app-controller-controller-manager \
  --namespace app-controller-system
kubectl rollout status deployment/app-controller-controller-manager \
  --namespace app-controller-system \
  --timeout=120s

echo "Applying sample App."
kubectl apply -f "${SAMPLE}"

echo "Validating generated workload."
kubectl get app demo-api -o wide
kubectl get deployment demo-api
kubectl rollout status deployment/demo-api --timeout=120s
kubectl get service demo-api
kubectl get ingress demo-api

echo "Inspect controller logs when debugging:"
echo "  kubectl logs -n app-controller-system deployment/app-controller-controller-manager"

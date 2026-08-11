#!/usr/bin/env bash
set -euo pipefail

IMG="${IMG:-app-controller:kind}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
KUSTOMIZATION="${REPO_ROOT}/platform/app-controller/config/manager/kustomization.yaml"
BACKUP="$(mktemp)"

cleanup() {
  cp "${BACKUP}" "${KUSTOMIZATION}"
  rm -f "${BACKUP}"
}

if ! command -v kubectl >/dev/null 2>&1; then
  echo "Missing required command: kubectl" >&2
  exit 1
fi

if ! command -v kind >/dev/null 2>&1; then
  echo "Missing required command: kind" >&2
  echo "Install it with: brew install kind" >&2
  exit 1
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
  echo "kubectl cannot reach a Kubernetes cluster." >&2
  echo "Create the kind cluster first: ./infra/kind/scripts/00-create-cluster.sh" >&2
  exit 1
fi

cp "${KUSTOMIZATION}" "${BACKUP}"
trap cleanup EXIT

make -C "${REPO_ROOT}/platform/app-controller" deploy IMG="${IMG}"
kubectl rollout status deployment/app-controller-controller-manager \
  --namespace app-controller-system \
  --timeout=120s

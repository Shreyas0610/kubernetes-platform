#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-kubernetes-platform}"
CERT_MANAGER_VERSION="${CERT_MANAGER_VERSION:-v1.21.1}"
MANIFEST_URL="https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ISSUER_MANIFEST="${REPO_ROOT}/infra/kind/cert-manager/local-selfsigned-clusterissuer.yaml"

require_command() {
  local name="$1"
  if ! command -v "${name}" >/dev/null 2>&1; then
    echo "Missing required command: ${name}" >&2
    exit 1
  fi
}

require_command docker
require_command kind
require_command kubectl

if ! docker info >/dev/null 2>&1; then
  echo "Docker is not running or is not reachable." >&2
  exit 1
fi

if ! kind get clusters | grep -qx "${CLUSTER_NAME}"; then
  echo "kind cluster '${CLUSTER_NAME}' does not exist." >&2
  echo "Create it first: ./infra/kind/scripts/00-create-cluster.sh" >&2
  exit 1
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
  echo "kubectl cannot reach a Kubernetes cluster." >&2
  echo "Create the kind cluster first: ./infra/kind/scripts/00-create-cluster.sh" >&2
  exit 1
fi

echo "Installing cert-manager from ${MANIFEST_URL}"
kubectl apply -f "${MANIFEST_URL}"

kubectl rollout status deployment/cert-manager \
  --namespace cert-manager \
  --timeout=180s
kubectl rollout status deployment/cert-manager-cainjector \
  --namespace cert-manager \
  --timeout=180s
kubectl rollout status deployment/cert-manager-webhook \
  --namespace cert-manager \
  --timeout=180s

kubectl apply -f "${ISSUER_MANIFEST}"
kubectl get clusterissuer platform-local-selfsigned

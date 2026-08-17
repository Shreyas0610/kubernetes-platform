#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-kubernetes-platform}"
NODE_NAME="${CLUSTER_NAME}-control-plane"
HOST="${HOST:-demo.local}"
URL="${URL:-http://localhost:8080/}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SAMPLE="${REPO_ROOT}/platform/app-controller/config/samples/platform_v1alpha1_app.yaml"

require_command() {
  local name="$1"
  if ! command -v "${name}" >/dev/null 2>&1; then
    echo "Missing required command: ${name}" >&2
    exit 1
  fi
}

require_command curl
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

if ! docker port "${NODE_NAME}" 80/tcp | grep -qE '(^|:)8080$'; then
  echo "kind node '${NODE_NAME}' is not exposing container port 80 on host port 8080." >&2
  echo "Recreate the cluster with ./infra/kind/scripts/00-create-cluster.sh after deleting the old kind cluster." >&2
  exit 1
fi

kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=180s

kubectl rollout status deployment/app-controller-controller-manager \
  --namespace app-controller-system \
  --timeout=120s

kubectl apply -f "${SAMPLE}"
kubectl patch app demo-api --type merge -p '{"spec":{"replicas":1}}'
kubectl apply -f "${SAMPLE}"

kubectl rollout status deployment/demo-api --timeout=120s
kubectl get service demo-api
kubectl get endpoints demo-api

actual_host="$(kubectl get ingress demo-api -o jsonpath='{.spec.rules[0].host}')"
if [ "${actual_host}" != "${HOST}" ]; then
  echo "Ingress/demo-api host mismatch: expected '${HOST}', got '${actual_host}'." >&2
  exit 1
fi

actual_class="$(kubectl get ingress demo-api -o jsonpath='{.spec.ingressClassName}')"
if [ "${actual_class}" != "nginx" ]; then
  echo "Ingress/demo-api class mismatch: expected 'nginx', got '${actual_class}'." >&2
  exit 1
fi

for attempt in $(seq 1 30); do
  body="$(curl --silent --show-error --max-time 5 --header "Host: ${HOST}" "${URL}" || true)"
  if printf '%s' "${body}" | grep -qi "Welcome to nginx"; then
    echo "HTTP routing validated through ingress-nginx."
    echo "Request: curl -H 'Host: ${HOST}' ${URL}"
    exit 0
  fi
  echo "Waiting for ingress route to serve ${HOST} (${attempt}/30)."
  sleep 2
done

echo "Ingress route did not return the expected nginx response." >&2
echo "Debug commands:" >&2
echo "  kubectl describe ingress demo-api" >&2
echo "  kubectl get endpoints demo-api -o wide" >&2
echo "  kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller" >&2
exit 1

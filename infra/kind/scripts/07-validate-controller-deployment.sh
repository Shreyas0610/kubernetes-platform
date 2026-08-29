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

for attempt in $(seq 1 30); do
  if kubectl get configmap demo-api-env >/dev/null 2>&1; then
    break
  fi
  if [ "${attempt}" = "30" ]; then
    echo "ConfigMap/demo-api-env was not created." >&2
    exit 1
  fi
  echo "Waiting for ConfigMap/demo-api-env (${attempt}/30)."
  sleep 2
done
log_level="$(kubectl get configmap demo-api-env -o jsonpath='{.data.LOG_LEVEL}')"
if [ "${log_level}" != "debug" ]; then
  echo "ConfigMap/demo-api-env LOG_LEVEL mismatch: expected 'debug', got '${log_level}'." >&2
  exit 1
fi

generated_env_ref="$(kubectl get deployment demo-api -o jsonpath='{.spec.template.spec.containers[0].envFrom[0].configMapRef.name}')"
if [ "${generated_env_ref}" != "demo-api-env" ]; then
  echo "Deployment/demo-api envFrom mismatch: expected generated ConfigMap 'demo-api-env', got '${generated_env_ref}'." >&2
  exit 1
fi

for attempt in $(seq 1 30); do
  phase="$(kubectl get app demo-api -o jsonpath='{.status.phase}')"
  ready_status="$(kubectl get app demo-api -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"
  if [ "${phase}" = "Ready" ] && [ "${ready_status}" = "True" ]; then
    echo "App/demo-api status validated: phase=Ready, Ready=True."
    break
  fi
  if [ "${attempt}" = "30" ]; then
    echo "App/demo-api did not report Ready after the Deployment rolled out." >&2
    kubectl get app demo-api -o yaml >&2
    exit 1
  fi
  echo "Waiting for App/demo-api status to report Ready (${attempt}/30)."
  sleep 2
done

kubectl get service demo-api
kubectl get ingress demo-api

echo "Inspect controller logs when debugging:"
echo "  kubectl logs -n app-controller-system deployment/app-controller-controller-manager"

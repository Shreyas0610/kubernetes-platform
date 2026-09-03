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

wait_for_generated_app_config() {
  for attempt in $(seq 1 30); do
    local env_ref
    local readiness_path
    local liveness_path
    local cpu_request
    local memory_limit

    env_ref="$(kubectl get deployment demo-api -o jsonpath='{.spec.template.spec.containers[0].envFrom[0].configMapRef.name}')"
    readiness_path="$(kubectl get deployment demo-api -o jsonpath='{.spec.template.spec.containers[0].readinessProbe.httpGet.path}')"
    liveness_path="$(kubectl get deployment demo-api -o jsonpath='{.spec.template.spec.containers[0].livenessProbe.httpGet.path}')"
    cpu_request="$(kubectl get deployment demo-api -o jsonpath='{.spec.template.spec.containers[0].resources.requests.cpu}')"
    memory_limit="$(kubectl get deployment demo-api -o jsonpath='{.spec.template.spec.containers[0].resources.limits.memory}')"

    if [ "${env_ref}" = "demo-api-env" ] &&
      [ "${readiness_path}" = "/" ] &&
      [ "${liveness_path}" = "/" ] &&
      [ "${cpu_request}" = "50m" ] &&
      [ "${memory_limit}" = "256Mi" ]; then
      return 0
    fi

    echo "Waiting for Deployment/demo-api generated config (${attempt}/30)."
    sleep 2
  done

  echo "Deployment/demo-api did not converge to the expected generated config." >&2
  kubectl get deployment demo-api -o yaml >&2
  exit 1
}

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

wait_for_generated_app_config

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

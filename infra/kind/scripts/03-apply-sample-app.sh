#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SAMPLE="${REPO_ROOT}/platform/app-controller/config/samples/platform_v1alpha1_app.yaml"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "Missing required command: kubectl" >&2
  exit 1
fi

kubectl apply -f "${SAMPLE}"
kubectl get app demo-api

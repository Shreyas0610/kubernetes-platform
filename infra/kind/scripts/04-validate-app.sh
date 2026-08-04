#!/usr/bin/env bash
set -euo pipefail

if ! command -v kubectl >/dev/null 2>&1; then
  echo "Missing required command: kubectl" >&2
  exit 1
fi

echo "Validating App resource."
kubectl get app demo-api -o wide

echo "Validating generated Deployment."
kubectl get deployment demo-api
kubectl rollout status deployment/demo-api --timeout=120s

echo "Validating generated Service."
kubectl get service demo-api

echo "Validating generated Ingress."
kubectl get ingress demo-api

echo "Inspect App status when debugging:"
echo "  kubectl describe app demo-api"

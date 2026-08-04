#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "Missing required command: kubectl" >&2
  exit 1
fi

make -C "${REPO_ROOT}/platform/app-controller" install
kubectl api-resources --api-group=platform.sarige.dev

#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-kubernetes-platform}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
DEFAULT_TAG="$(git -C "${REPO_ROOT}" rev-parse --short HEAD)"
IMG="${IMG:-app-controller:kind-${DEFAULT_TAG}}"

require_command() {
  local name="$1"
  if ! command -v "${name}" >/dev/null 2>&1; then
    echo "Missing required command: ${name}" >&2
    exit 1
  fi
}

require_command docker
require_command kind

if ! docker info >/dev/null 2>&1; then
  echo "Docker is not running or is not reachable." >&2
  exit 1
fi

if ! kind get clusters | grep -qx "${CLUSTER_NAME}"; then
  echo "kind cluster '${CLUSTER_NAME}' does not exist." >&2
  echo "Create it first: ./infra/kind/scripts/00-create-cluster.sh" >&2
  exit 1
fi

make -C "${REPO_ROOT}/platform/app-controller" docker-build IMG="${IMG}"
kind load docker-image "${IMG}" --name "${CLUSTER_NAME}"
echo "Loaded ${IMG} into kind cluster '${CLUSTER_NAME}'."

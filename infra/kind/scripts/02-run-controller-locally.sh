#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

echo "Starting app-controller against the current kubeconfig context."
echo "Leave this process running while applying App resources from another terminal."

make -C "${REPO_ROOT}/platform/app-controller" run

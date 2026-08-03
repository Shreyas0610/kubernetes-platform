# kind Local Controller Demo Implementation Plan

Date: 2026-08-02

## Scope

Add a local demo workflow for validating the `App` CRD/controller against a disposable kind cluster.

## Steps

1. Add `infra/kind/kind-cluster.yaml` with port mappings reserved for future ingress testing.
2. Add scripts for cluster creation, CRD install, local controller execution, sample app apply, and validation.
3. Add `infra/kind/README.md` explaining purpose, prerequisites, commands, expected results, and production tradeoffs.
4. Update the root README and roadmap to point to the local demo path.
5. Verify shell syntax, YAML parsing, and controller tests.

## Out Of Scope

- Installing ingress-nginx.
- Building and deploying the controller image into kind.
- Running the full kubeadm cluster.
- Pushing or committing anything to GitHub.

# Implementation Plan: Local HTTPS With cert-manager

Date: 2026-08-25

## Scope

Implement local HTTPS support for the kind development cluster.

## Steps

1. Add failing controller tests for TLS Ingress generation and HTTPS status URL.
2. Add `spec.tls` to `AppSpec`.
3. Update `ingressForApp` to add cert-manager issuer annotations and `spec.tls`.
4. Regenerate CRD manifests.
5. Add a local self-signed cert-manager ClusterIssuer manifest.
6. Add kind scripts to install cert-manager and validate HTTPS routing.
7. Update README and roadmap documentation.
8. Run unit/static verification and, when Docker is available, runtime validation.
9. Commit and push the focused milestone.

## Verification

Static verification:

```bash
cd platform/app-controller
make manifests generate fmt
go test ./internal/controller -run 'TestIngressForAppAddsCertManagerTLSWhenEnabled|TestURLForAppUsesHTTPSWhenTLSEnabled' -count=1
make test
cd ../..
bash -n infra/kind/scripts/*.sh
```

Runtime verification:

```bash
./infra/kind/scripts/00-create-cluster.sh
./infra/kind/scripts/05-build-load-controller.sh
./infra/kind/scripts/06-deploy-controller.sh
./infra/kind/scripts/08-install-ingress-nginx.sh
./infra/kind/scripts/10-install-cert-manager.sh
./infra/kind/scripts/11-validate-https-routing.sh
```

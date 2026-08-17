# kind HTTP Routing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Install ingress-nginx in kind and validate HTTP routing from `localhost:8080` to `App/demo-api`.

**Architecture:** The App controller creates `Ingress/demo-api` with `ingressClassName: nginx`. ingress-nginx watches that Ingress and routes host-header traffic from the kind node port mapping to `Service/demo-api` and then to nginx pods. Local controller images use git-SHA tags so kind redeploys do not accidentally reuse an old image.

**Tech Stack:** Kubernetes, kind, ingress-nginx, Bash, curl, Go, controller-runtime.

## Global Constraints

- Keep this milestone HTTP-only.
- Do not install cert-manager in this milestone.
- Use `curl -H 'Host: demo.local' http://localhost:8080/` for validation.
- Pin ingress-nginx to `controller-v1.15.1`.

---

### Task 1: Make App Ingress Class Explicit

**Files:**
- Modify: `platform/app-controller/internal/controller/app_controller.go`
- Modify: `platform/app-controller/internal/controller/app_controller_test.go`

**Interfaces:**
- Produces: generated Ingresses include `spec.ingressClassName: nginx`.

- [ ] Add a failing unit test assertion that `ingressForApp` sets `IngressClassName` to `nginx`.
- [ ] Run `go test ./internal/controller -run TestIngressForAppBuildsExpectedIngress -count=1` and confirm it fails with missing ingress class.
- [ ] Set `ingress.Spec.IngressClassName` to `nginx` in `ingressForApp`.
- [ ] Rerun the focused test and confirm it passes.

### Task 2: Add kind ingress-nginx Scripts

**Files:**
- Create: `infra/kind/scripts/08-install-ingress-nginx.sh`
- Create: `infra/kind/scripts/09-validate-http-routing.sh`

**Interfaces:**
- Consumes: kind cluster `kubernetes-platform`.
- Consumes: sample `platform/app-controller/config/samples/platform_v1alpha1_app.yaml`.
- Produces: repeatable HTTP routing validation.

- [ ] Add an idempotent ingress-nginx install script using the upstream kind manifest.
- [ ] Add a validation script that checks kind port mapping, ingress-nginx readiness, controller readiness, App rollout, Service endpoints, Ingress host/class, and nginx HTTP response.
- [ ] Make scripts executable.
- [ ] Run `bash -n infra/kind/scripts/*.sh`.

### Task 3: Update Documentation

**Files:**
- Modify: `README.md`
- Modify: `infra/kind/README.md`
- Modify: `docs/architecture-and-roadmap.md`
- Create: `docs/superpowers/specs/2026-08-17-kind-http-routing-design.md`

**Interfaces:**
- Produces: project documentation that explains Ingress vs ingress controller.

- [ ] Document the HTTP routing command path.
- [ ] Explain that an Ingress object is desired state and ingress-nginx is the proxy implementing it.
- [ ] Mark HTTPS/cert-manager as the next milestone.

### Task 4: Verify

**Files:**
- Test: `platform/app-controller/internal/controller/app_controller_test.go`
- Test: `infra/kind/scripts/*.sh`

**Interfaces:**
- Consumes: all previous tasks.
- Produces: evidence for merge/commit.

- [ ] Run `bash -n infra/kind/scripts/*.sh`.
- [ ] Run YAML validation for kind and sample manifests.
- [ ] Run `make test` under `platform/app-controller`.
- [ ] If Docker/kind is available, run `08-install-ingress-nginx.sh` and `09-validate-http-routing.sh`.

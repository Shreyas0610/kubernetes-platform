# kind In-Cluster Controller Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a repeatable kind workflow that builds, loads, deploys, and validates the `app-controller` as an in-cluster Kubernetes Deployment.

**Architecture:** Use the existing Kubebuilder `Makefile` and manifests as the source of truth. Add thin kind-specific scripts that build `app-controller:kind`, load it into the `kubernetes-platform` kind cluster, deploy with `make deploy IMG=app-controller:kind`, and validate the controller plus sample `App`.

**Tech Stack:** Bash, kind, Docker, kubectl, Go, Kubebuilder, controller-runtime.

## Global Constraints

- Do not install ingress-nginx or cert-manager in this milestone.
- Do not duplicate Kubebuilder deployment YAML under `infra/kind`.
- Use image tag `app-controller:kind`.
- Leave local scheduler helper files out of Git.

---

### Task 1: Add In-Cluster kind Scripts

**Files:**
- Create: `infra/kind/scripts/05-build-load-controller.sh`
- Create: `infra/kind/scripts/06-deploy-controller.sh`
- Create: `infra/kind/scripts/07-validate-controller-deployment.sh`

**Interfaces:**
- Consumes: existing `platform/app-controller/Makefile` targets `docker-build` and `deploy`.
- Produces: scripts callable from repo root.

- [ ] **Step 1: Write shell scripts**

Create scripts that:

- Build `app-controller:kind`.
- Load that image into the `kubernetes-platform` kind cluster.
- Deploy controller manifests with `IMG=app-controller:kind`.
- Validate controller rollout and sample reconciliation.

- [ ] **Step 2: Make scripts executable**

Run:

```bash
chmod +x infra/kind/scripts/05-build-load-controller.sh infra/kind/scripts/06-deploy-controller.sh infra/kind/scripts/07-validate-controller-deployment.sh
```

- [ ] **Step 3: Verify shell syntax**

Run:

```bash
bash -n infra/kind/scripts/*.sh
```

Expected: no output and exit code `0`.

### Task 2: Update Documentation

**Files:**
- Modify: `infra/kind/README.md`
- Modify: `README.md`
- Modify: `docs/architecture-and-roadmap.md`

**Interfaces:**
- Consumes: scripts from Task 1.
- Produces: a clear local demo path for both local and in-cluster controller modes.

- [ ] **Step 1: Document the new flow**

Update docs to explain:

- `make run` is for local controller development.
- in-cluster Deployment is closer to production.
- Ingress object creation is not the same as HTTPS traffic routing.

- [ ] **Step 2: Verify docs mention the new scripts**

Run:

```bash
rg -n "05-build-load-controller|06-deploy-controller|07-validate-controller-deployment|in-cluster" README.md infra/kind/README.md docs/architecture-and-roadmap.md
```

Expected: each script name appears in the kind README.

### Task 3: Run Final Verification

**Files:**
- Test: `platform/app-controller/internal/controller/app_controller_test.go`
- Test: `infra/kind/scripts/*.sh`

**Interfaces:**
- Consumes: all previous tasks.
- Produces: verified milestone ready for review.

- [ ] **Step 1: Run static checks**

Run:

```bash
bash -n infra/kind/scripts/*.sh
ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_stream(File.read(f)); puts "ok #{f}" }' infra/kind/kind-cluster.yaml platform/app-controller/config/samples/platform_v1alpha1_app.yaml
```

Expected: shell check exits `0`; YAML command prints `ok` for both files.

- [ ] **Step 2: Run controller tests**

Run:

```bash
cd platform/app-controller
make test
```

Expected: all non-e2e Go tests pass.

# App CRD Controller Design

Date: 2026-08-01

## Goal

Build the first Kubernetes-native platform milestone: a Go controller that introduces a high-level `App` custom resource and reconciles it into standard Kubernetes runtime resources.

This creates the portfolio story:

> I built a Kubernetes cluster from scratch, then built a Kubernetes-native platform on top of it using CRDs/controllers.

## Scope

This milestone intentionally focuses on the Kubernetes extension layer, not the full Railway/Render platform.

In scope:

- Go controller scaffolded with Kubebuilder/controller-runtime.
- `App` CRD under `platform.sarige.dev/v1alpha1`.
- Reconciliation of `Deployment`, `Service`, and `Ingress`.
- Status fields and conditions.
- Unit/controller tests using controller-runtime envtest.
- Local development documentation.

Out of scope for this milestone:

- GitHub App integration.
- Docker image build automation.
- Registry push automation.
- cert-manager `Certificate` resources.
- ArgoCD bootstrap.
- Multi-user API server and Postgres metadata store.

These belong in later milestones after the core controller is working.

## Custom Resource

Example:

```yaml
apiVersion: platform.sarige.dev/v1alpha1
kind: App
metadata:
  name: demo-api
spec:
  image: nginx:1.27
  port: 80
  replicas: 2
  host: demo.local
status:
  phase: Ready
  url: http://demo.local
  conditions:
    - type: Ready
      status: "True"
```

## API Fields

`spec.image`

- Required container image.
- Used by the generated `Deployment`.

`spec.port`

- Required container port.
- Must be between `1` and `65535`.
- Used by the `Deployment`, `Service`, and `Ingress`.

`spec.replicas`

- Optional replica count.
- Defaults to `1`.
- Must be at least `1`.

`spec.host`

- Optional hostname.
- When set, the controller creates an `Ingress`.
- When empty, only `Deployment` and `Service` are created.

`status.phase`

- `Pending`, `Reconciling`, `Ready`, or `Error`.

`status.url`

- Set to `http://<host>` when `spec.host` exists.

`status.conditions`

- Uses Kubernetes-style conditions.
- Initial condition types: `Ready`, `Reconciling`, `Stalled`.

## Controller Behavior

The controller watches `App` resources and owns the generated child resources.

For each `App`, it should:

1. Validate the spec.
2. Create or update a `Deployment`.
3. Create or update a `Service`.
4. Create or update an `Ingress` when `spec.host` is set.
5. Set owner references on generated resources.
6. Update `status.phase`, `status.url`, and `status.conditions`.

Generated resource names should match the `App` name.

Generated labels:

```text
app.kubernetes.io/name=<app-name>
app.kubernetes.io/managed-by=kubernetes-platform
platform.sarige.dev/app=<app-name>
```

## Error Handling

Invalid specs should not create partial resources.

The controller should set:

- `status.phase=Error`
- `Ready=False`
- `Stalled=True`
- A clear reason such as `InvalidSpec`

Kubernetes schema validation should prevent invalid `port` and `replicas` values where possible. Runtime validation should still guard controller logic.

## Testing Strategy

Use controller-runtime envtest.

Required tests:

- Creating an `App` creates a matching `Deployment`.
- Creating an `App` creates a matching `Service`.
- Creating an `App` with `spec.host` creates a matching `Ingress`.
- Status updates include `Ready=True` and the expected URL when reconciliation succeeds.
- Invalid specs are rejected by API schema or marked with an error condition.

## Architecture Decision

Use Go with Kubebuilder/controller-runtime.

Why:

- Kubernetes itself is written in Go.
- Most production controllers/operators use Go.
- Kubebuilder creates standard project layout, CRD generation, RBAC manifests, and envtest wiring.
- This directly demonstrates CRDs, reconciliation loops, owner references, and Kubernetes API extension.

## Future Milestones

After this controller works:

1. Add cert-manager `Certificate` support.
2. Add `ConfigMap` and `Secret` references.
3. Add GitHub Actions build-to-GHCR workflow.
4. Add ArgoCD deployment of the controller.
5. Move from local kind testing to the kubeadm cluster.
6. Add a platform API server that creates `App` resources for users.


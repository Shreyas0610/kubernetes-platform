# ☸️ Kubernetes Platform

A production-style developer platform built on a self-managed Kubernetes foundation. The project combines kubeadm infrastructure with a Kubernetes-native **App** API that turns high-level application intent into Deployments, Services, Ingresses, and observable status.

The goal is to understand Kubernetes below the managed-service layer, then extend it using the same reconciliation model that powers Kubernetes itself.

## 📦 Technologies

- Kubernetes
- kubeadm
- containerd
- Go
- Kubebuilder
- controller-runtime
- CustomResourceDefinitions
- envtest and Ginkgo
- Cilium
- ingress-nginx
- cert-manager
- GitHub Actions and GHCR
- Argo CD
- Prometheus, Grafana, Loki, and OpenTelemetry

Items after **envtest and Ginkgo** are roadmap components and are not yet complete.

## 🦄 Features

### Kubernetes-native App API

Developers declare an application through one custom resource:

~~~yaml
apiVersion: platform.sarige.dev/v1alpha1
kind: App
metadata:
  name: example-app
spec:
  image: ghcr.io/example/app:latest
  port: 8080
  replicas: 2
  host: example.local
~~~

The controller reconciles that desired state into:

- A Deployment for application pods
- A ClusterIP Service for stable networking
- An optional Ingress for HTTP routing
- Owner references for lifecycle management
- Platform labels and selectors
- Phase, URL, and Kubernetes-style status conditions

The API validates image, port, and replica fields directly through CRD schema rules.

### Self-managed cluster foundation

The kubeadm infrastructure documents and automates:

- Linux kernel, sysctl, and swap prerequisites
- containerd installation and CRI configuration
- kubeadm, kubelet, and kubectl installation
- Initial control-plane bootstrap
- Worker and additional control-plane joins
- Cluster validation
- etcd backup and recovery
- Node troubleshooting
- Secret and certificate safety practices

### Controller testing

The controller includes focused unit tests and an envtest-backed reconciliation test. Current tests verify:

- Platform labels
- Default and explicit replica counts
- Deployment generation
- Service generation
- Ingress generation
- Owner references
- Status-condition replacement
- Successful reconciliation against a test API server

## 🏗️ Architecture

~~~mermaid
flowchart TD
    A["Developer"] --> B["App custom resource"]
    B --> C["Go controller"]
    C --> D["Deployment"]
    C --> E["Service"]
    C --> F["Ingress"]
    C --> G["Status conditions"]
    H["Container registry"] --> D
    F --> I["Application URL"]
    J["kubeadm cluster"] --- C
~~~

The target platform will add image builds, HTTPS, GitOps, observability, and tenant isolation around this reconciliation core.

## 👩‍🍳 The Process

I began below the managed Kubernetes layer by documenting how Linux nodes, containerd, kubelet, kubeadm, certificates, networking prerequisites, and etcd fit together.

Next, I used Kubebuilder to define a domain-specific **App** resource. Instead of asking developers to manage several Kubernetes objects, the resource captures only the application intent: image, port, replica count, and optional hostname.

I then implemented a controller with **CreateOrPatch**, owner references, and explicit RBAC rules. The controller continuously moves the cluster toward the declared state, which makes the platform declarative and self-healing rather than a collection of deployment scripts.

Finally, I added unit and envtest coverage before moving into real-cluster networking, HTTPS, build pipelines, GitOps, and observability.

## 📁 Repository Structure

~~~text
.
├── docs/
│   ├── architecture-and-roadmap.md
│   └── superpowers/
├── infra/
│   ├── kind/
│   │   └── scripts/
│   └── kubeadm/
│       ├── configs/
│       ├── runbooks/
│       └── scripts/
└── platform/
    └── app-controller/
        ├── api/
        ├── cmd/
        ├── config/
        ├── internal/
        └── test/
~~~

## 🚦 Running the Current Code

### Run controller tests

~~~bash
cd platform/app-controller
make test
~~~

### Inspect the sample App

~~~bash
cat platform/app-controller/config/samples/platform_v1alpha1_app.yaml
~~~

### Validate infrastructure scripts

~~~bash
bash -n infra/kubeadm/scripts/*.sh
~~~

### Run the local kind controller demo

~~~bash
less infra/kind/README.md
~~~

Latest local runtime validation:

- kind cluster `kubernetes-platform` created successfully.
- Controller image `app-controller:kind` built and loaded into kind.
- `app-controller-controller-manager` rolled out in `app-controller-system`.
- Sample `App/demo-api` reconciled into a rolled-out Deployment, ClusterIP Service, and Ingress.
- HTTP routing through ingress-nginx validated locally.

### Review the cluster bootstrap workflow

~~~bash
less infra/kubeadm/README.md
~~~

The self-managed cluster has not yet been fully executed end to end. The current repository contains the controller implementation, tests, configuration, scripts, and operational foundation.

## 🗺️ Roadmap

### Completed

- Define platform architecture and milestones
- Create the **App** CRD
- Implement Deployment, Service, and optional Ingress reconciliation
- Add RBAC and owner references
- Add phase, URL, and condition reporting
- Add unit and envtest coverage
- Create kubeadm bootstrap scripts, templates, validation, and runbooks
- Add a local kind workflow for validating the App CRD/controller
- Deploy the controller into kind as an in-cluster Kubernetes workload
- Add scripts and docs for HTTP routing through ingress-nginx in kind
- Add `spec.tls` support plus cert-manager-compatible Ingress TLS reconciliation
- Add scripts and docs for local HTTPS validation through cert-manager
- Derive App readiness from generated Deployment availability

### Next

1. Bootstrap real control-plane and worker nodes.
2. Install Cilium and validate pod networking.
3. Deploy the controller into the kubeadm cluster.
4. Build application images with GitHub Actions and push to GHCR.
5. Manage platform components with Argo CD.
6. Add Prometheus, Grafana, Loki, and OpenTelemetry.
7. Add namespaces, RBAC, quotas, and network isolation.
8. Add a CLI or web entry point for one-command deployments.

## 📚 What I Learned

- How Kubernetes control-plane and node components work together.
- How containerd and kubelet communicate through the CRI.
- How CRDs extend the Kubernetes API with domain-specific resources.
- How controllers reconcile observed state toward desired state.
- How ingress-nginx and cert-manager work together to expose HTTPS applications.
- Why owner references, idempotency, RBAC, and status conditions matter.
- How to test controllers without a full production cluster.
- How internal developer platforms reduce application deployment complexity.
- Why platform engineering includes operations, recovery, observability, and tenancy.

## 💭 How It Can Be Improved

- Report reconciliation errors from failed child-resource operations.
- Remove stale Ingress objects when a hostname is removed.
- Add health probes, resource requests, limits, secrets, and environment variables to the App API.
- Add end-to-end tests on a real cluster.
- Publish a complete deployment demo and architecture screenshots.
- Measure reconciliation latency, rollout time, and recovery behavior.
- Add safe rollback and promotion workflows.

## 🔐 Safety Notes

Never commit kubeadm join tokens, certificate keys, kubeconfigs, registry credentials, or application secrets. A self-managed API server should be protected with firewalls or private networking, and etcd should be backed up before upgrades or major cluster changes.

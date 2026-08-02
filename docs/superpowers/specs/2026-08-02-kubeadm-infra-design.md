# kubeadm Infrastructure Design

Date: 2026-08-02

## Goal

Create a repeatable, educational kubeadm cluster foundation for the Kubernetes platform project.

The goal is not to hide kubeadm behind a large automation framework yet. The goal is to document and script the exact sequence used to assemble a production-style Kubernetes cluster foundation.

## Scope

In scope:

- `infra/kubeadm` documentation.
- Example inventory for control-plane and worker nodes.
- kubeadm config templates for first control plane and joins.
- Node preparation scripts for containerd and Kubernetes packages.
- Bootstrap scripts for first control-plane, additional control-plane, and worker nodes.
- Validation script.
- Runbooks for etcd backup/restore and node troubleshooting.

Out of scope:

- Ansible or Terraform.
- Real cloud VM provisioning.
- Secret material committed to git.
- Fully automated remote SSH orchestration.
- Cilium, MetalLB, ingress-nginx, cert-manager installation. Those are follow-up `infra/addons` milestones.

## Target Architecture

Production-style target:

```text
3 control-plane nodes
2-3 worker nodes
containerd runtime
kubeadm bootstrap
API server load balancer endpoint
CNI installed after kubeadm init
```

Practical first pass:

```text
1 control-plane node
1-2 worker nodes
same scripts and config shape
documented path to HA control-plane later
```

## Design Principles

- Keep scripts readable and idempotent where practical.
- Prefer explicit commands over hidden orchestration.
- Use kubeadm config files instead of long ad hoc command lines.
- Avoid committing join tokens, certificate keys, kubeconfigs, or private IPs for real machines.
- Make every step explain what Kubernetes component it prepares.
- Separate node preparation from control-plane initialization.

## Directory Layout

```text
infra/kubeadm/
  README.md
  inventory.example.yaml
  configs/
    kubeadm-control-plane.yaml
    kubeadm-join-control-plane.yaml
    kubeadm-join-worker.yaml
  scripts/
    00-prereqs.sh
    01-containerd.sh
    02-kubernetes-packages.sh
    03-init-first-control-plane.sh
    04-join-control-plane.sh
    05-join-worker.sh
    06-validate-cluster.sh
  runbooks/
    etcd-backup-restore.md
    node-troubleshooting.md
```

## Validation

Static verification:

- Shell scripts pass `bash -n`.
- YAML parses with Ruby `Psych` if available.
- No placeholder secrets such as real tokens are committed.

Operational verification, when nodes exist:

- `kubectl get nodes -o wide`
- `kubectl -n kube-system get pods`
- `kubectl get --raw='/readyz?verbose'`
- `kubectl -n kube-system logs -l k8s-app=kube-dns --tail=50`


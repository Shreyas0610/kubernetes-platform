# kubeadm Cluster Foundation

This directory documents the self-managed Kubernetes cluster foundation for the platform.

The goal is to understand how Kubernetes is assembled before the platform controller is deployed onto it. These files are intentionally explicit: they show the node preparation, container runtime setup, kubeadm bootstrap, join flow, and validation steps.

## Target Architecture

Production-style target:

```text
3 control-plane nodes
2-3 worker nodes
1 stable API server load balancer endpoint
containerd runtime
kubeadm bootstrap
Cilium CNI installed after kubeadm init
```

Practical first pass:

```text
1 control-plane node
1-2 worker nodes
same scripts and config shape
HA control-plane added later
```

## Why Each Piece Exists

| Component | Purpose |
|---|---|
| `containerd` | CRI runtime that kubelet uses to run containers |
| `kubelet` | Node agent that starts Pods and reports node state |
| `kubeadm` | Bootstrap tool for control-plane and node joins |
| `kubectl` | Admin/client CLI for the Kubernetes API |
| API server load balancer | Stable endpoint for HA control-plane nodes |
| etcd | Durable key-value store for cluster state |
| CNI | Pod networking implementation; installed after `kubeadm init` |

## Bootstrap Order

Run on every node:

```bash
sudo ./scripts/00-prereqs.sh
sudo ./scripts/01-containerd.sh
sudo ./scripts/02-kubernetes-packages.sh
```

Run on the first control-plane node:

```bash
sudo ./scripts/03-init-first-control-plane.sh ./configs/kubeadm-control-plane.yaml
```

Install a CNI before expecting the cluster to become fully ready. The project roadmap recommends Cilium, but the CNI add-on is a separate milestone.

Run on each additional control-plane node:

```bash
sudo ./scripts/04-join-control-plane.sh ./configs/kubeadm-join-control-plane.yaml
```

Run on each worker node:

```bash
sudo ./scripts/05-join-worker.sh ./configs/kubeadm-join-worker.yaml
```

Validate from an admin machine with kubeconfig access:

```bash
./scripts/06-validate-cluster.sh
```

## Safety Notes

- Do not commit real kubeadm join tokens.
- Do not commit certificate keys from `--upload-certs`.
- Do not commit real kubeconfigs.
- Do not expose the Kubernetes API publicly without firewalling or VPN access.
- Keep etcd on fast local SSD-backed storage.
- Snapshot etcd before upgrades and after major cluster changes.

## Files

```text
inventory.example.yaml                 Example topology and network values
configs/kubeadm-control-plane.yaml      First control-plane kubeadm init config
configs/kubeadm-join-control-plane.yaml Additional control-plane join config
configs/kubeadm-join-worker.yaml        Worker join config
scripts/00-prereqs.sh                   Kernel, sysctl, and swap prep
scripts/01-containerd.sh                containerd installation and config
scripts/02-kubernetes-packages.sh       kubeadm/kubelet/kubectl installation
scripts/03-init-first-control-plane.sh  kubeadm init wrapper
scripts/04-join-control-plane.sh        kubeadm control-plane join wrapper
scripts/05-join-worker.sh               kubeadm worker join wrapper
scripts/06-validate-cluster.sh          Cluster validation checks
runbooks/etcd-backup-restore.md         etcd snapshot and restore notes
runbooks/node-troubleshooting.md        Node failure troubleshooting notes
```

## Next Add-on Milestones

After the kubeadm foundation works:

1. Install Cilium and validate pod networking.
2. Install MetalLB or cloud LoadBalancer integration.
3. Install ingress-nginx.
4. Install cert-manager.
5. Deploy the `platform/app-controller` onto the cluster.


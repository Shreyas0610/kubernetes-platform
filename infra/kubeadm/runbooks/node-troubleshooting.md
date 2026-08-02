# Node Troubleshooting Runbook

Use this runbook when a node is `NotReady`, Pods are stuck, or kubeadm join fails.

## First Checks

```bash
kubectl get nodes -o wide
kubectl describe node <node-name>
kubectl -n kube-system get pods -o wide
```

On the affected node:

```bash
sudo systemctl status kubelet --no-pager
sudo journalctl -u kubelet -n 200 --no-pager
sudo systemctl status containerd --no-pager
sudo crictl ps
sudo crictl pods
```

## Common Causes

| Symptom | Likely Cause | Check |
|---|---|---|
| Node `NotReady` | CNI missing or broken | `kubectl -n kube-system get pods` |
| kubelet cannot start Pods | containerd down or CRI config wrong | `systemctl status containerd` |
| DNS failures | CoreDNS not ready or CNI issue | `kubectl -n kube-system logs -l k8s-app=kube-dns` |
| Join fails with token error | Expired token or wrong CA hash | `kubeadm token list` |
| Pods stuck Pending | No schedulable nodes or resource pressure | `kubectl describe pod <pod>` |

## Regenerate Join Command

On a control-plane node:

```bash
kubeadm token create --print-join-command
```

For additional control-plane joins, also generate or reuse the certificate key:

```bash
sudo kubeadm init phase upload-certs --upload-certs
```

## Reset A Failed Node

Only run this on a node you intend to remove or rejoin:

```bash
sudo kubeadm reset
sudo systemctl restart containerd
```

Then rerun the appropriate join script.


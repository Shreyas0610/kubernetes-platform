# etcd Backup And Restore Runbook

etcd stores Kubernetes cluster state. If etcd is lost, the cluster loses its source of truth.

## Snapshot

Run on a healthy control-plane node:

```bash
sudo ETCDCTL_API=3 etcdctl snapshot save /var/backups/etcd/snapshot.db \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key
```

Verify the snapshot:

```bash
sudo ETCDCTL_API=3 etcdctl snapshot status /var/backups/etcd/snapshot.db --write-out=table
```

## Restore Notes

Restoring stacked etcd is disruptive. Practice this in a lab before relying on it.

High-level flow:

1. Stop kubelet on the affected control-plane node.
2. Move the old etcd data directory out of the way.
3. Restore the snapshot with `etcdctl snapshot restore`.
4. Point the static etcd pod manifest at the restored data directory.
5. Start kubelet.
6. Verify API server health with `kubectl get --raw='/readyz?verbose'`.

## Production Practice

- Snapshot before cluster upgrades.
- Store snapshots off-node.
- Encrypt backups at rest.
- Test restores, not just backup creation.


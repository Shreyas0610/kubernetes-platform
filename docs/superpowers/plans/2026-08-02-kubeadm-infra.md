# kubeadm Infrastructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first `infra/kubeadm` milestone with readable scripts, kubeadm config templates, inventory, and runbooks for a production-style self-managed Kubernetes cluster.

**Architecture:** The milestone separates node preparation, container runtime setup, Kubernetes package installation, control-plane bootstrap, node joins, and validation. It documents a 3-control-plane/2-3-worker target while supporting a smaller first-pass cluster.

**Tech Stack:** bash, kubeadm, kubelet, kubectl, containerd, systemd, YAML, Ubuntu/Debian apt repositories.

## Global Constraints

- Repository root is `/Users/sarige/kubernetes-platform`.
- Files live under `infra/kubeadm`.
- Do not commit real join tokens, certificate keys, kubeconfigs, IPs, or private hostnames.
- Scripts must use `set -euo pipefail`.
- Scripts must be readable learning artifacts, not opaque automation.
- Add validation docs and shell syntax verification.

---

### Task 1: Add kubeadm Directory And README

**Files:**
- Create: `infra/kubeadm/README.md`
- Create: `infra/kubeadm/inventory.example.yaml`

**Interfaces:**
- Produces human-readable kubeadm bootstrap sequence.
- Produces example host inventory used by future automation.

Steps:

- [ ] Create the directory layout.
- [ ] Write README with purpose, architecture, node roles, bootstrap order, prerequisites, and safety notes.
- [ ] Write `inventory.example.yaml` with placeholder control-plane, worker, load balancer, pod CIDR, and service CIDR values.
- [ ] Verify YAML parses.
- [ ] Commit.

### Task 2: Add kubeadm Config Templates

**Files:**
- Create: `infra/kubeadm/configs/kubeadm-control-plane.yaml`
- Create: `infra/kubeadm/configs/kubeadm-join-control-plane.yaml`
- Create: `infra/kubeadm/configs/kubeadm-join-worker.yaml`

**Interfaces:**
- Produces kubeadm config templates that users copy and replace placeholder values in.

Steps:

- [ ] Add first control-plane init config.
- [ ] Add additional control-plane join config.
- [ ] Add worker join config.
- [ ] Verify YAML parses.
- [ ] Commit.

### Task 3: Add Node Setup Scripts

**Files:**
- Create: `infra/kubeadm/scripts/00-prereqs.sh`
- Create: `infra/kubeadm/scripts/01-containerd.sh`
- Create: `infra/kubeadm/scripts/02-kubernetes-packages.sh`

**Interfaces:**
- Produces scripts to prepare Linux nodes for kubeadm.

Steps:

- [ ] Add kernel module/sysctl/swap prerequisite script.
- [ ] Add containerd installation/configuration script.
- [ ] Add Kubernetes package installation script.
- [ ] Run `bash -n` on scripts.
- [ ] Commit.

### Task 4: Add Bootstrap And Validation Scripts

**Files:**
- Create: `infra/kubeadm/scripts/03-init-first-control-plane.sh`
- Create: `infra/kubeadm/scripts/04-join-control-plane.sh`
- Create: `infra/kubeadm/scripts/05-join-worker.sh`
- Create: `infra/kubeadm/scripts/06-validate-cluster.sh`

**Interfaces:**
- Produces explicit command wrappers around kubeadm init/join and cluster validation.

Steps:

- [ ] Add first control-plane init wrapper.
- [ ] Add control-plane join wrapper.
- [ ] Add worker join wrapper.
- [ ] Add validation script.
- [ ] Run `bash -n` on scripts.
- [ ] Commit.

### Task 5: Add Runbooks And Roadmap Link

**Files:**
- Create: `infra/kubeadm/runbooks/etcd-backup-restore.md`
- Create: `infra/kubeadm/runbooks/node-troubleshooting.md`
- Modify: `docs/architecture-and-roadmap.md`

**Interfaces:**
- Produces operational documentation for failure handling and recruiter review.

Steps:

- [ ] Add etcd backup/restore runbook.
- [ ] Add node troubleshooting runbook.
- [ ] Link `infra/kubeadm` from architecture roadmap.
- [ ] Run static verification.
- [ ] Commit.


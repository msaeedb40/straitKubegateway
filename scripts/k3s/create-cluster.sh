#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — k3s: Create Cluster (via k3d)
# ============================================================================
# Creates a lightweight k3s cluster via k3d with Flannel, network policy,
# Traefik, and ServiceLB disabled — straitKubegateway provides its own CNI
# and kube-proxy replacement.
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds k3d kubectl docker

print_banner "k3s (k3d) — Create Cluster"

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
export K3S_SERVERS="${K3S_SERVERS:-1}"
export K3S_AGENTS="${K3S_AGENTS:-2}"

log_info "Creating k3d cluster '${CLUSTER_NAME}' (${K3S_SERVERS} server(s), ${K3S_AGENTS} agent(s)) ..."

k3d cluster create "${CLUSTER_NAME}" \
  --servers "${K3S_SERVERS}" \
  --agents "${K3S_AGENTS}" \
  --k3s-arg "--flannel-backend=none@server:*" \
  --k3s-arg "--disable-network-policy@server:*" \
  --k3s-arg "--disable=traefik@server:*" \
  --k3s-arg "--disable=servicelb@server:*" \
  --k3s-arg "--disable-kube-proxy@server:*" \
  --no-lb \
  --wait

log_info "Verifying cluster is ready ..."
kubectl cluster-info --context "k3d-${CLUSTER_NAME}"
kubectl get nodes

log_success "k3d cluster '${CLUSTER_NAME}' is ready (flannel/kube-proxy disabled)."
log_info "Next step: run ./install.sh to deploy straitKubegateway."

#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — k3s: Delete Cluster (via k3d)
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds k3d

print_banner "k3s (k3d) — Delete Cluster"

log_info "Deleting k3d cluster '${CLUSTER_NAME}' ..."
k3d cluster delete "${CLUSTER_NAME}"

log_success "k3d cluster '${CLUSTER_NAME}' deleted."

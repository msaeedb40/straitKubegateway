#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — kind: Delete Cluster
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds kind

print_banner "kind — Delete Cluster"

log_info "Deleting Kind cluster '${CLUSTER_NAME}' ..."
kind delete cluster --name "${CLUSTER_NAME}"

log_success "Kind cluster '${CLUSTER_NAME}' deleted."

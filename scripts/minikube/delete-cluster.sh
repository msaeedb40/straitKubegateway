#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — minikube: Delete Cluster
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds minikube

print_banner "minikube — Delete Cluster"

log_info "Deleting Minikube cluster '${CLUSTER_NAME}' ..."
minikube delete --profile="${CLUSTER_NAME}"

log_success "Minikube cluster '${CLUSTER_NAME}' deleted."

#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — minikube: Uninstall
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds helm kubectl

print_banner "minikube — Uninstall straitKubegateway"

helm_uninstall

log_success "straitKubegateway removed from Minikube cluster '${CLUSTER_NAME}'."

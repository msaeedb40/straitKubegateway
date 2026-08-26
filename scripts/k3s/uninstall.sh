#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — k3s: Uninstall
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds helm kubectl

print_banner "k3s (k3d) — Uninstall straitKubegateway"

helm_uninstall

log_success "straitKubegateway removed from k3d cluster '${CLUSTER_NAME}'."

#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — kind: Uninstall
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds helm kubectl

print_banner "kind — Uninstall straitKubegateway"

helm_uninstall

log_success "straitKubegateway removed from Kind cluster '${CLUSTER_NAME}'."

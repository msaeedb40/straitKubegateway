#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — k3s: Install
# ============================================================================
# Loads locally-built Docker images into k3d and installs
# straitKubegateway via Helm.
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds k3d kubectl helm

print_banner "k3s (k3d) — Install straitKubegateway"

# ---------------------------------------------------------------------------
# Load images into k3d (if built locally)
# ---------------------------------------------------------------------------
IMAGES=(
  "${REGISTRY}/straitd:${IMAGE_TAG}"
  "${REGISTRY}/sg-controller:${IMAGE_TAG}"
  "${REGISTRY}/sg-cli:${IMAGE_TAG}"
  "${REGISTRY}/ui:${IMAGE_TAG}"
)

for img in "${IMAGES[@]}"; do
  if docker image inspect "${img}" &>/dev/null; then
    log_info "Importing image '${img}' into k3d cluster ..."
    k3d image import "${img}" --cluster "${CLUSTER_NAME}"
  else
    log_warn "Image '${img}' not found locally — skipping import (will pull from registry)."
  fi
done

# ---------------------------------------------------------------------------
# Helm install with k3s-specific overrides
# ---------------------------------------------------------------------------
helm_install \
  --set straitd.kubeProxyReplacement=true \
  --set straitd.kubeProxyMode=none \
  --set straitd.image.pullPolicy=IfNotPresent \
  --set sgController.image.pullPolicy=IfNotPresent

wait_for_pods "${NAMESPACE}"

log_success "straitKubegateway installed on k3d cluster '${CLUSTER_NAME}'."

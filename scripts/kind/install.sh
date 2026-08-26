#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — kind: Install
# ============================================================================
# Loads locally-built Docker images into Kind and installs
# straitKubegateway via Helm.
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds kind kubectl helm docker

print_banner "kind — Install straitKubegateway"

# ---------------------------------------------------------------------------
# Load images into Kind (if built locally)
# ---------------------------------------------------------------------------
IMAGES=(
  "${REGISTRY}/straitd:${IMAGE_TAG}"
  "${REGISTRY}/sg-controller:${IMAGE_TAG}"
  "${REGISTRY}/sg-cli:${IMAGE_TAG}"
  "${REGISTRY}/ui:${IMAGE_TAG}"
)

for img in "${IMAGES[@]}"; do
  if docker image inspect "${img}" &>/dev/null; then
    log_info "Loading image '${img}' into Kind cluster ..."
    kind load docker-image "${img}" --name "${CLUSTER_NAME}"
  else
    log_warn "Image '${img}' not found locally — skipping load (will pull from registry)."
  fi
done

# ---------------------------------------------------------------------------
# Helm install with Kind-specific overrides
# ---------------------------------------------------------------------------
helm_install \
  --set straitd.image.repository="${REGISTRY}/straitd" \
  --set sgController.image.repository="${REGISTRY}/sg-controller" \
  --set straitd.image.pullPolicy=IfNotPresent \
  --set sgController.image.pullPolicy=IfNotPresent

wait_for_pods "${NAMESPACE}"

log_success "straitKubegateway installed on Kind cluster '${CLUSTER_NAME}'."

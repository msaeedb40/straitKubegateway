#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — minikube: Install
# ============================================================================
# Builds and loads images into Minikube (via docker-env) and installs
# straitKubegateway via Helm.
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds minikube kubectl helm

print_banner "minikube — Install straitKubegateway"

# ---------------------------------------------------------------------------
# Load images into Minikube
# ---------------------------------------------------------------------------
IMAGES=(
  "${REGISTRY}/straitd:${IMAGE_TAG}"
  "${REGISTRY}/sg-controller:${IMAGE_TAG}"
  "${REGISTRY}/sg-cli:${IMAGE_TAG}"
  "${REGISTRY}/ui:${IMAGE_TAG}"
)

for img in "${IMAGES[@]}"; do
  if docker image inspect "${img}" &>/dev/null 2>&1; then
    log_info "Loading image '${img}' into Minikube ..."
    minikube image load "${img}" --profile="${CLUSTER_NAME}"
  else
    log_warn "Image '${img}' not found locally — skipping load (will pull from registry)."
  fi
done

# ---------------------------------------------------------------------------
# Helm install
# ---------------------------------------------------------------------------
helm_install \
  --set straitd.image.repository="${REGISTRY}/straitd" \
  --set sgController.image.repository="${REGISTRY}/sg-controller" \
  --set straitd.image.pullPolicy=IfNotPresent \
  --set sgController.image.pullPolicy=IfNotPresent

wait_for_pods "${NAMESPACE}"

log_success "straitKubegateway installed on Minikube cluster '${CLUSTER_NAME}'."

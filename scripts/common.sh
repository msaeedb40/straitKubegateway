#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — Shared Helper Functions
# ============================================================================
# Source this file from any per-environment script:
#   SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#   source "${SCRIPT_DIR}/../common.sh"
# ============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Colors & Logging
# ---------------------------------------------------------------------------
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

log_info()    { echo -e "${CYAN}[INFO]${NC}  $*"; }
log_success() { echo -e "${GREEN}[OK]${NC}    $*"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }

die() { log_error "$@"; exit 1; }

# ---------------------------------------------------------------------------
# Default Variables (override via environment)
# ---------------------------------------------------------------------------
export CLUSTER_NAME="${CLUSTER_NAME:-straitkubegateway}"
export NAMESPACE="${NAMESPACE:-kube-system}"
export HELM_RELEASE="${HELM_RELEASE:-straitkubegateway}"

# Resolve the repo root relative to this file (scripts/ is one level down)
COMMON_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export REPO_ROOT="${REPO_ROOT:-$(cd "${COMMON_DIR}/.." && pwd)}"
export HELM_CHART="${HELM_CHART:-${REPO_ROOT}/straitKubegateway-helm-repo}"

# Image registry & tag
export REGISTRY="${REGISTRY:-ghcr.io/straitkubegateway}"
export IMAGE_TAG="${IMAGE_TAG:-latest}"

# ---------------------------------------------------------------------------
# Prerequisite Checks
# ---------------------------------------------------------------------------
require_cmd() {
  local cmd="$1"
  if ! command -v "${cmd}" &>/dev/null; then
    die "'${cmd}' is required but not found in PATH."
  fi
}

require_cmds() {
  for cmd in "$@"; do
    require_cmd "${cmd}"
  done
}

# ---------------------------------------------------------------------------
# Gateway API CRDs Helper
# ---------------------------------------------------------------------------
export GATEWAY_API_CRDS_URL="${GATEWAY_API_CRDS_URL:-https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/experimental-install.yaml}"

install_gateway_api_crds() {
  log_info "Ensuring Gateway API experimental CRDs are installed ..."
  kubectl apply -f "${GATEWAY_API_CRDS_URL}"
}

# ---------------------------------------------------------------------------
# Helm Helpers
# ---------------------------------------------------------------------------
helm_install() {
  local extra_args=("$@")

  install_gateway_api_crds

  log_info "Installing Helm release '${HELM_RELEASE}' in namespace '${NAMESPACE}' ..."
  helm upgrade --install "${HELM_RELEASE}" "${HELM_CHART}" \
    --namespace "${NAMESPACE}" \
    --create-namespace \
    --set global.imageRegistry="${REGISTRY}" \
    --set straitd.image.tag="${IMAGE_TAG}" \
    --set sgController.image.tag="${IMAGE_TAG}" \
    --wait \
    --timeout 300s \
    "${extra_args[@]+"${extra_args[@]}"}"

  log_success "Helm release '${HELM_RELEASE}' installed."
}

helm_uninstall() {
  log_info "Uninstalling Helm release '${HELM_RELEASE}' from namespace '${NAMESPACE}' ..."
  if helm status "${HELM_RELEASE}" --namespace "${NAMESPACE}" &>/dev/null; then
    helm uninstall "${HELM_RELEASE}" --namespace "${NAMESPACE}" --wait
    log_success "Helm release '${HELM_RELEASE}' uninstalled."
  else
    log_warn "Helm release '${HELM_RELEASE}' not found — nothing to uninstall."
  fi
}

# ---------------------------------------------------------------------------
# Kubectl Helpers
# ---------------------------------------------------------------------------
wait_for_pods() {
  local ns="${1:-${NAMESPACE}}"
  local timeout="${2:-120s}"
  log_info "Waiting for all pods in namespace '${ns}' to be ready (timeout ${timeout}) ..."
  kubectl wait --for=condition=Ready pods --all \
    --namespace "${ns}" \
    --timeout="${timeout}" 2>/dev/null || true
  kubectl get pods --namespace "${ns}"
}

# ---------------------------------------------------------------------------
# Summary Banner
# ---------------------------------------------------------------------------
print_banner() {
  local env_name="$1"
  echo ""
  echo "============================================================"
  echo "  straitKubegateway — ${env_name}"
  echo "============================================================"
  echo "  Cluster  : ${CLUSTER_NAME}"
  echo "  Namespace: ${NAMESPACE}"
  echo "  Release  : ${HELM_RELEASE}"
  echo "  Chart    : ${HELM_CHART}"
  echo "============================================================"
  echo ""
}

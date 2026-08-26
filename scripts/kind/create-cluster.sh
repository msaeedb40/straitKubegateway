#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — kind: Create Cluster
# ============================================================================
# Creates a local Kind cluster with the default CNI disabled so that
# straitKubegateway can be installed as the sole CNI/kube-proxy replacement.
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds kind kubectl docker

print_banner "kind — Create Cluster"

# Kind cluster configuration (inline)
KIND_CONFIG=$(cat <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ${CLUSTER_NAME}
networking:
  # Disable the default CNI so straitKubegateway replaces it
  disableDefaultCNI: true
  # Disable kube-proxy — straitKubegateway provides full kube-proxy replacement
  kubeProxyMode: "none"
nodes:
  - role: control-plane
  - role: worker
  - role: worker
EOF
)

log_info "Creating Kind cluster '${CLUSTER_NAME}' ..."
echo "${KIND_CONFIG}" | kind create cluster --config=-

log_info "Verifying cluster is ready ..."
kubectl cluster-info --context "kind-${CLUSTER_NAME}"
kubectl get nodes

log_success "Kind cluster '${CLUSTER_NAME}' is ready."
log_info "Next step: run ./install.sh to deploy straitKubegateway."

#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — minikube: Create Cluster
# ============================================================================
# Starts a Minikube cluster with the default CNI and kube-proxy disabled
# so straitKubegateway can serve as the sole CNI and kube-proxy replacement.
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../common.sh"

require_cmds minikube kubectl

print_banner "minikube — Create Cluster"

export MINIKUBE_DRIVER="${MINIKUBE_DRIVER:-docker}"
export MINIKUBE_CPUS="${MINIKUBE_CPUS:-2}"
export MINIKUBE_MEMORY="${MINIKUBE_MEMORY:-4096}"
export MINIKUBE_NODES="${MINIKUBE_NODES:-2}"

log_info "Starting Minikube cluster '${CLUSTER_NAME}' ..."
minikube start \
  --profile="${CLUSTER_NAME}" \
  --driver="${MINIKUBE_DRIVER}" \
  --cpus="${MINIKUBE_CPUS}" \
  --memory="${MINIKUBE_MEMORY}" \
  --nodes="${MINIKUBE_NODES}" \
  --cni=false \
  --network-plugin=cni \
  --extra-config=kubeadm.skip-phases=addon/kube-proxy

log_info "Verifying cluster is ready ..."
kubectl cluster-info --context "${CLUSTER_NAME}"
kubectl get nodes

log_success "Minikube cluster '${CLUSTER_NAME}' is ready."
log_info "Next step: run ./install.sh to deploy straitKubegateway."

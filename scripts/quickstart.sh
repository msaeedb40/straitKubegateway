#!/usr/bin/env bash
# Copyright 2026 straitKubegateway authors.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

CLUSTER_NAME=${CLUSTER_NAME:-strait-dev}
REGISTRY=${REGISTRY:-ghcr.io/straitkubegateway}
IMAGE_TAG=${IMAGE_TAG:-latest}
HELM_TIMEOUT=${HELM_TIMEOUT:-5m}

echo "=========================================================="
echo " Starting straitKubegateway Local Development Cluster     "
echo "=========================================================="

# ============================================================================
# Prerequisite checks
# ============================================================================

check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "Error: '$1' is required but not found. Please install it first."
        exit 1
    fi
}

echo "==> Checking prerequisites..."
check_command docker
check_command kind
check_command helm
check_command kubectl
echo "    ✓ All prerequisites found"

# ============================================================================
# Create Kind cluster
# ============================================================================

echo "==> Creating Kind cluster '${CLUSTER_NAME}' with disabled default CNI and kube-proxy..."
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "    Cluster '${CLUSTER_NAME}' already exists, skipping creation"
else
    cat <<EOF | kind create cluster --name "${CLUSTER_NAME}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true
  kubeProxyMode: "none"
nodes:
- role: control-plane
- role: worker
- role: worker
EOF
fi

# ============================================================================
# Build and load container images
# ============================================================================

echo "==> Building container images..."
make docker-build

echo "==> Loading images into Kind cluster..."
for component in straitd sg-controller sg-cli ui; do
    echo "    Loading ${component}..."
    kind load docker-image "${REGISTRY}/${component}:${IMAGE_TAG}" --name "${CLUSTER_NAME}"
done
echo "    ✓ All images loaded"

# ============================================================================
# Install via Helm
# ============================================================================

echo "==> Installing straitKubegateway via Helm..."
helm upgrade --install straitkubegateway ./straitKubegateway-helm-repo \
  --namespace kube-system \
  --set straitd.image.repository="${REGISTRY}/straitd" \
  --set straitd.image.tag="${IMAGE_TAG}" \
  --set sgController.image.repository="${REGISTRY}/sg-controller" \
  --set sgController.image.tag="${IMAGE_TAG}" \
  --wait \
  --timeout "${HELM_TIMEOUT}"

# ============================================================================
# Verify installation
# ============================================================================

echo ""
echo "==> Verifying installation..."
kubectl get pods -n kube-system -l app.kubernetes.io/name=straitKubegateway-helm-repo

echo ""
echo "=========================================================="
echo " ✓ straitKubegateway installed and ready!                 "
echo "=========================================================="
echo ""
echo " Cluster:  ${CLUSTER_NAME}"
echo " Registry: ${REGISTRY}"
echo " Tag:      ${IMAGE_TAG}"
echo ""
echo " Useful commands:"
echo "   kubectl get pods -n kube-system"
echo "   kubectl get straitnetwork"
echo "   kubectl get straitnodes"
echo ""

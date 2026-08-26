#!/usr/bin/env bash
# Copyright 2026 straitKubegateway authors.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

CLUSTER_NAME=${CLUSTER_NAME:-strait-dev}

echo "=========================================================="
echo " Starting straitKubegateway Local Development Cluster     "
echo "=========================================================="

# Check for kind or minikube
if command -v kind &> /dev/null; then
    echo "==> Creating Kind cluster with disabled default CNI and kube-proxy..."
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
else
    echo "Error: kind is required for the local development quickstart."
    exit 1
fi

echo "==> Building container images..."
make docker-build

echo "==> Loading images into Kind cluster..."
kind load docker-image ghcr.io/straitkubegateway/straitd:latest --name "${CLUSTER_NAME}"
kind load docker-image ghcr.io/straitkubegateway/sg-controller:latest --name "${CLUSTER_NAME}"

echo "==> Installing straitKubegateway via Helm..."
helm upgrade --install straitkubegateway ./straitKubegateway-helm-repo \
  --namespace kube-system \
  --set straitd.image.tag=latest \
  --set sgController.image.tag=latest

echo "✓ straitKubegateway installed and ready!"
kubectl get pods -n kube-system -l app.kubernetes.io/name=straitKubegateway-helm-repo

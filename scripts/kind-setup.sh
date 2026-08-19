#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME=${CLUSTER_NAME:-strait-dev}

echo "Creating kind cluster '$CLUSTER_NAME' without default CNI / kube-proxy..."

cat <<EOF | kind create cluster --name "$CLUSTER_NAME" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true
  kubeProxyMode: "none"
nodes:
  - role: control-plane
  - role: worker
EOF

echo "kind cluster '$CLUSTER_NAME' created successfully."

#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME=${CLUSTER_NAME:-strait-0}

echo "Creating kind cluster '$CLUSTER_NAME' without default CNI / kube-proxy..."

cat <<EOF | kind create cluster --name "$CLUSTER_NAME" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true
  kubeProxyMode: "none"
  podSubnet: 10.25.0.0/16 

nodes:
  - role: control-plane
  - role: worker
  - role: worker
EOF

echo "kind cluster '$CLUSTER_NAME' created successfully."

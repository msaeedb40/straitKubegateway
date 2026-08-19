#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME=${CLUSTER_NAME:-strait-0}
NAMESPACE="strait-system"
REGISTRY=${REGISTRY:-ghcr.io/straitkubegateway}
IMAGE_TAG=${IMAGE_TAG:-dev}

echo "Deploying straitKubegateway locally to cluster '$CLUSTER_NAME'..."
helm upgrade --install straitkubegateway ./straitKubegateway-helm \
    --namespace "$NAMESPACE" \
    --create-namespace \
    --set global.imageRegistry="$REGISTRY" \
    --set agent.image.repository="$REGISTRY/straitd" \
    --set agent.image.tag="$IMAGE_TAG" \
    --set agent.image.pullPolicy="Never" \
    --set controller.image.repository="$REGISTRY/sg-controller" \
    --set controller.image.tag="$IMAGE_TAG" \
    --set controller.image.pullPolicy="Never" \
    --set ui.image.repository="$REGISTRY/ui" \
    --set ui.image.tag="$IMAGE_TAG" \
    --set ui.image.pullPolicy="Never"

echo "straitKubegateway deployed successfully to '$CLUSTER_NAME'."


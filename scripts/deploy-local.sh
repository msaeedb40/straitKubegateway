#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="strait-system"

echo "Deploying straitKubegateway locally..."
helm upgrade --install straitkubegateway ./straitKubegateway-helm \
    --namespace "$NAMESPACE" \
    --create-namespace \
    --set global.imageRegistry="ghcr.io/straitkubegateway" \
    --set agent.image.pullPolicy="Never"

echo "straitKubegateway deployed successfully."

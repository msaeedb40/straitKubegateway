#!/usr/bin/env bash
set -euo pipefail

REGISTRY=${REGISTRY:-ghcr.io/straitkubegateway}
IMAGE_TAG=${IMAGE_TAG:-dev}

echo "Building container images with tag: $IMAGE_TAG..."
docker build -t "$REGISTRY/straitd:$IMAGE_TAG" -f build/Dockerfile.straitd .
docker build -t "$REGISTRY/sg-controller:$IMAGE_TAG" -f build/Dockerfile.sg-controller .
docker build -t "$REGISTRY/sg-cli:$IMAGE_TAG" -f build/Dockerfile.sg-cli .
docker build -t "$REGISTRY/ui:$IMAGE_TAG" -f build/Dockerfile.ui .
echo "Container image builds completed."

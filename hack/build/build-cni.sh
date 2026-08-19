#!/usr/bin/env bash
set -euo pipefail

echo "Building straitKubegateway CNI plugin..."
mkdir -p bin
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/straitkubegateway-cni ./cni/plugin
echo "CNI plugin built successfully at bin/straitkubegateway-cni."

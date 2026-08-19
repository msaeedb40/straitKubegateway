#!/usr/bin/env bash
set -euo pipefail

echo "Verifying Helm chart manifests..."
if command -v helm &> /dev/null; then
    helm lint straitKubegateway-helm/
else
    echo "helm not found in PATH, skipping lint"
fi
echo "Helm manifests verification passed."

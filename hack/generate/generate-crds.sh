#!/usr/bin/env bash
set -euo pipefail

echo "==> Generating CRD manifests..."
if command -v controller-gen &> /dev/null; then
    controller-gen crd paths="./api/..." output:crd:artifacts:config=straitKubegateway-helm-repo/charts/crds
else
    echo "controller-gen not installed, skipping CRD regeneration."
fi
echo "✓ CRD generation complete."

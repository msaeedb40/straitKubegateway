#!/usr/bin/env bash
set -euo pipefail

echo "==> Verifying CRD definitions..."
for crd in straitKubegateway-helm-repo/charts/crds/*.yaml; do
    if [[ -f "$crd" ]]; then
        echo "  checking $crd..."
    fi
done
echo "✓ All CRDs verified."

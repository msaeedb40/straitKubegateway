#!/usr/bin/env bash
set -euo pipefail

echo "Verifying CRD definitions..."
if [ ! -f "straitKubegateway-helm/crds/strait.io_crds.yaml" ]; then
    echo "Error: strait.io_crds.yaml not found!"
    exit 1
fi
echo "CRD verification passed."

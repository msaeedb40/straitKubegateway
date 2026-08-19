#!/usr/bin/env bash
set -euo pipefail

echo "Generating CRD manifests..."
if command -v controller-gen &> /dev/null; then
    controller-gen crd paths="./api/..." output:crd:artifacts:config=straitKubegateway-helm/charts/crd/templates
else
    echo "controller-gen not in PATH, using go run"
    go run sigs.k8s.io/controller-tools/cmd/controller-gen crd paths="./api/..." output:crd:artifacts:config=straitKubegateway-helm/charts/crd/templates || true
fi

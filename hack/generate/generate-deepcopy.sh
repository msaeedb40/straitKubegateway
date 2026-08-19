#!/usr/bin/env bash
set -euo pipefail

echo "Generating deepcopy methods..."
if command -v controller-gen &> /dev/null; then
    controller-gen object:headerFile="hack/boilerplate/license_header.txt" paths="./api/..."
else
    echo "controller-gen not in PATH, using go run"
    go run sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile="hack/boilerplate/license_header.txt" paths="./api/..." || true
fi

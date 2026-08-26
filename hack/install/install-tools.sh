#!/usr/bin/env bash
set -euo pipefail

echo "==> Installing development tools..."
go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0
go install github.com/bufbuild/buf/cmd/buf@v1.30.0
echo "✓ Tools installed successfully."

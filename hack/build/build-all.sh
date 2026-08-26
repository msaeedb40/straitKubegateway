#!/usr/bin/env bash
set -euo pipefail

echo "==> Building all straitKubegateway binaries..."
mkdir -p bin
go build -trimpath -o bin/straitd ./cmd/straitd/
go build -trimpath -o bin/sg-controller ./cmd/sg-controller/
go build -trimpath -o bin/sg-cli ./cmd/sg-cli/
echo "✓ Binaries built successfully in bin/"

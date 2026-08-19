#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BIN_DIR="${ROOT_DIR}/bin"
mkdir -p "${BIN_DIR}"

echo "==> Building all straitKubegateway binaries..."
CGO_ENABLED=0 go build -o "${BIN_DIR}/straitd" "${ROOT_DIR}/cmd/straitd" || true
CGO_ENABLED=0 go build -o "${BIN_DIR}/sg-controller" "${ROOT_DIR}/cmd/sg-controller" || true
CGO_ENABLED=0 go build -o "${BIN_DIR}/sg-cli" "${ROOT_DIR}/cmd/sg-cli" || true
CGO_ENABLED=0 go build -o "${BIN_DIR}/straitkubegateway-cni" "${ROOT_DIR}/cni/plugin" || true

echo "==> Build complete."

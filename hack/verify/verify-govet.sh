#!/usr/bin/env bash
set -euo pipefail

echo "==> Verifying go vet..."
go vet ./...
echo "✓ go vet passed."

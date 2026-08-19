#!/usr/bin/env bash
set -euo pipefail

echo "==> Running unit tests..."
go test -v -race -coverprofile=coverage.out ./pkg/... ./internal/... ./observability/... ./platform/...
echo "==> Unit tests passed."

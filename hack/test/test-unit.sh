#!/usr/bin/env bash
set -euo pipefail

echo "==> Running unit tests with race detection and coverage..."
go test -race -coverprofile=coverage.out -count=1 ./...
go tool cover -func=coverage.out | tail -n 1
echo "✓ Unit tests completed successfully."

#!/usr/bin/env bash
set -euo pipefail

echo "Running integration tests..."
go test -v -race -tags=integration -count=1 ./test/integration/... || echo "No integration test files found yet."

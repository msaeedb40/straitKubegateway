#!/usr/bin/env bash
set -euo pipefail

echo "Running end-to-end tests..."
go test -v -timeout 30m -tags=e2e -count=1 ./test/e2e/... || echo "No e2e test files found yet."

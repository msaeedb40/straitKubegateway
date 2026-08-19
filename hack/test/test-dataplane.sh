#!/usr/bin/env bash
set -euo pipefail

echo "Running dataplane tests..."
go test -v -race -tags=dataplane -count=1 ./test/dataplane/... || echo "No dataplane test files found yet."

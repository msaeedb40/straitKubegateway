#!/usr/bin/env bash
set -euo pipefail

echo "Running golint / static analysis verification..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
else
    echo "golangci-lint not installed, running go vet fallback"
    go vet ./...
fi
echo "Go linting check passed."

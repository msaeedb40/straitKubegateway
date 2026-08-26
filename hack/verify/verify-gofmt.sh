#!/usr/bin/env bash
set -euo pipefail

echo "==> Verifying gofmt..."
files=$(gofmt -s -l $(find . -type f -name '*.go' -not -path './vendor/*' -not -path './ui/*'))
if [[ -n "${files}" ]]; then
    echo "The following files are not gofmt'ed:"
    echo "${files}"
    exit 1
fi
echo "✓ gofmt verified."

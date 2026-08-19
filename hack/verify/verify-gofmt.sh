#!/usr/bin/env bash
set -euo pipefail

echo "==> Verifying gofmt..."
files=$(gofmt -l $(find . -name "*.go" -not -path "*/vendor/*" -not -path "*/ui/*"))
if [[ -n "${files}" ]]; then
  echo "The following files are not formatted with gofmt:"
  echo "${files}"
  exit 1
fi
echo "gofmt verification passed."

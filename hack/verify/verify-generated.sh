#!/usr/bin/env bash
set -euo pipefail

echo "==> Verifying generated code is up-to-date..."
git diff --exit-code api/ || {
  echo "Generated files are out of date. Please run 'make generate'."
  exit 1
}
echo "Generated code verification passed."

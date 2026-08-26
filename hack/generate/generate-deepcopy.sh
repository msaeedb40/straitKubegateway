#!/usr/bin/env bash
set -euo pipefail

echo "==> Generating deepcopy methods..."
if command -v controller-gen &> /dev/null; then
    controller-gen object paths="./api/..."
else
    echo "controller-gen not installed, skipping deepcopy regeneration."
fi
echo "✓ Deepcopy generation complete."

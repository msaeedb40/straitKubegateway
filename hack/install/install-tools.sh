#!/usr/bin/env bash
set -euo pipefail

echo "Installing development tools from tools/..."
if [ -d "tools" ]; then
    (cd tools && go mod tidy)
fi
echo "Tools setup complete."

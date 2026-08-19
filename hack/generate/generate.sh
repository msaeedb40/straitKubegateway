#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Running all code generators..."
bash "$SCRIPT_DIR/generate-deepcopy.sh"
bash "$SCRIPT_DIR/generate-crds.sh"
bash "$SCRIPT_DIR/generate-client.sh"
bash "$SCRIPT_DIR/generate-bpf.sh"
echo "All code generation completed successfully."

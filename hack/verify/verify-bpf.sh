#!/usr/bin/env bash
set -euo pipefail

echo "Verifying BPF source files..."
BPF_FILES=$(find bpf -name "*.c")
for f in $BPF_FILES; do
    if [ ! -s "$f" ]; then
        echo "Error: BPF source $f is empty!"
        exit 1
    fi
    echo "  Found BPF source: $f"
done
echo "BPF source verification passed."

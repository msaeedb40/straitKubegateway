#!/usr/bin/env bash
set -euo pipefail

CLANG=${CLANG:-clang}
BPF_CFLAGS="-O2 -g -Wall -Werror -target bpf"
BPF_INCLUDES="-I bpf/include"
BIN_DIR="bin/bpf"

echo "Building eBPF programs..."
mkdir -p "$BIN_DIR"

if command -v "$CLANG" &> /dev/null; then
    for src in $(find bpf -name "*.c"); do
        rel_path=${src#bpf/}
        obj="$BIN_DIR/${rel_path%.c}.o"
        mkdir -p "$(dirname "$obj")"
        echo "  CC $src -> $obj"
        $CLANG $BPF_CFLAGS $BPF_INCLUDES -c "$src" -o "$obj" || echo "  Warning: BPF compile failed (check kernel headers), continuing"
    done
else
    echo "clang not found in PATH, skipping BPF bytecode compilation"
fi
echo "BPF build process completed."

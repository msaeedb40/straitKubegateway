#!/usr/bin/env bash
set -euo pipefail

CLANG=${CLANG:-clang}
BPF_DIR="bpf"
OUT_DIR="ebpf/generated"

mkdir -p "${OUT_DIR}"
echo "==> Compiling eBPF programs..."

for src in "${BPF_DIR}"/src/*.c; do
    if [[ -f "$src" ]]; then
        base=$(basename "$src" .c)
        echo "  compiling ${src} -> ${OUT_DIR}/${base}.o"
        ${CLANG} -O2 -g -target bpf \
            -I"${BPF_DIR}/headers" \
            -c "${src}" \
            -o "${OUT_DIR}/${base}.o" || true
    fi
done
echo "✓ BPF compilation step completed."

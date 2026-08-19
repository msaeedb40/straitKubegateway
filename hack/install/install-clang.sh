#!/usr/bin/env bash
set -euo pipefail

echo "Checking Clang/LLVM toolchain..."
if command -v clang &> /dev/null; then
    echo "Clang version: $(clang --version | head -n 1)"
else
    echo "Clang is not installed. Please install llvm and clang packages."
fi

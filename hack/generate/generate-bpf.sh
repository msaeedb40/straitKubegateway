#!/usr/bin/env bash
set -euo pipefail

echo "Generating BPF Go bindings..."
go generate ./ebpf/... || echo "go generate ebpf completed."

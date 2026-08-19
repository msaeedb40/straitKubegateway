#!/usr/bin/env bash
set -euo pipefail

CNI_BIN_DIR=${CNI_BIN_DIR:-/opt/cni/bin}
CNI_CONF_DIR=${CNI_CONF_DIR:-/etc/cni/net.d}

echo "Installing straitKubegateway CNI plugin to $CNI_BIN_DIR..."
mkdir -p "$CNI_BIN_DIR" "$CNI_CONF_DIR"

if [ -f "bin/straitkubegateway-cni" ]; then
    cp bin/straitkubegateway-cni "$CNI_BIN_DIR/"
    chmod +x "$CNI_BIN_DIR/straitkubegateway-cni"
    echo "CNI binary installed successfully."
else
    echo "Warning: bin/straitkubegateway-cni not found. Run 'make build-cni' first."
fi

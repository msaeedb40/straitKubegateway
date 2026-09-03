#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — Upgrade Container Runtime on Kind Nodes
# ============================================================================
# Upgrades containerd (to 2.3.4) and runc (to 1.5.1) across all control plane
# and worker/agent nodes in the specified Kind cluster.
# ============================================================================

set -euo pipefail

CLUSTER_NAME="${1:-${CLUSTER_NAME:-ci-kind-test}}"
CONTAINERD_VERSION="${CONTAINERD_VERSION:-2.3.4}"
RUNC_VERSION="${RUNC_VERSION:-1.5.1}"

echo "============================================================"
echo " Upgrading Container Runtime on Kind Cluster: ${CLUSTER_NAME}"
echo " containerd target: v${CONTAINERD_VERSION}"
echo " runc target:       v${RUNC_VERSION}"
echo "============================================================"

# Discover all cluster nodes (control-plane and workers)
if command -v kind &>/dev/null; then
  NODES=$(kind get nodes --name "${CLUSTER_NAME}" 2>/dev/null || true)
fi

if [[ -z "${NODES:-}" ]]; then
  NODES=$(docker ps --filter "name=${CLUSTER_NAME}" --format "{{.Names}}")
fi

if [[ -z "${NODES:-}" ]]; then
  echo "Error: No nodes found for cluster '${CLUSTER_NAME}'" >&2
  exit 1
fi

for node in ${NODES}; do
  echo "==> Upgrading node: ${node} ..."

  docker exec "${node}" bash -c "
    set -euo pipefail
    ARCH=\$(uname -m)
    case \"\${ARCH}\" in
      x86_64)  ARCH='amd64' ;;
      aarch64) ARCH='arm64' ;;
      *) echo \"Unsupported architecture: \${ARCH}\" >&2; exit 1 ;;
    esac

    echo \"[${node}] Downloading containerd v${CONTAINERD_VERSION} (\${ARCH})...\"
    DOWNLOAD_CMD=''
    if command -v curl &>/dev/null; then
      curl -sSL \"https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-\${ARCH}.tar.gz\" | tar -xz -C /usr/local/
    elif command -v wget &>/dev/null; then
      wget -qO- \"https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-\${ARCH}.tar.gz\" | tar -xz -C /usr/local/
    else
      echo \"Neither curl nor wget available on ${node}\" >&2; exit 1
    fi

    echo \"[${node}] Downloading runc v${RUNC_VERSION} (\${ARCH})...\"
    RUNC_TARGET=\"\$(which runc || echo '/usr/local/sbin/runc')\"
    if command -v curl &>/dev/null; then
      curl -sSL -o \"\${RUNC_TARGET}\" \"https://github.com/opencontainers/runc/releases/download/v${RUNC_VERSION}/runc.\${ARCH}\"
    else
      wget -qO \"\${RUNC_TARGET}\" \"https://github.com/opencontainers/runc/releases/download/v${RUNC_VERSION}/runc.\${ARCH}\"
    fi
    chmod +x \"\${RUNC_TARGET}\"

    echo \"[${node}] Restarting containerd and kubelet...\"
    systemctl restart containerd
    systemctl restart kubelet
  "
  echo "==> Node ${node} upgrade finished."
done

echo ""
echo "============================================================"
echo " Runtime Verification Across Nodes"
echo "============================================================"
for node in ${NODES}; do
  echo "--- Node: ${node} ---"
  echo -n "containerd: "
  docker exec "${node}" containerd --version
  echo -n "runc:       "
  docker exec "${node}" runc --version
done
echo "============================================================"
echo "Container runtime upgrade complete!"

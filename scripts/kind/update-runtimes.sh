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

    CONTAINERD_BIN=\$(which containerd || echo '/usr/local/bin/containerd')
    if [[ \"\${CONTAINERD_BIN}\" != '/usr/local/bin/containerd' && -f /usr/local/bin/containerd ]]; then
      cp /usr/local/bin/containerd \"\${CONTAINERD_BIN}\"
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
echo " Configuring Cluster Network & Bootstrap API Server Routing"
echo "============================================================"
CONTROL_PLANE_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${CLUSTER_NAME}-control-plane" 2>/dev/null || true)
if [[ -z "${CONTROL_PLANE_IP}" ]]; then
  CONTROL_PLANE_IP=$(docker exec "${CLUSTER_NAME}-control-plane" hostname -I | awk '{print $1}' 2>/dev/null || true)
fi

WORKER_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${CLUSTER_NAME}-worker" 2>/dev/null || true)
WORKER2_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${CLUSTER_NAME}-worker2" 2>/dev/null || true)

for node in ${NODES}; do
  echo "==> Configuring networking on ${node} ..."
  docker exec "${node}" bash -c "
    sysctl -w net.ipv4.ip_forward=1 >/dev/null 2>&1 || true
    iptables -P FORWARD ACCEPT 2>/dev/null || true
    iptables -t nat -A POSTROUTING -s 10.244.0.0/16 -j MASQUERADE 2>/dev/null || true

    # DNAT kubernetes.default.svc (10.96.0.1:443) -> API server (${CONTROL_PLANE_IP}:6443)
    if [[ -n '${CONTROL_PLANE_IP}' ]]; then
      iptables -t nat -C PREROUTING -p tcp -d 10.96.0.1 --dport 443 -j DNAT --to-destination ${CONTROL_PLANE_IP}:6443 2>/dev/null || \
      iptables -t nat -A PREROUTING -p tcp -d 10.96.0.1 --dport 443 -j DNAT --to-destination ${CONTROL_PLANE_IP}:6443 2>/dev/null || true
      iptables -t nat -C OUTPUT -p tcp -d 10.96.0.1 --dport 443 -j DNAT --to-destination ${CONTROL_PLANE_IP}:6443 2>/dev/null || \
      iptables -t nat -A OUTPUT -p tcp -d 10.96.0.1 --dport 443 -j DNAT --to-destination ${CONTROL_PLANE_IP}:6443 2>/dev/null || true
    fi

    # Inter-node pod CIDR routes
    if [[ '${node}' != '${CLUSTER_NAME}-control-plane' && -n '${CONTROL_PLANE_IP}' ]]; then
      ip route replace 10.244.0.0/24 via ${CONTROL_PLANE_IP} 2>/dev/null || true
    fi
    if [[ '${node}' != '${CLUSTER_NAME}-worker' && -n '${WORKER_IP}' ]]; then
      ip route replace 10.244.1.0/24 via ${WORKER_IP} 2>/dev/null || true
    fi
    if [[ '${node}' != '${CLUSTER_NAME}-worker2' && -n '${WORKER2_IP}' ]]; then
      ip route replace 10.244.2.0/24 via ${WORKER2_IP} 2>/dev/null || true
    fi
  "
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
echo "Container runtime and network configuration complete!"

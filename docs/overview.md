# Overview of StraitKubeGateway

**StraitKubeGateway** is a high-performance, Kubernetes-native transit gateway and eBPF networking platform. It combines a production-grade CNI, an eBPF-powered data plane, kernel-native service load balancing, stateful firewall policies, multi-cluster transit networking, Gateway API v1.6.1 conformance, and BGP/BFD dynamic routing into a unified system.

Built entirely in Go (v1.26.5+) and modern Linux kernel eBPF technologies (NetKit, TCX, XDP, sockops, cgroup v2, LSM), StraitKubeGateway is designed to operate seamlessly across public clouds (AWS, GCP, Azure), bare-metal clusters, edge environments, hybrid clouds, and air-gapped deployments with zero cloud-provider lock-in.

---

## Key Features & Pillars

### 1. High-Performance eBPF Dataplane
- **NetKit & TCX Pod Networking**: Direct container-to-host eBPF device attachment via Linux NetKit for ultra-low latency and zero-copy packet processing.
- **XDP Fast Path**: Earliest packet filtering, DDoS mitigation, and acceleration at the network interface driver level.
- **Complete Kube-Proxy Replacement**: Enabled by default (`kubeProxyReplacement: true`, `kubeProxyMode: "none"`). Services and endpoints are compiled directly into eBPF maps, eliminating `iptables` or `ipvs` overhead.
- **Kernel-Native Load Balancing**: Maglev consistent hashing (127-slot lookup tables), round-robin, least-connections, IP-hash, and Direct Server Return (DSR).

### 2. Multi-Cluster Transit Gateway
- **Transit Segments**: 32-bit segment identifiers (`0` to `4294967295`) where Segment `0` acts as the backbone segment by default.
- **Flexible Topologies**: Hub-and-spoke, full mesh, peer-to-peer, and gateway-to-gateway peering.
- **Inter-Segment Routing**: Controlled communication between isolated segments via `TransitSegmentAttachment` CRDs.

### 3. Gateway API v1.6.1 Conformance
- Full support for Kubernetes Gateway API standard resources:
  - `GatewayClass` (Controller: `gateway.straitkubegateway.io/controller`)
  - `Gateway` (TCP, UDP, HTTP, HTTPS listeners)
  - `HTTPRoute` (Prefix, Exact, and Regex path matching, header and query parameter filtering, weighted backends)
  - `TCPRoute`, `UDPRoute`, `GRPCRoute`, `TLSRoute`

### 4. Zero-Hardcoding Dynamic IPAM & Dual-Stack
- **Dynamic CIDR Discovery**: Automatically discovers cluster PodCIDRs and ServiceCIDRs directly from Kubernetes Node specs and cluster ConfigMaps (`kubeadm-config`, `kube-proxy`).
- **IPv4 / IPv6 Dual-Stack**: Native dual-stack routing with RFC 1918/RFC 4193 subnets and NAT64 (RFC 6052 `64:ff9b::/96`) translation.
- **Broadcast Protection**: Automatic exclusion of IPv4 subnet broadcast and gateway addresses during pod IP allocation.

### 5. Advanced Policy & Stateful Security
- **Extended Selectors**: `StraitNetworkPolicy` supports matching by pod labels, namespaces, clusters, transit segments, gateways, and specific Gateway API routes.
- **Deterministic Compiler Rules**: Strict priority evaluation (0–255, lower = higher priority), deny-by-default ingress, allow-by-default egress, and deny overriding allow at equal priority.
- **WireGuard & IPsec Encryption**: Kernel-level pod-to-pod and node-to-node encryption using Curve25519 (X25519) and IPsec ESP.

### 6. Dynamic Routing & Fast Failover
- **BGP-4 Peering**: Dynamic route advertisement and RIB management.
- **BFD (Bidirectional Forwarding Detection)**: Sub-second peer link failure detection (RFC 5880) triggering instantaneous route withdrawal and traffic re-routing.

---

## Architectural Invariants

StraitKubeGateway enforces 15 strict architectural invariants to guarantee stability and prevent failure cascades:

1. **CNI bootstrap must not depend on the Service dataplane.**
2. **Kubernetes API connectivity must not depend on kube-proxy replacement.**
3. **CNI ADD must not synchronously depend on Service, Policy, NAT, Gateway, Transit, or BGP convergence.**
4. **Controllers produce desired network state (IR); they never manipulate BPF maps directly.**
5. **Dataplane IR is the sole boundary between the control plane and kernel.**
6. **The Dataplane Compiler is the single writer that translates IR into BPF maps and netlink state.**
7. **NetKit, TCX, and XDP are fast-path mechanisms.**
8. **cgroup v2, LSM, and sockops are control and security mechanisms.**
9. **kprobes, tracepoints, and perf/ring buffers are observability mechanisms, never normal forwarding mechanisms.**
10. **Cluster CIDR, service CIDR, pod CIDR, and EndpointSlice state are discovered dynamically.**
11. **CNI, service, policy, gateway, transit, and BGP readiness are independent conditions.**
12. **BPF map layouts are versioned contracts.**
13. **Every state transition is generation- and revision-based.**
14. **The dataplane continues forwarding packets even if the Kubernetes API server becomes temporarily unavailable.**
15. **Bootstrap API path (`straitd` → API :6443) uses direct node routing and never passes through Service LB.**

---

## Documentation Roadmap

- [Architecture & Diagrams](architecture.md): Deep dive into system components, dataflow, and kernel subsystems.
- [Lifecycle & Workflows](workflow.md): Detailed CNI ADD/DEL, packet path, and transit routing workflows with color-coded Mermaid diagrams.
- [Capabilities](capability.md): Comprehensive feature inventory and protocol support.
- [Security](security.md): Linux capabilities, RBAC, stateful firewall, and encryption architecture.
- [Observability](observability.md): Metadata model, Prometheus metrics, distributed tracing, and flow logs.
- [Operator Guide](guide.md): Installation, Helm deployment, setup, testing, and troubleshooting.
- [CLI Reference](cli.md): `sg-cli` commands, flags, and examples.

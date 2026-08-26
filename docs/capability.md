# Capabilities of StraitKubeGateway

This document provides a comprehensive feature matrix and capability catalog for **StraitKubeGateway**.

---

## 1. Feature Matrix Overview

| Functional Domain | Capability / Protocol | Implementation Mechanism | Status |
|---|---|---|---|
| **Core CNI** | CNI Spec 1.1+ (ADD, DEL, CHECK, GC, VERSION) | Go binary + Netlink | ✅ Supported |
| **Pod Networking** | Linux NetKit Link Pairs | `netkit` device mode | ✅ Supported |
| **Ingress Fast Path** | XDP (eXpress Data Path) | `bpf/src/xdp_ingress.c` | ✅ Supported |
| **Host Dataplane** | TCX / TC Ingress & Egress | `bpf/src/service_lb.c` | ✅ Supported |
| **Socket Acceleration**| Cgroup `connect4` / `sockops` | `SEC("cgroup/connect4")` | ✅ Supported |
| **Kube-Proxy Replacement** | Maglev Consistent Hash (127 slots) | eBPF Hash Maps | ✅ Supported |
| **Gateway API** | Gateway API v1.6.1 Specification | Standard CRDs & Controllers | ✅ Supported |
| **Multi-Cluster Transit**| Segments 0..4294967295 (Segment 0 Backbone)| `TransitGateway` CRD | ✅ Supported |
| **Dynamic Routing** | BGP-4 Route Peering & RIB Management | RFC 4271 engine | ✅ Supported |
| **Failure Detection** | BFD (Bidirectional Forwarding Detection) | RFC 5880 sub-second heartbeats | ✅ Supported |
| **Dynamic IPAM** | Cluster CIDR Discovery (Node/ConfigMap) | Zero hardcoded CIDRs | ✅ Supported |
| **Dual-Stack** | IPv4 & IPv6 Dual-Stack + NAT64 | Dual IPAM Pool & IPv6 translation | ✅ Supported |
| **Stateful Firewall** | `StraitNetworkPolicy` (Priority 0..255) | LRU Conntrack & Identity Map | ✅ Supported |
| **Encryption** | WireGuard (X25519) & IPsec ESP | Kernel crypto & Netlink | ✅ Supported |
| **Platform Control** | systemd watchdog & cgroup v2 resource control | `platform.Supervisor` | ✅ Supported |

---

## 2. Detailed Capability Breakdown

### A. Kube-Proxy Replacement & Load Balancing
- **Load Balancing Algorithms**:
  - `Maglev`: 127-entry permutation lookup table ensuring consistent connection hashing even during endpoint churn.
  - `RoundRobin`: Smooth cyclic distribution across healthy backends.
  - `LeastConnection`: Tracks active session count via `ct_map` and routes to the least loaded endpoint.
  - `IPHash`: Source IP hashing for sticky affinity.
- **Service Types Supported**: `ClusterIP`, `NodePort`, `LoadBalancer`, `ExternalName`, and `Headless` services.
- **Session Affinity**: ClientIP session affinity with configurable timeout seconds.
- **Direct Server Return (DSR)**: Preserves client source IP for backend pods without traversing reverse SNAT tunnels.

### B. Gateway API v1.6.1 Conformance
- **Resource Types Supported**:
  - `GatewayClass`: Custom gateway controller registration.
  - `Gateway`: Multi-port listeners supporting HTTP, HTTPS, TCP, UDP, TLS, and gRPC.
  - `HTTPRoute`: Path matching (`Exact`, `PathPrefix`, `RegularExpression`), header matching, query parameter filtering, and backend weighted routing.
  - `TCPRoute` & `UDPRoute`: High-performance L4 ingress stream forwarding.
  - `GRPCRoute`: Method and service level gRPC routing.
  - `TLSRoute`: SNI-based TLS passthrough.

### C. Multi-Cluster Transit Gateway
- **32-Bit Segmentation**: Full isolation across transit segments (`0` to `4294967295`), where Segment `0` serves as the universal transit backbone.
- **Topologies**:
  - `Hub-and-Spoke`: Central transit gateway connecting multiple isolated spoke clusters.
  - `Full Mesh`: Any-to-any direct peering between clusters.
  - `Peer-to-Peer`: Explicit point-to-point segment attachments via `TransitSegmentAttachment`.
- **Cross-Cluster Identity Preservation**: Security identities are preserved across transit tunnels, allowing policies to reference remote cluster pods.

### D. Dynamic Routing (BGP & BFD)
- **BGP-4 Engine**:
  - Autonomous System Number (ASN) configuration for eBGP and iBGP.
  - Dynamic PodCIDR and Service VIP advertisement to upstream data center routers.
  - Multi-hop BGP peering support.
- **BFD Engine (RFC 5880)**:
  - Sub-second link failure detection (e.g., 300ms transmit interval with 3x multiplier = 900ms failover).
  - Immediate trigger for BGP route withdrawal before TCP timeout occurs.

### E. Dynamic Zero-Hardcoding IPAM
- **Discovery Sources**:
  - Inspects `Node.Spec.PodCIDRs` and `Node.Spec.PodCIDR`.
  - Fallback discovery via `kube-system/kubeadm-config` (`podSubnet`, `serviceSubnet`) and `kube-system/kube-proxy` ConfigMaps.
- **Arbitrary Prefix Lengths**: Supports `/8` down to `/30` subnets.
- **Safety Filters**: Automatically excludes network base IP (`.0`), default gateway (`.1`), and broadcast IP (`.255` on `/24` or corresponding mask) from pod assignment.

### F. Traffic Encapsulation & Overlay Modes
- **Supported Encapsulation Protocols**:
  - `VXLAN` (UDP port 4789)
  - `Geneve` (UDP port 6081)
  - `GRE`
  - `WireGuard` (UDP port 51820)
  - `IPsec` ESP (IP Protocol 50)
- **Direct Routing Mode**: Native BGP routing without encapsulation when underlay MTU allows.

### G. Linux Platform Supervision
- **systemd Integration**: Native `sd_notify("READY=1")` and recurring `sd_notify("WATCHDOG=1")` ping preventing systemd service kills.
- **cgroup v2 Resource Enforcement**: Automated memory limits (`memory.max`), CPU quotas (`cpu.max`), and IO weight regulation on Linux cgroup v2 unified hierarchies.

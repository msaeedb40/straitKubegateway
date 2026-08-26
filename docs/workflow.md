# Workflow & Traffic Lifecycles

This document details the primary lifecycle and packet-forwarding workflows of **StraitKubeGateway**, complete with color-coded Mermaid diagrams and lifecycle step breakdowns.

---

## 1. CNI Pod Creation Workflow (CNI ADD)

The CNI ADD execution is divided into a **synchronous fast path** (must return within milliseconds) and **asynchronous dataplane reconciliation** (runs without blocking pod initialization).

```mermaid
sequenceDiagram
    autonumber
    actor Kubelet
    participant CNI as straitKubegateway CNI
    participant IPAM as Dynamic IPAM
    participant NetKit as NetKit Manager
    participant IdAlloc as Identity Allocator
    participant Compiler as Dataplane Compiler
    participant Kernel as Linux Kernel / eBPF

    rect rgb(225, 245, 254)
        Note over Kubelet,IdAlloc: Synchronous CNI ADD Fast-Path (Never blocks on Service / Policy)
        Kubelet->>CNI: CNI ADD (NetNS, ContainerID, IfName)
        CNI->>CNI: Validate NetworkConfig
        CNI->>IPAM: Allocate() [Dynamic Pool, Exclude Broadcast]
        IPAM-->>CNI: Pod IP & Default Gateway
        CNI->>NetKit: SetupNetKit(hostName, containerName, NetNS)
        NetKit->>Kernel: Create NetKit Link Pair (Host <-> Container)
        CNI->>Kernel: LockOSThread() & Setns() -> Assign IP/MTU & Default Route
        CNI->>IdAlloc: Allocate(DimensionLabels)
        IdAlloc-->>CNI: BPF Numeric Security Identity
        CNI-->>Kubelet: Return CNI AddResult (IP, Gateway, HostIfIndex, Identity)
    end

    rect rgb(232, 245, 233)
        Note over Compiler,Kernel: Asynchronous Dataplane Reconciliation Loop
        CNI-)Compiler: Queue Endpoint State Change
        Compiler->>Kernel: Program endpoint_map (IP -> Identity, IfIndex)
        Compiler->>Kernel: Compile Service LB & Policy Maps
    end
```

---

## 2. Comprehensive Traffic Flow & Transit Workflow

The following flowchart illustrates end-to-end packet processing across North-South Gateway ingress, East-West service load balancing, stateful policy evaluation, and Multi-Cluster Transit Gateway routing:

```mermaid
flowchart TD
    subgraph NS_Ingress["North-South Ingress (Gateway API)"]
        Client[External Client] -->|TCP / HTTP / gRPC| NIC[Physical NIC]
        NIC -->|Raw Packet| XDP[XDP Early Ingress Filter]
        XDP -->|Pass / Non-Blocked| TC_GW[TC Ingress: Gateway API Listener]
        TC_GW -->|Match HTTPRoute / Regex / Headers| GW_DNAT[Gateway DNAT to Backend Service]
    end

    subgraph EW_Service["East-West Service Load Balancing (Kube-Proxy Replacement)"]
        GW_DNAT --> LB_Lookup[Service Map Lookup: VIP + Port]
        PodClient[Pod Client] -->|ClusterIP / NodePort| Sockops[Cgroup Connect4 Hook]
        Sockops -->|Socket Translation| Maglev[Maglev Hash: 127 Slots]
        LB_Lookup --> Maglev
        Maglev -->|Select Backend| DNAT[DNAT Rewrite: ip->daddr & L4 Checksum]
    end

    subgraph Sec_Policy["Security & NetworkPolicy Engine"]
        DNAT --> Conntrack[Conntrack Table Lookup: ct_map]
        Conntrack -->|New Flow| PolicyEval[Policy Map Lookup: policy_map]
        Conntrack -->|Established Flow| FastFwd[Bypass Policy Engine]
        PolicyEval -->|Match Priority Rule 0..255| Verdict{Action?}
        Verdict -->|Deny / Reject| Drop[Drop Packet & Record Metric]
        Verdict -->|Allow| FastFwd
    end

    subgraph Transit_Routing["Multi-Cluster Transit & Forwarding"]
        FastFwd --> SegmentCheck{Destination Scope?}
        SegmentCheck -->|"Local Node Pod"| NetKitHost[NetKit Host Endpoint]
        SegmentCheck -->|"Remote Node (Same Cluster)"| Overlay[Overlay Tunnel: VXLAN / WireGuard]
        SegmentCheck -->|"Cross-Cluster (Segment 0..N)"| TransitGW[Transit Gateway Forwarding]
        
        TransitGW -->|Attachment Route Lookup| SegPeer[Peer Gateway Endpoint]
        SegPeer -->|Encrypted WireGuard / IPsec| RemoteCluster[Remote Cluster Ingress]
        NetKitHost -->|bpf_redirect| PodContainer[Destination Pod Network Namespace]
    end

    subgraph Failover_Observability["Observability & BFD Failover"]
        RemoteCluster -.->|BFD Heartbeat Check| BFD[BFD Manager]
        BFD -.->|Timeout Multiplier Exceeded| Withdraw[Instant BGP Route Withdrawal]
        Conntrack -.->|Flow Event| Observability[Observability: Prometheus / Trace]
    end

    classDef nsIngress fill:#e1f5fe,stroke:#0288d1,stroke-width:2px,color:#01579b;
    classDef ewService fill:#e8f5e9,stroke:#43a047,stroke-width:2px,color:#1b5e20;
    classDef secPolicy fill:#fff3e0,stroke:#fb8c00,stroke-width:2px,color:#e65100;
    classDef transit fill:#ede7f6,stroke:#7e57c2,stroke-width:2px,color:#311b92;
    classDef obsFailover fill:#fffde7,stroke:#fbc02d,stroke-width:2px,color:#f57f17;

    class NS_Ingress,Client,NIC,XDP,TC_GW,GW_DNAT nsIngress;
    class EW_Service,LB_Lookup,PodClient,Sockops,Maglev,DNAT ewService;
    class Sec_Policy,Conntrack,PolicyEval,Verdict,Drop,FastFwd secPolicy;
    class Transit_Routing,SegmentCheck,NetKitHost,Overlay,TransitGW,SegPeer,RemoteCluster,PodContainer transit;
    class Failover_Observability,BFD,Withdraw,Observability obsFailover;
```

---

## 3. Color Code Legend & Architecture Key

| Color Code | Subsystem Domain | Description & Responsibilities |
|---|---|---|
| 🔵 **Blue (`#e1f5fe`)** | **North-South & Ingress Control** | Gateway API v1.6.1 reconciliation, listener binding, HTTPRoute matching (PathPrefix, Exact, Regex), and initial packet ingress handling via XDP and TC hooks. |
| 🟢 **Green (`#e8f5e9`)** | **Fast-Path & Service LB** | High-performance packet forwarding, NetKit device attachments, socket-level `connect4` acceleration, and Maglev consistent hashing (127-slot lookup) with L3/L4 checksum recalculation. |
| 🟠 **Orange (`#fff3e0`)** | **Stateful Security & Policy** | BPF identity resolution, stateful connection tracking (`ct_map`), priority-based `StraitNetworkPolicy` enforcement (0–255, lower=higher), deny-by-default ingress, and stateful flow bypass. |
| 🟣 **Purple (`#ede7f6`)** | **Multi-Cluster Transit & Encapsulation** | Cross-cluster inter-segment routing (Segment 0 backbone), Hub-and-Spoke and Mesh overlays, WireGuard Curve25519 (X25519) encryption, and NetKit namespace delivery. |
| 🟡 **Yellow (`#fffde7`)** | **Observability & BFD Fast Failover** | Canonical 11-attribute metadata telemetry propagation, Prometheus metrics emission, and RFC 5880 BFD sub-second link failure detection with dynamic BGP route withdrawal. |

---

## 4. Pod Teardown Workflow (CNI DEL)

When a pod is deleted by Kubernetes, CNI DEL executes synchronous resource reclamation:

1. **Flush NetNS Routes**: Removes all kernel default and host routing table entries in the pod namespace.
2. **Remove BPF Identity**: De-allocates the numeric security identity from `identity.Allocator`.
3. **Purge Policy State**: Cleans up endpoint entries in `endpoint_map` and associated conntrack sessions in `ct_map`.
4. **Release IP to IPAM**: Releases the IPv4/IPv6 address back into the dynamic `ipam.Pool`, making it immediately available for new allocations.
5. **Destroy NetKit Link Pair**: Removes the host NetKit interface (`netlink.LinkDel`), closing kernel device handles.

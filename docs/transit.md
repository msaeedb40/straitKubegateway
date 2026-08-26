# Multi-Cluster Transit Gateway & Topologies

**StraitKubeGateway** includes a native, high-performance multi-cluster transit gateway that provides cross-cluster networking, 32-bit segment isolation, BGP-4 dynamic routing, sub-second BFD failover, and automated WireGuard/IPsec encryption.

---

## 1. Core Transit Networking Model

StraitKubeGateway uses a **segment-based network model** built upon 32-bit unsigned segment identifiers (`0` to `4,294,967,295`):

```mermaid
flowchart TD
    subgraph Backbone["Segment 0: Core Backbone Transit"]
        BB["Backbone Transit Gateway (Segment 0)"]
    end

    subgraph Segments["Isolated Application Segments"]
        Seg100["Segment 100<br/>(Production App A)"]
        Seg200["Segment 200<br/>(Payment Services)"]
        Seg300["Segment 300<br/>(Shared Analytics)"]
    end

    BB <-->|Encrypted Peering| Seg100
    BB <-->|Encrypted Peering| Seg200
    BB <-->|Encrypted Peering| Seg300

    classDef backbone fill:#ede7f6,stroke:#7e57c2,stroke-width:2px,color:#311b92;
    classDef segment fill:#e1f5fe,stroke:#0288d1,stroke-width:2px,color:#01579b;

    class Backbone,BB backbone;
    class Segments,Seg100,Seg200,Seg300 segment;
```

### Segment Rules & Invariants

1. **Segment IDs**: 32-bit unsigned integers (`uint32`) ranging from `0` to `4,294,967,295`.
2. **Backbone Segment `0`**: Segment `0` is reserved as the universal transit backbone.
3. **Default Zero-Trust Isolation**: All segments are isolated by default. Connecting to the backbone does not automatically grant inter-segment reachability.
4. **Explicit Routing Relationships**: Segment-to-segment communication requires explicit `TransitSegmentAttachment` or `StraitNetworkPolicy` bindings.
5. **Cross-Cluster Identity Preservation**: Workload numeric security identities are embedded in tunnel headers, preserving policy enforcement across cluster boundaries.

---

## 2. Supported Topologies

StraitKubeGateway supports four primary multi-cluster transit topologies:

### Topology 1: Hub-and-Spoke

In a Hub-and-Spoke topology, a central **Hub Gateway** (running on Segment `0`) acts as the transit router, security inspection engine, and shared service aggregator for multiple spoke clusters.

```mermaid
flowchart TB
    subgraph HubCluster["Hub Transit Cluster (Segment 0)"]
        HubGW["Hub Gateway Node<br/>• BGP-4 Router<br/>• WireGuard Terminator<br/>• Central Firewall & Inspection"]
        SharedSvc["Shared Services<br/>(Vault, DB, Logging)"]
        HubGW --- SharedSvc
    end

    subgraph SpokeA["Spoke Cluster A (Segment 100)"]
        GWA["Spoke Gateway A<br/>Node / Pods"]
    end

    subgraph SpokeB["Spoke Cluster B (Segment 200)"]
        GWB["Spoke Gateway B<br/>Node / Pods"]
    end

    subgraph SpokeC["Spoke Cluster C (Segment 300)"]
        GWC["Spoke Gateway C<br/>Node / Pods"]
    end

    GWA <-->|WireGuard Tunnel / BGP| HubGW
    GWB <-->|WireGuard Tunnel / BGP| HubGW
    GWC <-->|WireGuard Tunnel / BGP| HubGW

    classDef hub fill:#ede7f6,stroke:#7e57c2,stroke-width:2px,color:#311b92;
    classDef spoke fill:#e8f5e9,stroke:#43a047,stroke-width:2px,color:#1b5e20;

    class HubCluster,HubGW,SharedSvc hub;
    class SpokeA,SpokeB,SpokeC,GWA,GWB,GWC spoke;
```

#### Key Characteristics & Use Cases:
- **Centralized Security Inspection**: All inter-spoke traffic (`Spoke A → Hub → Spoke B`) traverses the Hub's eBPF policy engine.
- **Shared Enterprise Services**: Direct access to identity, vault, logging, and databases without duplicating services.
- **Controlled Egress**: Centralized internet breakout and NAT gateways.

---

### Topology 2: Full Mesh

In a Full Mesh topology, every cluster gateway establishes direct encrypted tunnels and BGP sessions with every other participating cluster gateway.

```mermaid
flowchart TD
    subgraph ClusterA["Cluster Alpha (Region US-East)"]
        GWA["Gateway Alpha<br/>Segment 10"]
    end

    subgraph ClusterB["Cluster Beta (Region US-West)"]
        GWB["Gateway Beta<br/>Segment 20"]
    end

    subgraph ClusterC["Cluster Gamma (Region EU-Central)"]
        GWC["Gateway Gamma<br/>Segment 30"]
    end

    GWA <-->|Direct Tunnel (Low Latency)| GWB
    GWB <-->|Direct Tunnel (Low Latency)| GWC
    GWC <-->|Direct Tunnel (Low Latency)| GWA

    classDef mesh fill:#fff3e0,stroke:#fb8c00,stroke-width:2px,color:#e65100;
    class ClusterA,ClusterB,ClusterC,GWA,GWB,GWC mesh;
```

#### Key Characteristics & Use Cases:
- **Lowest Latency (Single-Hop)**: Traffic flows directly between clusters without traversing an intermediate hub.
- **High Fault Tolerance**: Failure of one cluster or gateway does not impact communication between the remaining clusters.
- **High-Performance Microservices**: Ideal for active-active multi-region databases and globally distributed low-latency APIs.

---

### Topology 3: Peer-to-Peer (P2P)

Peer-to-Peer connectivity establishes direct point-to-point links between specific clusters or transit segments without requiring a full mesh or a permanent centralized hub.

```mermaid
flowchart LR
    subgraph PartnerCluster["Partner / Vendor Cluster"]
        GW_P["Partner Gateway<br/>Segment 500"]
    end

    subgraph PrimaryCluster["Primary Application Cluster"]
        GW_A["Primary Gateway<br/>Segment 100"]
    end

    subgraph DRCluster["Disaster Recovery Cluster"]
        GW_DR["DR Gateway<br/>Segment 100-Replica"]
    end

    GW_P <-->|Restricted P2P Tunnel<br/>(Specific Service CIDRs Only)| GW_A
    GW_A <-->|Replication P2P Tunnel<br/>(Continuous Sync)| GW_DR

    classDef p2p fill:#e1f5fe,stroke:#0288d1,stroke-width:2px,color:#01579b;
    class PartnerCluster,PrimaryCluster,DRCluster,GW_P,GW_A,GW_DR p2p;
```

#### Key Characteristics & Use Cases:
- **B2B & Partner Integration**: Secure cross-organization connectivity scoped strictly to required services.
- **Database Replication**: Dedicated high-throughput point-to-point replication links for disaster recovery.
- **Granular Access Control**: Least-privilege peering between two specific segment IDs.

---

### Topology 4: Gateway-to-Gateway (Multi-Cloud / Hybrid)

Gateway-to-Gateway topology connects on-premises data centers, edge nodes, and disparate cloud providers (AWS, GCP, Azure) through dedicated StraitKubeGateway border nodes.

```mermaid
flowchart TB
    subgraph CloudAWS["AWS EKS Cluster (VPC 10.100.0.0/16)"]
        GWEKS["AWS Transit Gateway Pod<br/>• XDP + WireGuard<br/>• BGP ASN 65001"]
    end

    subgraph CloudGCP["GCP GKE Cluster (VPC 10.200.0.0/16)"]
        GWGKE["GCP Transit Gateway Pod<br/>• XDP + WireGuard<br/>• BGP ASN 65002"]
    end

    subgraph OnPrem["On-Premises Bare-Metal (192.168.0.0/16)"]
        GWDC["On-Prem Gateway Node<br/>• Linux Kernel 6.12 LTS<br/>• BGP ASN 65000 + BFD"]
    end

    GWEKS <-->|Encrypted Cross-Cloud Tunnel| GWGKE
    GWEKS <-->|Direct Connect / IPSec VPN| GWDC
    GWGKE <-->|Cloud Interconnect / IPSec VPN| GWDC

    classDef cloud fill:#ede7f6,stroke:#7e57c2,stroke-width:2px,color:#311b92;
    classDef onprem fill:#e8f5e9,stroke:#43a047,stroke-width:2px,color:#1b5e20;

    class CloudAWS,CloudGCP,GWEKS,GWGKE cloud;
    class OnPrem,GWDC onprem;
```

#### Key Characteristics & Use Cases:
- **Hybrid Cloud Migration**: Seamless pod-to-pod routing between legacy bare-metal servers and cloud-native Kubernetes clusters.
- **Multi-Cloud Redundancy**: Avoid vendor lock-in with unified routing and policy enforcement across AWS, GCP, and Azure.
- **Fast Failover with BFD**: Link degradations are detected in under 900ms, triggering immediate BGP path shifts.

---

## 3. Transit CRD Reference

### `TransitGateway` CRD

Defines the local cluster's transit gateway identity, segment configuration, and routing parameters.

```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: TransitGateway
metadata:
  name: hub-transit-gw
  namespace: kube-system
spec:
  segmentID: 0
  topology: "hub-and-spoke" # hub-and-spoke | mesh | peer-to-peer | gateway-to-gateway
  listenPort: 51820
  bgp:
    asn: 65000
    routerID: "10.0.0.1"
    peers:
      - ip: "10.0.1.1"
        asn: 65001
        bfd:
          enabled: true
          interval: 300
          multiplier: 3
  encryption:
    type: "wireguard" # wireguard | ipsec
    wireguard:
      keepalive: 25
```

### `TransitSegmentAttachment` CRD

Authorizes routing and policy evaluation between two distinct transit segments.

```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: TransitSegmentAttachment
metadata:
  name: prod-to-shared-services
  namespace: kube-system
spec:
  sourceSegment: 100
  targetSegment: 300
  allowedCIDRs:
    - "10.244.10.0/24"
    - "10.96.0.0/12"
  policyMode: "enforce"
```

---

## 4. End-to-End Packet Walk (Cross-Cluster Transit)

```mermaid
sequenceDiagram
    autonumber
    actor PodA as Pod A (Cluster 1, Seg 100)
    participant NetKitA as NetKit Host (Node 1)
    participant TC_A as TC Egress (Node 1)
    participant TGW_1 as Transit Gateway 1
    participant Enc as WireGuard Crypto Engine
    participant TGW_2 as Transit Gateway 2
    participant TC_B as TC Ingress (Node 2)
    actor PodB as Pod B (Cluster 2, Seg 200)

    PodA->>NetKitA: Send packet to 10.244.20.15:8080
    NetKitA->>TC_A: TC Ingress lookup in fib_map
    TC_A->>TC_A: Identify destination in Segment 200 via Transit GW
    TC_A->>TGW_1: Forward packet to local Transit Gateway
    TGW_1->>TGW_1: Check TransitSegmentAttachment (100 -> 200) & Policy
    TGW_1->>Enc: Encapsulate & Encrypt (WireGuard X25519)
    Enc->>TGW_2: Transport over Underlay / Internet
    TGW_2->>TGW_2: Decrypt & Verify Security Identity
    TGW_2->>TC_B: Route to Destination Node
    TC_B->>PodB: Deliver via NetKit Container Link
```

---

## 5. Summary Matrix of Topologies

| Topology | Best For | Complexity | Latency | Resiliency |
|---|---|---|---|---|
| **Hub-and-Spoke** | Centralized security, shared enterprise services, controlled egress | Low | Medium (2 hops) | High (Redundant Hubs) |
| **Full Mesh** | Low-latency microservices, active-active multi-region | Medium | Lowest (1 hop) | Maximum (No single point of failure) |
| **Peer-to-Peer** | B2B partner links, DR database replication | Low | Lowest (1 hop) | High (Dedicated link) |
| **Gateway-to-Gateway** | Multi-cloud, hybrid bare-metal/cloud, edge computing | Medium | Low | High (BFD-assisted fast failover) |

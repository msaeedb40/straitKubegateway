# Custom Resource Definitions (CRD) Reference

straitKubegateway provides a suite of Custom Resource Definitions (CRDs) under the `straitkubegateway.io/v1alpha1` API group. This document provides complete field specifications, validation rules, default behaviors, and YAML examples for all resources.

---

## 📑 Index of Resources

1. [`Gateway`](#1-gateway) — Ingress / Transit Gateway listener definitions.
2. [`StraitNetworkPolicy`](#2-straitnetworkpolicy) — Hierarchical L3–L7 security policies.
3. [`TransitGateway`](#3-transitgateway) — Multi-cluster transit topology configuration.
4. [`TransitAttachment`](#4-transitattachment) — Attaches a cluster to a TransitGateway.
5. [`Segment`](#5-segment) — 32-bit isolated network segment definition.
6. [`TransitSegmentAttachment`](#6-transitsegmentattachment) — Inter-segment connection attachment.
7. [`TransitSegmentRoute`](#7-transitsegmentroute) — Cross-segment routing rule.
8. [`BGPPeer`](#8-bgppeer) — External BGP peering session configuration.
9. [`IPAMPool`](#9-ipampool) — Pod IP address allocation pool.
10. [`ClusterLink`](#10-clusterlink) — Multi-cluster federation link.

---

## 1. `Gateway`

Defines a Gateway instance with one or more network listeners bound to an isolated segment.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Namespaced`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `mode` | `string` | Yes | Operating mode: `standalone`, `hub`, `spoke`, `mesh`. |
| `segmentId` | `uint32` | No | Target segment ID (0–4294967295). Default `0`. |
| `listeners` | `[]GatewayListener` | Yes | List of port/protocol listeners. |
| `nodeSelector` | `map[string]string`| No | Selects nodes scheduled to host gateway datapath. |
| `encryption` | `*EncryptionConfig` | No | Transit encryption settings (`wireguard`, `ipsec`, `none`). |

#### `GatewayListener`
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | `string` | Yes | Unique name for the listener. |
| `protocol` | `string` | Yes | Protocol: `HTTP`, `HTTPS`, `TLS`, `TCP`, `UDP`, `gRPC`. |
| `port` | `int32` | Yes | Port number (1–65535). |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: Gateway
metadata:
  name: prod-gateway
  namespace: default
spec:
  mode: standalone
  segmentId: 100
  listeners:
    - name: http
      protocol: HTTP
      port: 80
    - name: https
      protocol: HTTPS
      port: 443
  encryption:
    type: wireguard
    keyRotationInterval: 86400
```

---

## 2. `StraitNetworkPolicy`

Enhanced security policy supporting multi-tier hierarchy (`Cluster` > `Segment` > `Namespace`), priority-based rule ordering, and Gateway API selectors.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster` or `Namespaced` (determined by `spec.scope`)

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `policyType` | `string` | Yes | `Ingress`, `Egress`, or `Both`. |
| `priority` | `uint32` | Yes | Evaluation priority (lower value = higher priority). |
| `defaultAction` | `string` | No | Fallback action: `Allow`, `Deny`, `Reject`. Default: `Deny` (Ingress), `Allow` (Egress). |
| `scope` | `string` | No | Authority level: `Cluster`, `Segment`, `Namespace`. Default: `Namespace`. |
| `ingress` | `[]PolicyRule` | No | Ingress rules evaluated in order of `ruleNo`. |
| `egress` | `[]PolicyRule` | No | Egress rules evaluated in order of `ruleNo`. |

#### `PolicyRule`
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `ruleNo` | `uint32` | Yes | 1-based incrementing rule index (`0` discards rule). |
| `action` | `string` | Yes | `Allow`, `Deny`, or `Reject`. |
| `from` | `[]PolicyPeer` | No | Source selectors (Ingress). |
| `to` | `[]PolicyPeer` | No | Destination selectors (Egress). |
| `ports` | `[]PolicyPort` | No | Protocols and port numbers/ranges. |
| `log` | `bool` | No | Enable real-time flow logging for matching traffic. |

#### `PolicyPeer` Selectors
- `podSelector`: Standard Kubernetes pod label selector.
- `namespaceSelector`: Namespace label selector.
- `clusterSelector`: Selects member clusters in multi-cluster topology.
- `segmentSelector`: Matches traffic by segment labels.
- `gatewaySelector`: Matches traffic from designated Gateway instances.
- `httprouteSelector`: Matches HTTPRoute resources.

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: StraitNetworkPolicy
metadata:
  name: cluster-security-baseline
spec:
  scope: Cluster
  policyType: Both
  priority: 10
  defaultAction: Deny
  ingress:
    - ruleNo: 1
      action: Allow
      from:
        - namespaceSelector:
            matchLabels:
              env: production
      ports:
        - protocol: TCP
          port: 443
      log: true
  egress:
    - ruleNo: 1
      action: Allow
      to:
        - podSelector:
            matchLabels:
              app: core-dns
      ports:
        - protocol: UDP
          port: 53
```

---

## 3. `TransitGateway`

Configures the multi-cluster transit routing topology.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `topology` | `string` | Yes | `hub-spoke`, `mesh`, or `peer-to-peer`. |
| `segmentId` | `uint32` | No | Backbone segment ID. Default `0`. |
| `tunnelType` | `string` | No | `geneve`, `vxlan`, `wireguard`, `gre`. Default `geneve`. |
| `encryption` | `*EncryptionConfig` | No | In-transit encryption settings. |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: TransitGateway
metadata:
  name: global-transit-hub
spec:
  topology: mesh
  segmentId: 0
  tunnelType: wireguard
  encryption:
    type: wireguard
    keyRotationInterval: 86400
```

---

## 4. `TransitAttachment`

Attaches an individual Kubernetes cluster or VPC network to a TransitGateway.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `transitGatewayRef` | `string` | Yes | References parent `TransitGateway` name. |
| `clusterId` | `string` | Yes | Unique ID of attaching cluster (e.g., `cluster-east-1`). |
| `segmentId` | `uint32` | Yes | Segment ID this cluster attaches into. |
| `podCidrs` | `[]string` | No | Pod CIDR blocks advertised by this cluster. |
| `serviceCidrs` | `[]string` | No | Service VIP CIDR blocks advertised. |
| `routes` | `[]TransitRoute` | No | Static routes advertised via this attachment. |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: TransitAttachment
metadata:
  name: attach-cluster-east
spec:
  transitGatewayRef: global-transit-hub
  clusterId: cluster-east-1
  segmentId: 100
  podCidrs:
    - "10.244.0.0/16"
  serviceCidrs:
    - "10.96.0.0/12"
```

---

## 5. `Segment`

Defines a 32-bit isolated network segment (VRF).

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | `uint32` | Yes | Unique 32-bit integer (0–4294967295). Segment 0 is backbone. |
| `isolated` | `bool` | No | If true, segment is strictly isolated from others. Default `true`. |
| `backboneConnected` | `bool` | No | If true, permits routing to Segment 0 backbone. Default `true`. |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: Segment
metadata:
  name: prod-segment-100
spec:
  id: 100
  isolated: true
  backboneConnected: true
```

---

## 6. `TransitSegmentAttachment`

Bridges two isolated segments together across the transit engine.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `sourceSegmentId` | `uint32` | Yes | Source segment ID. |
| `targetSegmentId` | `uint32` | Yes | Target segment ID. |
| `transitGatewayRef` | `string` | Yes | References parent TransitGateway. |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: TransitSegmentAttachment
metadata:
  name: bridge-100-to-300
spec:
  sourceSegmentId: 100
  targetSegmentId: 300
  transitGatewayRef: global-transit-hub
```

---

## 7. `TransitSegmentRoute`

Defines an inter-segment routing rule directing traffic for a CIDR destination to a specific attachment.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `cidr` | `string` | Yes | Destination CIDR prefix (e.g. `10.250.0.0/16` or `0.0.0.0/0`). |
| `nextHopAttachment` | `string` | Yes | Name of target `TransitAttachment` or `TransitSegmentAttachment`. |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: TransitSegmentRoute
metadata:
  name: route-to-cluster-west
spec:
  cidr: "10.250.0.0/16"
  nextHopAttachment: attach-cluster-west
```

---

## 8. `BGPPeer`

Configures external BGP peering sessions with Top-of-Rack (ToR) routers or cloud gateways.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `peerAsn` | `uint32` | Yes | Remote autonomous system number. |
| `localAsn` | `uint32` | Yes | Local autonomous system number. |
| `peerAddress` | `string` | Yes | Remote peer IP address. |
| `localAddress` | `string` | No | Local IP address to establish peering session from. |
| `holdTime` | `int32` | No | BGP hold time in seconds. Default `90`. |
| `keepaliveInterval` | `int32` | No | Keepalive interval in seconds. Default `30`. |
| `bfdEnabled` | `bool` | No | Enables sub-second Bidirectional Forwarding Detection. |
| `advertisedPrefixes`| `[]string` | No | List of CIDR prefixes advertised to this peer. |
| `routeFilters` | `[]BGPRouteFilter` | No | Inbound/outbound route filtering rules. |
| `nodeSelector` | `map[string]string`| No | Nodes that should establish this peering session. |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: BGPPeer
metadata:
  name: tor-switch-1
spec:
  peerAsn: 65001
  localAsn: 64512
  peerAddress: "192.168.1.1"
  bfdEnabled: true
  advertisedPrefixes:
    - "10.244.0.0/16"
    - "10.96.0.0/12"
  routeFilters:
    - prefix: "0.0.0.0/0"
      action: accept
      matchType: exact
```

---

## 9. `IPAMPool`

Configures cluster IP address pools and per-node subnet slicing.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `cidrs` | `[]string` | Yes | List of CIDR pools available for allocation. |
| `perNodeMaskSize` | `int32` | Yes | Prefix length allocated to each node (8–30). E.g., `24` for `/24` per node. |
| `addressFamily` | `string` | Yes | `IPv4` or `IPv6`. |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: IPAMPool
metadata:
  name: default-ipv4-pool
spec:
  addressFamily: IPv4
  cidrs:
    - "10.244.0.0/16"
  perNodeMaskSize: 24
```

---

## 10. `ClusterLink`

Defines a multi-cluster federation link for cross-cluster service synchronization.

- **Group:** `straitkubegateway.io`
- **Version:** `v1alpha1`
- **Scope:** `Cluster`

### Spec Fields
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `clusterId` | `string` | Yes | Unique identifier of remote cluster. |
| `apiEndpoint` | `string` | Yes | Kubernetes API endpoint of remote cluster. |
| `podCidrs` | `[]string` | No | Pod CIDRs belonging to remote cluster. |
| `serviceCidrs` | `[]string` | No | Service CIDRs belonging to remote cluster. |
| `secretRef` | `string` | No | Name of Secret containing remote cluster kubeconfig. |

### Example Manifest
```yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: ClusterLink
metadata:
  name: link-cluster-west
spec:
  clusterId: cluster-west-1
  apiEndpoint: "https://10.0.2.10:6443"
  podCidrs:
    - "10.246.0.0/16"
  serviceCidrs:
    - "10.98.0.0/12"
  secretRef: cluster-west-kubeconfig
```

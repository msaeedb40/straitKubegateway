# sg-cli Command Line Reference

`sg-cli` is the official administrative and diagnostics command-line utility for **straitKubegateway**. It allows platform engineers and cluster administrators to inspect the eBPF dataplane, verify CNI readiness, simulate network policies, dump kernel BPF maps, manage multi-cluster transit attachments, and launch the web UI.

---

## 📑 Global Flags

Every `sg-cli` command inherits the following persistent flags:

| Flag | Shorthand | Type | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--config` | | `string` | `$HOME/.straitkubegateway.yaml` | Path to custom CLI configuration file. |
| `--namespace` | `-n` | `string` | `default` | Target Kubernetes namespace for resource operations. |
| `--output` | `-o` | `string` | `text` | Output format. Supported: `text`, `json`, `yaml`. |

---

## 🛠 Command Reference

### 1. `sg-cli status`
Displays the real-time operational status of all straitKubegateway components across the cluster.

```bash
sg-cli status
```

**Example Output:**
```
straitKubegateway Status:
  CNI:           Ready (NetKit eBPF)
  Control Plane: Running (sg-controller)
  Node Agent:    Active (straitd)
  Dataplane:     Linux Kernel 6.7+ eBPF (NetKit + TCX + XDP)
  Kube-Proxy:    Replacement Active (eBPF Service LB)
  BGP Speaker:   Running
  Transit Hub:   Segment 0 Backbone Connected
```

---

### 2. `sg-cli node`
Inspects `straitd` node agents and dumps Linux kernel eBPF state.

#### `sg-cli node list`
Lists all cluster nodes running `straitd`, their Pod CIDRs, and kernel version.
```bash
sg-cli node list
```
**Output:**
```
NODE            STATUS    POD CIDR          DATAPLANE         KERNEL
node-1          Ready     10.244.0.0/24     NetKit+TCX+XDP    6.8.0-generic
node-2          Ready     10.244.1.0/24     NetKit+TCX+XDP    6.8.0-generic
node-3          Ready     10.244.2.0/24     NetKit+TCX+XDP    6.8.0-generic
```

#### `sg-cli node bpf-maps`
Dumps loaded eBPF maps, map types, and active entry counts on the local node.
```bash
sg-cli node bpf-maps
```
**Output:**
```
ID      NAME              TYPE      ENTRIES    MAX_ENTRIES
1       endpoints_map     hash      12         65536
2       services_map      hash      8          16384
3       policies_map      hash      24         65536
4       conntrack_map     lru       156        131072
5       snat_map          hash      42         65536
6       segment_map       hash      3          4096
```

---

### 3. `sg-cli endpoint`
Manages and inspects active pod network endpoints provisioned by the NetKit CNI.

#### `sg-cli endpoint list`
Lists all active pod endpoints on the node or cluster.
```bash
sg-cli endpoint list -n default
```
**Output:**
```
CONTAINER ID    NAMESPACE    POD             IP             IFINDEX    SEGMENT    STATE
c8a1f4b209d1    default      frontend-7b9    10.244.0.15    12         0          Ready
d9b2e5c310a2    default      backend-9c4     10.244.0.16    13         100        Ready
```

#### `sg-cli endpoint get <container-id>`
Displays detailed information for a specific container endpoint.
```bash
sg-cli endpoint get c8a1f4b209d1
```
**Output:**
```
Endpoint Details for container c8a1f4b209d1:
  State:    Ready
  Datapath: NetKit (host-veth: sg-c8a1f4b2)
  IP:       10.244.0.15/24
  Gateway:  10.244.0.1
  Identity: 1001
  Segment:  0 (Backbone)
```

---

### 4. `sg-cli gateway`
Inspects and manages Gateway API resources and active listener instances.

#### `sg-cli gateway list`
Lists active Gateway API instances and their programmed addresses.
```bash
sg-cli gateway list -n default
```
**Output:**
```
NAMESPACE    NAME       CLASS     ADDRESS       PROGRAMMED
default      main-gw    strait    10.96.0.1     True
prod         edge-gw    strait    192.168.1.50  True
```

#### `sg-cli gateway get <name>`
Retrieves listeners and route bindings for a gateway.
```bash
sg-cli gateway get main-gw -n default
```

---

### 5. `sg-cli policy`
Inspects active `StraitNetworkPolicy` instances and runs simulated packet verdicts.

#### `sg-cli policy list`
Lists compiled network policies, scope, priority, and default actions.
```bash
sg-cli policy list -n default
```
**Output:**
```
NAMESPACE    NAME              TYPE       PRIORITY    DEFAULT ACTION    STATUS
default      allow-frontend    Ingress    10          Deny              Enforced
default      egress-strict     Egress     20          Allow             Enforced
```

#### `sg-cli policy simulate`
Simulates a policy evaluation between a source pod and destination endpoint without sending traffic.
```bash
sg-cli policy simulate
```
**Output:**
```
Simulating flow:
  Source:      default/pod-a (Identity: 1001)
  Destination: default/pod-b (Identity: 1002, Port: 8080/TCP)
  Verdict:     ALLOW (Matched rule #1 in allow-frontend)
```

---

### 6. `sg-cli transit`
Inspects multi-cluster transit topologies, segments, and route tables.

#### `sg-cli transit gateways`
Lists multi-cluster transit gateways.
```bash
sg-cli transit gateways
```
**Output:**
```
NAME          TOPOLOGY      BACKBONE SEGMENT    ATTACHED CLUSTERS    STATUS
global-tgw    hub-spoke     0                   3                    Ready
```

#### `sg-cli transit segments`
Lists network segments, isolation state, and attached endpoint counts.
```bash
sg-cli transit segments
```
**Output:**
```
SEGMENT ID    NAME          ISOLATED    BACKBONE CONNECTED    ENDPOINTS
0             backbone      false       true                  48
100           prod-vpc      true        true                  16
200           dev-vpc       true        false                 8
```

---

### 7. `sg-cli bgp`
Monitors BGP peering sessions and advertised routing prefixes.

#### `sg-cli bgp peers`
Displays remote BGP neighbor sessions and uptime.
```bash
sg-cli bgp peers
```
**Output:**
```
PEER IP        REMOTE ASN    LOCAL ASN    STATE          UPTIME
192.168.1.1    65001         64512        Established    14h32m
192.168.1.2    65002         64512        Established    14h30m
```

#### `sg-cli bgp routes`
Lists prefixes advertised to or received from BGP neighbors.
```bash
sg-cli bgp routes
```
**Output:**
```
PREFIX           NEXT HOP         STATUS        COMMUNITY
10.244.0.0/16    192.168.1.100    Advertised    64512:100
10.96.0.0/12     192.168.1.100    Advertised    64512:200
```

---

### 8. `sg-cli cluster`
Manages multi-cluster federation links.

#### `sg-cli cluster list`
Lists all connected remote Kubernetes clusters and heartbeat health.
```bash
sg-cli cluster list
```
**Output:**
```
CLUSTER ID      ENDPOINT                   POD CIDRS        CONNECTED    LAST HEARTBEAT
cluster-east    https://10.0.1.10:6443     10.245.0.0/16    True         5s ago
cluster-west    https://10.0.2.10:6443     10.246.0.0/16    True         2s ago
```

---

### 9. `sg-cli wireguard` & `sg-cli ipsec`
Inspects transit encryption tunnels.

#### `sg-cli wireguard status`
```bash
sg-cli wireguard status
```
**Output:**
```
INTERFACE    PUBLIC KEY                                LISTEN PORT    PEERS
sg-wg0       Wk5b9...K8Za=                             51820          2
```

#### `sg-cli ipsec status`
```bash
sg-cli ipsec status
```
**Output:**
```
TUNNEL ID    REMOTE ENDPOINT    SPI (IN/OUT)             CIPHER         STATUS
ipsec-1      198.51.100.2       0xc12a4b / 0xd88e1a      AES-GCM-256    Established
```

---

### 10. `sg-cli config`
Displays active cluster networking configuration.

```bash
sg-cli config view
```
**Output:**
```yaml
cluster:
  name: primary-cluster
  podCIDRs:
    - 10.244.0.0/16
  serviceCIDRs:
    - 10.96.0.0/12
agent:
  kubeProxyReplacement: true
  directServerReturn: true
  maglevLookupTableSize: 128
```

---

### 11. `sg-cli ui`
Port-forwards the Angular 22 dashboard for local browser access.

```bash
sg-cli ui
```
**Output:**
```
Connecting to straitKubegateway UI service...
Forwarding local port 4200 -> svc/straitKubegateway-ui:80
Dashboard available at: http://localhost:4200
```

---

### 12. `sg-cli export` & `sg-cli import`
Backs up or restores cluster networking CRD resources.

```bash
# Export all straitKubegateway CRD manifests to file
sg-cli export -o yaml > strait-backup.yaml

# Apply configuration from backup
sg-cli import -f strait-backup.yaml
```

---

### 13. `sg-cli install` & `sg-cli upgrade`
Automates cluster deployment and rolling upgrades via the Helm chart.

```bash
# Install onto current kubeconfig context
sg-cli install

# Perform zero-downtime rolling upgrade
sg-cli upgrade --version v1.1.0
```

---

### 14. `sg-cli version`
Prints version and build commit info.

```bash
sg-cli version
```
**Output:**
```
straitKubegateway sg-cli v1.0.0 (commit: a1b2c3d, built: 2026-08-19T00:00:00Z)
```

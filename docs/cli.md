# sg-cli: StraitKubeGateway CLI Reference

`sg-cli` is the command-line management tool for inspecting, configuring, and troubleshooting the StraitKubeGateway platform.

---

## Global Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--config` | — | `$HOME/.sg-cli.yaml` | Path to sg-cli configuration file |
| `--verbose` | `-v` | `false` | Enable verbose output |

---

## Commands

### `sg-cli status`

Show the current status of all StraitKubeGateway components.

```bash
sg-cli status
```

**Output:**
```
straitKubegateway Cluster Status:
  CNI Dataplane:      Ready
  Service LB:         Ready (kube-proxy replacement: true)
  NetworkPolicy:      Active (LSM/cgroup hooks)
  Gateway API v1.6.1: Active
  Transit Gateway:    Segment 0 (Backbone)
  WireGuard:          Operational
```

---

### `sg-cli node`

Manage and inspect node networking status.

```bash
# Show all nodes and their networking state
sg-cli node

# Show detailed info for a specific node
sg-cli node --name worker-01
```

---

### `sg-cli endpoint`

Inspect pod network endpoints and allocated BPF identities.

```bash
# List all endpoints
sg-cli endpoint

# Filter by namespace
sg-cli endpoint --namespace production

# Show endpoint details for a specific pod
sg-cli endpoint --pod orders-api-7b8f6d-xz9k2 --namespace production
```

---

### `sg-cli policy`

Manage and debug StraitKubeGateway network policies.

```bash
# List all compiled policies
sg-cli policy

# Show policy details
sg-cli policy --name web-policy --namespace default

# Trace policy evaluation for a specific flow
sg-cli policy trace --src-pod frontend-abc --dst-pod backend-xyz \
  --port 8080 --protocol TCP
```

---

### `sg-cli gateway`

Inspect and manage Gateway API resources.

```bash
# List all Gateways
sg-cli gateway

# Show Gateway details and attached routes
sg-cli gateway --name strait-gw --namespace default

# Test route matching
sg-cli gateway match --gateway strait-gw --path "/api/v1/orders" --method GET
```

---

### `sg-cli transit`

Manage multi-cluster transit gateway segments and routes.

```bash
# Show transit gateway status
sg-cli transit

# List peers in a segment
sg-cli transit --segment 0

# Show segment attachments
sg-cli transit attachments
```

---

### `sg-cli bgp`

Inspect BGP neighbors and route advertisements.

```bash
# Show BGP neighbor table
sg-cli bgp

# Show advertised prefixes
sg-cli bgp advertised-routes

# Show received routes from a peer
sg-cli bgp received-routes --peer 10.0.0.1
```

---

### `sg-cli wireguard`

Inspect WireGuard tunnels and encryption state.

```bash
# Show WireGuard interface status
sg-cli wireguard

# Show peer public keys and handshake times
sg-cli wireguard peers
```

---

### `sg-cli ipsec`

Inspect IPsec Security Associations and XFRM states.

```bash
# Show IPsec SA database
sg-cli ipsec

# Show XFRM policies
sg-cli ipsec policies
```

---

### `sg-cli cluster`

Manage federated Kubernetes clusters and peering.

```bash
# List federated clusters
sg-cli cluster

# Show cluster peering status
sg-cli cluster --name remote-west
```

---

### `sg-cli config`

View and modify straitKubegateway runtime configuration.

```bash
# Show current runtime config
sg-cli config

# Set a configuration value
sg-cli config set networking.tunnelMode geneve
```

---

### `sg-cli export`

Export straitKubegateway configuration and policies.

```bash
# Export all configuration to YAML
sg-cli export > straitkubegateway-backup.yaml

# Export policies only
sg-cli export --policies-only > policies-backup.yaml
```

---

### `sg-cli import [file]`

Import straitKubegateway configuration from a file.

```bash
sg-cli import straitkubegateway-backup.yaml
```

---

### `sg-cli install`

Install straitKubegateway in the current Kubernetes cluster.

```bash
sg-cli install --namespace kube-system --set wireguard.enabled=true
```

---

### `sg-cli upgrade`

Upgrade straitKubegateway components to the latest version.

```bash
sg-cli upgrade --namespace kube-system
```

---

### `sg-cli version`

Print the straitKubegateway CLI version.

```bash
sg-cli version
```

**Output:**
```
sg-cli version: v1.0.0
  Go:        go1.26.7
  OS/Arch:   linux/amd64
  Built:     2026-08-26T10:00:00Z
```

---

## Command Summary

| Command | Description |
|---|---|
| `sg-cli status` | Show component health and readiness |
| `sg-cli node` | Inspect node networking state |
| `sg-cli endpoint` | List pod endpoints and BPF identities |
| `sg-cli policy` | Manage and trace network policies |
| `sg-cli gateway` | Inspect Gateway API resources and route matching |
| `sg-cli transit` | Manage multi-cluster transit segments |
| `sg-cli bgp` | Inspect BGP neighbors and routes |
| `sg-cli wireguard` | Inspect WireGuard tunnel state |
| `sg-cli ipsec` | Inspect IPsec SAs and XFRM policies |
| `sg-cli cluster` | Manage federated cluster peering |
| `sg-cli config` | View/modify runtime configuration |
| `sg-cli export` | Export configuration to YAML |
| `sg-cli import` | Import configuration from file |
| `sg-cli install` | Install StraitKubeGateway |
| `sg-cli upgrade` | Upgrade StraitKubeGateway |
| `sg-cli version` | Print CLI version information |

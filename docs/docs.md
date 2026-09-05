# Documentation Index for StraitKubeGateway

The complete documentation suite for **StraitKubeGateway** is structured into the following modules with Mermaid diagram support and standardized color coding:

### Interactive Hub & Live Application

- **[Interactive Web Hub & Helm Portal](https://msaeedb40.github.io/straitKubegateway)**: Live interactive web application featuring dynamic Helm configurator, platform guides (Kind, K3s, Minikube, Kubeadm), topological visualizers, eBPF hook explorer, policy simulator, `sg-cli` terminal playground, and Prometheus/Grafana dashboard.

### Core Documentation Modules

1. [`overview.md`](overview.md): High-level overview of StraitKubeGateway, core feature pillars, and architectural invariants.
2. [`architecture.md`](architecture.md): System architecture, control plane reconcilers, node agent (`straitd`), and eBPF dataplane with Mermaid architecture diagrams.
3. [`workflow.md`](workflow.md): CNI ADD/DEL lifecycles and comprehensive packet processing flows with color-coded Mermaid diagrams and architecture keys.
4. [`transit.md`](transit.md): Multi-cluster transit gateway architecture, 32-bit segment isolation, and 4 primary topologies (**Hub-and-Spoke**, **Full Mesh**, **Peer-to-Peer**, **Gateway-to-Gateway**) with CRD specifications and packet walk flows.
5. [`capability.md`](capability.md): Comprehensive feature matrix, protocol support, algorithms, and dynamic IPAM capabilities.
6. [`security.md`](security.md): Linux capability boundary model, Kubernetes RBAC, stateful firewall compiler, and WireGuard/IPsec encryption architecture.
7. [`observability.md`](observability.md): Canonical 11-attribute telemetry metadata model, Prometheus metrics families, pinned `kube-prometheus-stack` (88.6.4), Grafana dashboards, zap logging, and OpenTelemetry distributed tracing pipeline.
8. [`guide.md`](guide.md): Comprehensive operator guide covering prerequisites, Helm deployment, automated `sg-cli` installation, platform deployments (Kind, K3s, Minikube, Kubeadm bare-metal), Prometheus/Grafana setup, post-install verification, upgrade, uninstall, testing, and troubleshooting.
9. [`cli.md`](cli.md): Full command-line reference, installation instructions (amd64/arm64), and interactive command examples for `sg-cli`.
10. [`README.md`](../README.md): Project root readme with logo, badges, quick start, platform matrix, and directory structure.

# Getting Started with straitKubegateway

This step-by-step guide walks you through installing, configuring, and verifying **straitKubegateway** on a Kubernetes cluster.

---

## 📋 Prerequisites

Before installing straitKubegateway, ensure your environment meets the following requirements:

### 1. Host Kernel & OS
- **Linux Kernel**: 6.7 or newer (Recommended: Linux 6.12 LTS+).
- **Required Kernel Modules & Configs**:
  - `CONFIG_BPF=y`
  - `CONFIG_BPF_SYSCALL=y`
  - `CONFIG_NETKIT=y`
  - `CONFIG_CGROUP_BPF=y`
  - `CONFIG_NET_CLS_ACT=y`
- **BPF Filesystem**: Mounted at `/sys/fs/bpf` (`mount -t bpf bpffs /sys/fs/bpf`).
- **Cgroup Mode**: Cgroup v2 (`/sys/fs/cgroup`).

### 2. Kubernetes Cluster
- **Kubernetes Version**: v1.30 or newer.
- **CNI State**: Clean cluster without conflicting CNI plugins (or configured for CNI chaining/replacement).
- **kube-proxy**: Can be disabled or set to `mode: none` when using full eBPF kube-proxy replacement.

### 3. Client Tools
- `kubectl` configured with cluster admin access.
- `helm` v3.12+ or v4.x.
- `sg-cli` installed locally.

---

## 🚀 Installation

### Step 1: Clone the Repository & Prepare Helm
```bash
git clone https://github.com/straitKubegateway/straitKubegateway.git
cd straitKubegateway
```

### Step 2: Install via Helm
Install straitKubegateway into the `strait-system` namespace:

```bash
helm install straitkubegateway ./straitKubegateway-helm \
  --namespace strait-system \
  --create-namespace \
  --set agent.kubeProxyReplacement=true \
  --set agent.directServerReturn=true \
  --set transit.enabled=true \
  --set ui.enabled=true
```

### Step 3: Verify Deployment
Check that all pods in `strait-system` reach the `Running` state:

```bash
kubectl get pods -n strait-system -o wide
```

Output:
```
NAME                             READY   STATUS    RESTARTS   AGE
sg-controller-74f8c85c5-x9p2m    1/1     Running   0          45s
sg-controller-74f8c85c5-z8k4b    1/1     Running   0          45s
straitd-node-1-w7x2q             1/1     Running   0          45s
straitd-node-2-b4y8m             1/1     Running   0          45s
straitKubegateway-ui-84d7...     1/1     Running   0          45s
```

Verify operational readiness with `sg-cli`:
```bash
sg-cli status
```

---

## 🛠 Basic Usage Walkthrough

### 1. Deploy a Sample Workload
Deploy a simple backend application:

```yaml
# sample-app.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-backend-v1
  namespace: default
  labels:
    app: demo-backend
    version: v1
spec:
  replicas: 2
  selector:
    matchLabels:
      app: demo-backend
      version: v1
  template:
    metadata:
      labels:
        app: demo-backend
        version: v1
    spec:
      containers:
        - name: app
          image: hashicorp/http-echo:latest
          args: ["-text=Hello from straitKubegateway v1!"]
          ports:
            - containerPort: 5678
---
apiVersion: v1
kind: Service
metadata:
  name: demo-backend
  namespace: default
spec:
  selector:
    app: demo-backend
  ports:
    - protocol: TCP
      port: 80
      targetPort: 5678
```

Apply the workload:
```bash
kubectl apply -f sample-app.yaml
```

Inspect the allocated NetKit pod endpoints:
```bash
sg-cli endpoint list -n default
```

---

### 2. Configure Gateway API & HTTP Routing
Deploy a `Gateway` and route traffic to the backend service:

```yaml
# gateway-demo.yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: Gateway
metadata:
  name: public-gateway
  namespace: default
spec:
  mode: standalone
  segmentId: 0
  listeners:
    - name: http
      protocol: HTTP
      port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend-route
  namespace: default
spec:
  parentRefs:
    - name: public-gateway
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /api/v1
      backendRefs:
        - name: demo-backend
          port: 80
          weight: 100
```

Apply the gateway configuration:
```bash
kubectl apply -f gateway-demo.yaml
```

Check the gateway status:
```bash
sg-cli gateway list
```

---

### 3. Enforce a Hierarchical StraitNetworkPolicy
Enforce zero-trust ingress security with a cluster-wide policy:

```yaml
# policy-demo.yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: StraitNetworkPolicy
metadata:
  name: secure-backend-ingress
  namespace: default
spec:
  scope: Namespace
  policyType: Ingress
  priority: 10
  defaultAction: Deny
  ingress:
    - ruleNo: 1
      action: Allow
      from:
        - gatewaySelector:
            matchLabels:
              gateway.networking.k8s.io/gateway-name: public-gateway
      ports:
        - protocol: TCP
          port: 5678
      log: true
```

Apply and simulate the policy:
```bash
kubectl apply -f policy-demo.yaml

# Test policy verdict without sending packets
sg-cli policy simulate
```

---

### 4. Configure Multi-Cluster Transit & Segmentation
Create an isolated segment (`Segment 100`) and establish a transit attachment:

```yaml
# transit-segment-demo.yaml
apiVersion: straitkubegateway.io/v1alpha1
kind: Segment
metadata:
  name: prod-segment
spec:
  id: 100
  isolated: true
  backboneConnected: true
---
apiVersion: straitkubegateway.io/v1alpha1
kind: TransitGateway
metadata:
  name: primary-tgw
spec:
  topology: mesh
  segmentId: 0
  tunnelType: wireguard
  encryption:
    type: wireguard
    keyRotationInterval: 86400
---
apiVersion: straitkubegateway.io/v1alpha1
kind: TransitAttachment
metadata:
  name: local-cluster-attachment
spec:
  transitGatewayRef: primary-tgw
  clusterId: primary-cluster
  segmentId: 100
  podCidrs:
    - "10.244.0.0/16"
  serviceCidrs:
    - "10.96.0.0/12"
```

Apply the transit configuration:
```bash
kubectl apply -f transit-segment-demo.yaml
```

Verify transit state:
```bash
sg-cli transit gateways
sg-cli transit segments
```

---

### 5. Access the Angular 22 Dashboard
Launch the web UI using the CLI helper:

```bash
sg-cli ui
```
Open [http://localhost:4200](http://localhost:4200) in your browser to view:
- Live network topology & cluster map.
- Real-time eBPF packet flow logs and drop counter metrics.
- Active Segment isolation matrix.
- Interactive policy simulation playground.

---

## 🔍 Troubleshooting & Diagnostics

| Symptom | Diagnostic Command | Remediation |
| :--- | :--- | :--- |
| **CNI Not Ready** | `sg-cli status` | Ensure `/sys/fs/bpf` is mounted and kernel is 6.7+. |
| **BPF Map Overflow** | `sg-cli node bpf-maps` | Increase `agent.bpf.conntrackMaxEntries` in `values.yaml`. |
| **Traffic Dropped** | `sg-cli policy simulate` | Review `StraitNetworkPolicy` rules and priority hierarchy. |
| **BGP Session Down** | `sg-cli bgp peers` | Check remote peer IP, ASN, and BFD connectivity on port 179. |
| **WireGuard Unreachable**| `sg-cli wireguard status` | Ensure UDP port 51820 is open between worker nodes/clusters. |

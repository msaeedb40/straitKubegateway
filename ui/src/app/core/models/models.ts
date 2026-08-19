export interface RuntimeConfig {
  api: {
    baseUrl: string;
    grpcWebUrl: string;
    websocketUrl: string;
    sseUrl: string;
  };
  features: {
    topology: boolean;
    flows: boolean;
    observability: boolean;
    policies: boolean;
    gateways: boolean;
    transit: boolean;
    bgp: boolean;
  };
  cluster: {
    name: string;
    region: string;
    zone: string;
  };
  refreshIntervalMs: number;
}

export interface TelemetryContextEnvelope {
  timestamp: string;
  severity: 'INFO' | 'WARN' | 'ERROR' | 'CRITICAL';
  service: {
    name: string;
    version: string;
    component: string;
  };
  environment: {
    name: string;
    region: string;
    zone: string;
  };
  cluster: {
    id: string;
    name: string;
    version: string;
  };
  node?: {
    id: string;
    name: string;
    ip: string;
  };
  namespace?: string;
  resource?: {
    type: string;
    id: string;
    name: string;
    version?: string;
  };
  operation?: {
    name: string;
    type: string;
    action: string;
  };
  network?: {
    source_ip: string;
    destination_ip: string;
    source_port: number;
    destination_port: number;
    protocol: string;
  };
  gateway?: {
    id: string;
    tunnel_id?: string;
    flow_id?: string;
  };
  telemetry: {
    trace_id: string;
    span_id?: string;
    request_id?: string;
  };
}

export interface LatencyBreakdown {
  total_ms: number;
  auth_ms: number;
  validation_ms: number;
  queue_ms: number;
  controller_ms: number;
  kubernetes_ms: number;
  serialization_ms: number;
  network_ms: number;
}

export interface LatencyPercentileMetrics {
  p50: number;
  p90: number;
  p95: number;
  p99: number;
  p99_9: number;
  max: number;
}

export interface GatewayItem {
  id: string;
  name: string;
  namespace: string;
  segmentId: number;
  mode: string;
  listeners: {
    name: string;
    port: number;
    protocol: string;
    hostname?: string;
    tls?: { mode: string; certificateRef: string };
  }[];
  routesCount: number;
  status: 'Ready' | 'Degraded' | 'Pending';
  rxBytes: number;
  txBytes: number;
  latency?: {
    p95_ms: number;
    p99_ms: number;
  };
  health?: {
    availabilityPct: number;
    errorRatePct: number;
  };
}

export interface RouteItem {
  id: string;
  name: string;
  namespace: string;
  gatewayRef: string;
  protocol: string;
  hostnames: string[];
  rules: {
    matches: {
      path?: { type: string; value: string };
      method?: string;
      headers?: { name: string; value: string; type: string }[];
    }[];
    backends: {
      ip: string;
      port: number;
      weight: number;
    }[];
    filters?: {
      type: string;
      setHeaders?: Record<string, string>;
    }[];
  }[];
  latency?: {
    p95_ms: number;
    p99_ms: number;
  };
}

export interface PolicyItem {
  id: string;
  name: string;
  namespace: string;
  scope: 'Cluster' | 'Segment' | 'Namespace';
  priority: number;
  ingressRules: {
    ruleNo: number;
    action: 'Allow' | 'Deny' | 'Reject';
    ports: { port: number; protocol: string }[];
    log: boolean;
  }[];
  egressRules: {
    ruleNo: number;
    action: 'Allow' | 'Deny' | 'Reject';
    ports: { port: number; protocol: string }[];
    log: boolean;
  }[];
  hitCount: number;
  evaluationLatencyP99_us?: number;
}

export interface SegmentItem {
  id: number;
  name: string;
  vni: number;
  endpointsCount: number;
  subnets: string[];
  isolated: boolean;
}

export interface TransitItem {
  id: string;
  name: string;
  asn: number;
  topology: string;
  segmentId: number;
  tunnelType: string;
  attachmentsCount: number;
  activeTunnels: number;
  status: 'Active' | 'Degraded';
}

export interface TunnelItem {
  id: string;
  name: string;
  type: 'WireGuard' | 'IPsec';
  publicKey?: string;
  endpoint: string;
  allowedIPs: string[];
  remoteCluster: string;
  status: 'Established' | 'Connecting' | 'Down';
  rxBytes: number;
  txBytes: number;
  lastHandshake: string;
  latency?: {
    p95_ms: number;
    p99_ms: number;
  };
  throughput?: {
    rxBps: number;
    txBps: number;
  };
  packetDropsPerSec?: number;
}

export interface BGPPeerItem {
  id: string;
  name: string;
  peerAsn: number;
  localAsn: number;
  peerAddress: string;
  localAddress?: string;
  state: 'Established' | 'Connect' | 'Active' | 'Idle';
  holdTime: number;
  keepaliveInterval: number;
  bfdEnabled: boolean;
  receivedPrefixes: number;
  advertisedPrefixes: number;
  uptime: string;
  rttP99_ms?: number;
}

export interface ServiceItem {
  id: string;
  name: string;
  namespace: string;
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName';
  clusterIP: string;
  ports: { port: number; targetPort: number; nodePort?: number; protocol: string }[];
  loadBalancerIPs?: string[];
  dsr: boolean;
  backendsCount: number;
  algorithm: string;
}

export interface ConntrackItem {
  srcIP: string;
  dstIP: string;
  srcPort: number;
  dstPort: number;
  protocol: string;
  state: 'NEW' | 'ESTABLISHED' | 'REPLY' | 'CLOSING' | 'CLOSED';
  natType: 'SNAT' | 'DNAT' | 'MASQUERADE' | 'DIRECT';
  translatedIP: string;
  translatedPort: number;
  bytes: number;
  packets: number;
  ttlRemainingSec: number;
}

export interface FlowItem {
  id: string;
  timestamp: string;
  srcIP: string;
  dstIP: string;
  srcPort: number;
  dstPort: number;
  protocol: string;
  srcIdentity: number;
  dstIdentity: number;
  direction: 'Ingress' | 'Egress';
  verdict: 'FORWARDED' | 'DROPPED' | 'REDIRECTED';
  bytes: number;
  interfaceName: string;
  gatewayId?: string;
  tunnelId?: string;
  latency_us?: number;
  traceId?: string;
}

export interface NodeItem {
  id: string;
  name: string;
  ip: string;
  kernelVersion: string;
  ebpfStatus: 'Active' | 'Degraded';
  bpffsMounted: boolean;
  cgroupV2: boolean;
  cniVersion: string;
  podCIDR: string;
  activeEndpoints: number;
  cpuUsagePct: number;
  memUsagePct: number;
  zone?: string;
  region?: string;
}

export interface EventItem {
  id: string;
  timestamp: string;
  type: 'Normal' | 'Warning' | 'Error';
  component: string;
  message: string;
  resourceRef: string;
  traceId?: string;
  requestId?: string;
}

export interface ClusterStats {
  activeGateways: number;
  totalRoutes: number;
  activePolicies: number;
  totalSegments: number;
  establishedTunnels: number;
  bgpPeersEstablished: number;
  servicesCount: number;
  activeFlowsPerSec: number;
  dropRatePct: number;
  rxThroughputMbps: number;
  txThroughputMbps: number;
  healthyNodes: number;
  totalNodes: number;
  latencyP95_ms?: number;
  latencyP99_ms?: number;
}

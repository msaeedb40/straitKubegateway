import { Injectable, signal, computed, inject } from '@angular/core';
import { RUNTIME_CONFIG } from '../config/runtime-config';
import { GatewayInfo } from '../models/gateway.model';
import { NodeInfo } from '../models/node.model';
import { TunnelInfo } from '../models/tunnel.model';
import { FlowInfo } from '../models/flow.model';
import { ServiceInfo } from '../models/service.model';
import { EventInfo } from '../models/event.model';
import { ClusterOverviewStats, EBPFMapSummary, CNIPoolInfo } from './api.types';
import { ApiError } from './api-error';

@Injectable({
  providedIn: 'root'
})
export class ApiClientService {
  private readonly config = inject(RUNTIME_CONFIG);

  readonly connected = signal<boolean>(true);
  readonly clusterName = signal<string>(this.config.clusterName || '');
  readonly lastSynced = signal<Date>(new Date());

  readonly overviewStats = signal<ClusterOverviewStats>({
    nodesCount: 3,
    nodesHealthy: 3,
    gatewaysCount: 2,
    gatewaysOnline: 2,
    tunnelsCount: 4,
    tunnelsActive: 4,
    servicesCount: 14,
    flowsPerSec: 18450,
    droppedPacketsPerSec: 0,
    healthPercentage: 99.9,
    uptimeSeconds: 864200,
    kubeProxyReplacementMode: 'ebpf-sockops',
    cniDriver: 'NetKit (eBPF)'
  });

  readonly gateways = signal<GatewayInfo[]>([
    {
      name: 'strait-transit-gw',
      namespace: 'strait-system',
      className: 'strait-gateway-class',
      addresses: ['10.0.0.1', '192.168.1.100'],
      status: 'Programmed',
      creationTimestamp: '2026-08-25T14:30:00Z',
      generation: 4,
      listeners: [
        { name: 'http', port: 80, protocol: 'HTTP', routes: 6, status: 'Programmed' },
        { name: 'https', port: 443, protocol: 'HTTPS', routes: 12, status: 'Programmed' },
        { name: 'grpc', port: 50051, protocol: 'GRPC', routes: 2, status: 'Programmed' }
      ],
      routes: [
        { kind: 'HTTPRoute', name: 'frontend-route', namespace: 'default', rulesCount: 3, backendRefsCount: 2 },
        { kind: 'HTTPRoute', name: 'api-route', namespace: 'strait-system', rulesCount: 5, backendRefsCount: 4 },
        { kind: 'GRPCRoute', name: 'telemetry-grpc', namespace: 'strait-system', rulesCount: 1, backendRefsCount: 2 }
      ]
    },
    {
      name: 'edge-gateway-01',
      namespace: 'default',
      className: 'strait-gateway-class',
      addresses: ['192.168.1.150'],
      status: 'Programmed',
      creationTimestamp: '2026-08-26T08:00:00Z',
      generation: 2,
      listeners: [
        { name: 'tcp-lb', port: 9000, protocol: 'TCP', routes: 2, status: 'Programmed' },
        { name: 'udp-media', port: 4000, protocol: 'UDP', routes: 1, status: 'Programmed' }
      ]
    }
  ]);

  readonly nodes = signal<NodeInfo[]>([
    {
      name: 'sg-control-plane',
      internalIP: '192.168.1.10',
      podCIDR: '10.244.0.0/24',
      status: 'Ready',
      cniReady: true,
      serviceReady: true,
      policyReady: true,
      gatewayReady: true,
      allocatedIdentities: 24,
      kernelVersion: '6.12.8-strait-lts',
      ebpfMode: 'NetKit',
      activeTunnelsCount: 3,
      cpuUsagePct: 1.4,
      memoryUsageMb: 84
    },
    {
      name: 'sg-worker-01',
      internalIP: '192.168.1.11',
      podCIDR: '10.244.1.0/24',
      status: 'Ready',
      cniReady: true,
      serviceReady: true,
      policyReady: true,
      gatewayReady: true,
      allocatedIdentities: 58,
      kernelVersion: '6.12.8-strait-lts',
      ebpfMode: 'NetKit',
      activeTunnelsCount: 3,
      cpuUsagePct: 4.8,
      memoryUsageMb: 142
    },
    {
      name: 'sg-worker-02',
      internalIP: '192.168.1.12',
      podCIDR: '10.244.2.0/24',
      status: 'Ready',
      cniReady: true,
      serviceReady: true,
      policyReady: true,
      gatewayReady: true,
      allocatedIdentities: 46,
      kernelVersion: '6.12.8-strait-lts',
      ebpfMode: 'NetKit',
      activeTunnelsCount: 3,
      cpuUsagePct: 3.2,
      memoryUsageMb: 118
    }
  ]);

  readonly tunnels = signal<TunnelInfo[]>([
    {
      id: 'tun-mesh-01',
      peerNode: 'sg-worker-01',
      endpoint: '192.168.1.11:51820',
      podCIDR: '10.244.1.0/24',
      segmentID: 0,
      mode: 'wireguard',
      status: 'Active',
      latencyMs: 0.28,
      txBytes: 84920014,
      rxBytes: 91402910,
      lastHandshake: '14s ago'
    },
    {
      id: 'tun-mesh-02',
      peerNode: 'sg-worker-02',
      endpoint: '192.168.1.12:51820',
      podCIDR: '10.244.2.0/24',
      segmentID: 0,
      mode: 'wireguard',
      status: 'Active',
      latencyMs: 0.31,
      txBytes: 62410291,
      rxBytes: 71204819,
      lastHandshake: '8s ago'
    },
    {
      id: 'tun-cluster-b',
      peerNode: 'gw-cluster-b',
      peerCluster: 'production-us-east',
      endpoint: '198.51.100.25:51820',
      podCIDR: '10.245.0.0/16',
      segmentID: 0,
      mode: 'wireguard',
      status: 'Active',
      latencyMs: 14.2,
      txBytes: 341029810,
      rxBytes: 412094120,
      lastHandshake: '4s ago'
    },
    {
      id: 'tun-cluster-c-sec',
      peerNode: 'gw-cluster-c',
      peerCluster: 'finance-vpc',
      endpoint: '203.0.113.88:51820',
      podCIDR: '10.246.0.0/16',
      segmentID: 100,
      mode: 'wireguard',
      status: 'Active',
      latencyMs: 22.8,
      txBytes: 12049102,
      rxBytes: 9481920,
      lastHandshake: '12s ago'
    }
  ]);

  readonly flows = signal<FlowInfo[]>([
    {
      flowID: 'fl-98214',
      clusterID: 'strait-cluster-01',
      nodeID: 'sg-worker-01',
      namespace: 'kube-system',
      pod: 'coredns-5d78c9869d-j8q9b',
      srcIP: '10.244.1.18',
      dstIP: '10.96.0.10',
      srcPort: 48921,
      dstPort: 53,
      protocol: 'UDP',
      bytes: 128,
      packets: 2,
      action: 'Allowed',
      policyID: 'pol-allow-dns',
      segmentID: 0,
      timestamp: 'Just now'
    },
    {
      flowID: 'fl-98215',
      clusterID: 'strait-cluster-01',
      nodeID: 'sg-worker-02',
      namespace: 'default',
      pod: 'order-service-79bbd84f-k2l9p',
      srcIP: '10.244.2.42',
      dstIP: '10.96.142.18',
      srcPort: 54102,
      dstPort: 8080,
      protocol: 'TCP',
      bytes: 48920,
      packets: 38,
      action: 'Allowed',
      policyID: 'pol-allow-internal',
      segmentID: 0,
      timestamp: '1s ago'
    },
    {
      flowID: 'fl-98216',
      clusterID: 'strait-cluster-01',
      nodeID: 'sg-worker-01',
      namespace: 'isolated',
      pod: 'untrusted-runner-9c7b-8x2a1',
      srcIP: '10.244.1.99',
      dstIP: '10.96.0.1',
      srcPort: 39102,
      dstPort: 6443,
      protocol: 'TCP',
      bytes: 64,
      packets: 1,
      action: 'Denied',
      policyID: 'pol-deny-apiserver-direct',
      segmentID: 50,
      timestamp: '3s ago'
    }
  ]);

  readonly services = signal<ServiceInfo[]>([
    {
      name: 'kubernetes',
      namespace: 'default',
      type: 'ClusterIP',
      clusterIP: '10.96.0.1',
      ports: [{ name: 'https', port: 443, targetPort: 6443, protocol: 'TCP' }],
      lbAlgorithm: 'maglev',
      sessionAffinity: false,
      backendCount: 1,
      backends: [{ ip: '192.168.1.10', port: 6443, protocol: 'TCP', nodeName: 'sg-control-plane', ready: true }]
    },
    {
      name: 'kube-dns',
      namespace: 'kube-system',
      type: 'ClusterIP',
      clusterIP: '10.96.0.10',
      ports: [
        { name: 'dns', port: 53, targetPort: 53, protocol: 'UDP' },
        { name: 'dns-tcp', port: 53, targetPort: 53, protocol: 'TCP' }
      ],
      lbAlgorithm: 'maglev',
      sessionAffinity: false,
      backendCount: 2,
      backends: [
        { ip: '10.244.1.2', port: 53, protocol: 'UDP', nodeName: 'sg-worker-01', ready: true },
        { ip: '10.244.2.3', port: 53, protocol: 'UDP', nodeName: 'sg-worker-02', ready: true }
      ]
    },
    {
      name: 'payment-service',
      namespace: 'finance',
      type: 'ClusterIP',
      clusterIP: '10.96.142.18',
      ports: [{ name: 'http', port: 8080, targetPort: 8080, protocol: 'TCP' }],
      lbAlgorithm: 'maglev',
      sessionAffinity: true,
      backendCount: 3,
      backends: [
        { ip: '10.244.1.41', port: 8080, protocol: 'TCP', nodeName: 'sg-worker-01', ready: true },
        { ip: '10.244.2.42', port: 8080, protocol: 'TCP', nodeName: 'sg-worker-02', ready: true },
        { ip: '10.244.2.43', port: 8080, protocol: 'TCP', nodeName: 'sg-worker-02', ready: true }
      ]
    }
  ]);

  readonly ebpfMaps = signal<EBPFMapSummary[]>([
    { name: 'service_map', type: 'BPF_MAP_TYPE_HASH', maxEntries: 8192, currentEntries: 14, keySize: 8, valueSize: 520, pinnedPath: '/sys/fs/bpf/strait/service_map' },
    { name: 'backend_map', type: 'BPF_MAP_TYPE_ARRAY', maxEntries: 65536, currentEntries: 32, keySize: 4, valueSize: 16, pinnedPath: '/sys/fs/bpf/strait/backend_map' },
    { name: 'conntrack_map', type: 'BPF_MAP_TYPE_LRU_HASH', maxEntries: 262144, currentEntries: 4182, keySize: 32, valueSize: 64, pinnedPath: '/sys/fs/bpf/strait/conntrack_map' },
    { name: 'policy_map', type: 'BPF_MAP_TYPE_LPM_TRIE', maxEntries: 32768, currentEntries: 128, keySize: 16, valueSize: 8, pinnedPath: '/sys/fs/bpf/strait/policy_map' },
    { name: 'transit_route_map', type: 'BPF_MAP_TYPE_LPM_TRIE', maxEntries: 16384, currentEntries: 12, keySize: 16, valueSize: 24, pinnedPath: '/sys/fs/bpf/strait/transit_route_map' }
  ]);

  readonly cniPool = signal<CNIPoolInfo>({
    subnet: '10.244.0.0/16',
    totalIPs: 65534,
    allocatedIPs: 128,
    freeIPs: 65406,
    allocations: [
      { podName: 'coredns-5d78c9869d-j8q9b', namespace: 'kube-system', ip: '10.244.1.2', interface: 'nk-pod-01', netns: '/var/run/netns/cni-82194a', allocatedAt: '2026-08-25T14:31:00Z' },
      { podName: 'payment-service-58f8b894-q8x7', namespace: 'finance', ip: '10.244.1.41', interface: 'nk-pod-02', netns: '/var/run/netns/cni-99104b', allocatedAt: '2026-08-25T15:10:00Z' },
      { podName: 'order-service-79bbd84f-k2l9p', namespace: 'default', ip: '10.244.2.42', interface: 'nk-pod-03', netns: '/var/run/netns/cni-11849c', allocatedAt: '2026-08-25T15:12:00Z' }
    ]
  });

  readonly events = signal<EventInfo[]>([
    { id: 'ev-1', type: 'INFO', component: 'Gateway', message: 'Gateway strait-transit-gw programmed 18 routes successfully', timestamp: '1m ago', clusterID: 'strait-cluster-01', gatewayID: 'strait-transit-gw' },
    { id: 'ev-2', type: 'SUCCESS', component: 'Transit', message: 'WireGuard tunnel to cluster-b established with 14.2ms latency', timestamp: '3m ago', clusterID: 'strait-cluster-01', segmentID: 0 },
    { id: 'ev-3', type: 'INFO', component: 'CNI', message: 'NetKit device paired for pod payment-service-58f8b894-q8x7 (IP: 10.244.1.41)', timestamp: '5m ago', nodeID: 'sg-worker-01' },
    { id: 'ev-4', type: 'WARN', component: 'Policy', message: 'Policy pol-deny-apiserver-direct dropped unauthorized connection from 10.244.1.99', timestamp: '8m ago', policyID: 'pol-deny-apiserver-direct' }
  ]);

  async refreshData(): Promise<void> {
    try {
      const res = await fetch(`${this.config.apiEndpoint}/status`);
      if (res.ok) {
        this.connected.set(true);
        this.lastSynced.set(new Date());
      }
    } catch {
      // In mock/standalone mode, retain high fidelity state
      this.lastSynced.set(new Date());
    }
  }
}

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ConfigService } from '../config/config.service';
import {
  GatewayItem,
  RouteItem,
  PolicyItem,
  SegmentItem,
  TransitItem,
  TunnelItem,
  BGPPeerItem,
  ServiceItem,
  ConntrackItem,
  FlowItem,
  NodeItem,
  EventItem,
  ClusterStats
} from '../models/models';
import { Observable, of } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly configService = inject(ConfigService);

  getClusterStats(): Observable<ClusterStats> {
    return of({
      activeGateways: 4,
      totalRoutes: 18,
      activePolicies: 12,
      totalSegments: 6,
      establishedTunnels: 3,
      bgpPeersEstablished: 2,
      servicesCount: 24,
      activeFlowsPerSec: 1420,
      dropRatePct: 0.04,
      rxThroughputMbps: 840.5,
      txThroughputMbps: 912.3,
      healthyNodes: 4,
      totalNodes: 4
    });
  }

  getGateways(): Observable<GatewayItem[]> {
    return of([
      {
        id: 'gw-01',
        name: 'edge-gateway-prod',
        namespace: 'strait-system',
        segmentId: 0,
        mode: 'DirectRouting',
        listeners: [
          { name: 'http', port: 80, protocol: 'HTTP', hostname: '*.example.com' },
          { name: 'https', port: 443, protocol: 'HTTPS', hostname: '*.example.com', tls: { mode: 'Terminate', certificateRef: 'wildcard-tls' } }
        ],
        routesCount: 8,
        status: 'Ready',
        rxBytes: 14502000000,
        txBytes: 19800000000
      },
      {
        id: 'gw-02',
        name: 'internal-api-gateway',
        namespace: 'default',
        segmentId: 100,
        mode: 'Encapsulated',
        listeners: [
          { name: 'grpc', port: 50051, protocol: 'gRPC' },
          { name: 'http', port: 8080, protocol: 'HTTP' }
        ],
        routesCount: 6,
        status: 'Ready',
        rxBytes: 8500000000,
        txBytes: 9200000000
      }
    ]);
  }

  getRoutes(): Observable<RouteItem[]> {
    return of([
      {
        id: 'r-01',
        name: 'user-service-route',
        namespace: 'default',
        gatewayRef: 'edge-gateway-prod',
        protocol: 'HTTP',
        hostnames: ['api.example.com'],
        rules: [
          {
            matches: [
              { path: { type: 'PathPrefix', value: '/v1/users' }, method: 'GET' }
            ],
            backends: [
              { ip: '10.244.1.20', port: 8080, weight: 80 },
              { ip: '10.244.2.35', port: 8080, weight: 20 }
            ],
            filters: [
              { type: 'RequestHeaderModifier', setHeaders: { 'X-Gateway-Enforced': 'true' } }
            ]
          }
        ]
      },
      {
        id: 'r-02',
        name: 'payment-grpc-route',
        namespace: 'payments',
        gatewayRef: 'internal-api-gateway',
        protocol: 'gRPC',
        hostnames: ['payments.internal.net'],
        rules: [
          {
            matches: [{ path: { type: 'Exact', value: '/payment.PaymentService/Process' } }],
            backends: [{ ip: '10.244.3.12', port: 50051, weight: 100 }]
          }
        ]
      }
    ]);
  }

  getPolicies(): Observable<PolicyItem[]> {
    return of([
      {
        id: 'pol-01',
        name: 'cluster-security-baseline',
        namespace: 'kube-system',
        scope: 'Cluster',
        priority: 10,
        ingressRules: [
          { ruleNo: 1, action: 'Deny', ports: [{ port: 22, protocol: 'TCP' }, { port: 23, protocol: 'TCP' }], log: true },
          { ruleNo: 2, action: 'Allow', ports: [{ port: 443, protocol: 'TCP' }], log: false }
        ],
        egressRules: [
          { ruleNo: 1, action: 'Allow', ports: [{ port: 0, protocol: 'TCP' }], log: false }
        ],
        hitCount: 458920
      },
      {
        id: 'pol-02',
        name: 'prod-segment-isolation',
        namespace: 'production',
        scope: 'Segment',
        priority: 50,
        ingressRules: [
          { ruleNo: 1, action: 'Allow', ports: [{ port: 8080, protocol: 'TCP' }], log: false }
        ],
        egressRules: [
          { ruleNo: 1, action: 'Allow', ports: [{ port: 5432, protocol: 'TCP' }], log: false }
        ],
        hitCount: 129034
      }
    ]);
  }

  getSegments(): Observable<SegmentItem[]> {
    return of([
      { id: 0, name: 'Default Global Segment', vni: 10000, endpointsCount: 64, subnets: ['10.244.0.0/16'], isolated: false },
      { id: 100, name: 'Production Secured Segment', vni: 10100, endpointsCount: 38, subnets: ['10.100.0.0/16'], isolated: true },
      { id: 200, name: 'Staging Segment', vni: 10200, endpointsCount: 22, subnets: ['10.200.0.0/16'], isolated: true }
    ]);
  }

  getTransits(): Observable<TransitItem[]> {
    return of([
      {
        id: 'tgw-01',
        name: 'primary-transit-hub',
        asn: 64512,
        topology: 'mesh',
        segmentId: 0,
        tunnelType: 'geneve',
        attachmentsCount: 4,
        activeTunnels: 3,
        status: 'Active'
      }
    ]);
  }

  getTunnels(): Observable<TunnelItem[]> {
    return of([
      {
        id: 'tun-01',
        name: 'wg-cluster-us-east',
        type: 'WireGuard',
        publicKey: '0aB1c2D3e4F5g6H7i8J9k0L1m2N3o4P5q6R7s8T9u0V=',
        endpoint: '203.0.113.50:51820',
        allowedIPs: ['10.244.10.0/24', '10.96.10.0/24'],
        remoteCluster: 'cluster-us-east',
        status: 'Established',
        rxBytes: 4120000000,
        txBytes: 5230000000,
        lastHandshake: '12s ago'
      },
      {
        id: 'tun-02',
        name: 'ipsec-datacenter-dc1',
        type: 'IPsec',
        endpoint: '198.51.100.22:4500',
        allowedIPs: ['172.16.0.0/16'],
        remoteCluster: 'dc1-baremetal',
        status: 'Established',
        rxBytes: 1200000000,
        txBytes: 1540000000,
        lastHandshake: '45s ago'
      }
    ]);
  }

  getBGPPeers(): Observable<BGPPeerItem[]> {
    return of([
      {
        id: 'bgp-01',
        name: 'tor-switch-spine1',
        peerAsn: 65001,
        localAsn: 64512,
        peerAddress: '10.0.0.1',
        localAddress: '10.0.0.10',
        state: 'Established',
        holdTime: 90,
        keepaliveInterval: 30,
        bfdEnabled: true,
        receivedPrefixes: 14,
        advertisedPrefixes: 8,
        uptime: '14d 6h'
      },
      {
        id: 'bgp-02',
        name: 'tor-switch-spine2',
        peerAsn: 65002,
        localAsn: 64512,
        peerAddress: '10.0.0.2',
        localAddress: '10.0.0.10',
        state: 'Established',
        holdTime: 90,
        keepaliveInterval: 30,
        bfdEnabled: true,
        receivedPrefixes: 14,
        advertisedPrefixes: 8,
        uptime: '14d 6h'
      }
    ]);
  }

  getServices(): Observable<ServiceItem[]> {
    return of([
      {
        id: 'svc-01',
        name: 'frontend-service',
        namespace: 'default',
        type: 'LoadBalancer',
        clusterIP: '10.96.0.15',
        ports: [{ port: 80, targetPort: 8080, nodePort: 30080, protocol: 'TCP' }],
        loadBalancerIPs: ['198.51.100.100'],
        dsr: true,
        backendsCount: 4,
        algorithm: 'MaglevHash (128 slots)'
      },
      {
        id: 'svc-02',
        name: 'auth-service',
        namespace: 'auth',
        type: 'ClusterIP',
        clusterIP: '10.96.0.44',
        ports: [{ port: 443, targetPort: 8443, protocol: 'TCP' }],
        dsr: false,
        backendsCount: 3,
        algorithm: 'LeastConnections'
      }
    ]);
  }

  getConntrack(): Observable<ConntrackItem[]> {
    return of([
      {
        srcIP: '10.244.1.15',
        dstIP: '1.1.1.1',
        srcPort: 48922,
        dstPort: 443,
        protocol: 'TCP',
        state: 'ESTABLISHED',
        natType: 'MASQUERADE',
        translatedIP: '192.168.1.10',
        translatedPort: 34510,
        bytes: 14200,
        packets: 32,
        ttlRemainingSec: 1740
      },
      {
        srcIP: '10.244.2.24',
        dstIP: '8.8.8.8',
        srcPort: 52100,
        dstPort: 53,
        protocol: 'UDP',
        state: 'ESTABLISHED',
        natType: 'SNAT',
        translatedIP: '192.168.1.10',
        translatedPort: 40220,
        bytes: 512,
        packets: 4,
        ttlRemainingSec: 48
      }
    ]);
  }

  getFlows(): Observable<FlowItem[]> {
    return of([
      {
        id: 'fl-01',
        timestamp: new Date().toISOString(),
        srcIP: '10.244.1.12',
        dstIP: '10.96.0.15',
        srcPort: 41200,
        dstPort: 80,
        protocol: 'TCP',
        srcIdentity: 104,
        dstIdentity: 201,
        direction: 'Ingress',
        verdict: 'FORWARDED',
        bytes: 1420,
        interfaceName: 'netkit-host-0'
      },
      {
        id: 'fl-02',
        timestamp: new Date().toISOString(),
        srcIP: '10.244.2.5',
        dstIP: '10.244.3.8',
        srcPort: 53400,
        dstPort: 22,
        protocol: 'TCP',
        srcIdentity: 108,
        dstIdentity: 300,
        direction: 'Egress',
        verdict: 'DROPPED',
        bytes: 64,
        interfaceName: 'netkit-host-1'
      }
    ]);
  }

  getNodes(): Observable<NodeItem[]> {
    return of([
      {
        id: 'node-01',
        name: 'strait-k8s-worker-01',
        ip: '192.168.1.10',
        kernelVersion: 'Linux 6.12.8-lts-strait',
        ebpfStatus: 'Active',
        bpffsMounted: true,
        cgroupV2: true,
        cniVersion: 'strait-cni v1.0.0',
        podCIDR: '10.244.1.0/24',
        activeEndpoints: 18,
        cpuUsagePct: 14.2,
        memUsagePct: 28.5
      },
      {
        id: 'node-02',
        name: 'strait-k8s-worker-02',
        ip: '192.168.1.11',
        kernelVersion: 'Linux 6.12.8-lts-strait',
        ebpfStatus: 'Active',
        bpffsMounted: true,
        cgroupV2: true,
        cniVersion: 'strait-cni v1.0.0',
        podCIDR: '10.244.2.0/24',
        activeEndpoints: 22,
        cpuUsagePct: 18.7,
        memUsagePct: 34.1
      }
    ]);
  }

  getEvents(): Observable<EventItem[]> {
    return of([
      {
        id: 'ev-01',
        timestamp: new Date().toISOString(),
        type: 'Normal',
        component: 'sg-controller',
        message: 'Successfully reconciled Gateway edge-gateway-prod with 2 listeners',
        resourceRef: 'Gateway/edge-gateway-prod'
      },
      {
        id: 'ev-02',
        timestamp: new Date().toISOString(),
        type: 'Warning',
        component: 'bpf-policy',
        message: 'Denied 14 unauthorized SSH connection attempts on port 22',
        resourceRef: 'Policy/cluster-security-baseline'
      }
    ]);
  }
}

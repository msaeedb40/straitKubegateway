export interface TunnelInfo {
  id: string;
  peerNode: string;
  peerCluster?: string;
  mode: 'wireguard' | 'vxlan' | 'geneve' | 'gre' | 'ipsec';
  endpoint: string;
  podCIDR: string;
  segmentID: number;
  status: 'Active' | 'Degraded' | 'Down';
  latencyMs: number;
  txBytes: number;
  rxBytes: number;
  lastHandshake?: string;
}

export interface FlowInfo {
  flowID: string;
  clusterID?: string;
  nodeID?: string;
  namespace?: string;
  pod?: string;
  srcIP: string;
  dstIP: string;
  srcPort: number;
  dstPort: number;
  protocol: 'TCP' | 'UDP' | 'ICMP';
  bytes: number;
  packets: number;
  action: 'Allowed' | 'Denied' | 'Dropped' | 'Rejected';
  policyID?: string;
  segmentID?: number;
  timestamp: string;
}

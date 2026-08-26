import { GatewayInfo } from '../models/gateway.model';
import { NodeInfo } from '../models/node.model';
import { TunnelInfo } from '../models/tunnel.model';
import { FlowInfo } from '../models/flow.model';
import { ServiceInfo } from '../models/service.model';
import { EventInfo } from '../models/event.model';

export interface ClusterOverviewStats {
  nodesCount: number;
  nodesHealthy: number;
  gatewaysCount: number;
  gatewaysOnline: number;
  tunnelsCount: number;
  tunnelsActive: number;
  servicesCount: number;
  flowsPerSec: number;
  droppedPacketsPerSec: number;
  healthPercentage: number;
  uptimeSeconds: number;
  kubeProxyReplacementMode: string;
  cniDriver: string;
}

export interface EBPFMapSummary {
  name: string;
  type: string;
  maxEntries: number;
  currentEntries: number;
  keySize: number;
  valueSize: number;
  pinnedPath: string;
}

export interface CNIPoolInfo {
  subnet: string;
  totalIPs: number;
  allocatedIPs: number;
  freeIPs: number;
  allocations: {
    podName: string;
    namespace: string;
    ip: string;
    interface: string;
    netns: string;
    allocatedAt: string;
  }[];
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
  timestamp: string;
}

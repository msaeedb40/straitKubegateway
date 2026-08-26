export interface NodeInfo {
  name: string;
  internalIP: string;
  podCIDR: string;
  status: 'Ready' | 'Initializing' | 'Error';
  cniReady: boolean;
  serviceReady: boolean;
  policyReady: boolean;
  gatewayReady: boolean;
  allocatedIdentities: number;
  kernelVersion: string;
  ebpfMode?: 'NetKit' | 'TCX' | 'XDP';
  activeTunnelsCount?: number;
  cpuUsagePct?: number;
  memoryUsageMb?: number;
}

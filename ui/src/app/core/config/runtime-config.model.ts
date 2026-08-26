export interface RuntimeConfig {
  apiEndpoint: string;
  metricsEndpoint: string;
  wsEndpoint: string;
  clusterName: string;
  refreshIntervalMs: number;
  enableD3Topology: boolean;
  theme: 'dark' | 'light';
  version?: string;
  features?: {
    cniMetrics?: boolean;
    ebpfMapInspector?: boolean;
    wireguardTransit?: boolean;
    bgpPeering?: boolean;
  };
}

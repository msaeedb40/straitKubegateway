export interface ServiceBackend {
  ip: string;
  port: number;
  protocol: string;
  nodeName?: string;
  ready: boolean;
  weight?: number;
}

export interface ServiceInfo {
  name: string;
  namespace: string;
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer';
  clusterIP: string;
  ports: {
    name?: string;
    port: number;
    nodePort?: number;
    targetPort: number | string;
    protocol: string;
  }[];
  lbAlgorithm?: 'maglev' | 'round-robin' | 'least-conn' | 'consistent-hash';
  sessionAffinity?: boolean;
  backends: ServiceBackend[];
  backendCount: number;
}

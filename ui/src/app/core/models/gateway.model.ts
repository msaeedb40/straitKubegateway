export interface GatewayListener {
  name: string;
  port: number;
  protocol: 'HTTP' | 'HTTPS' | 'TCP' | 'UDP' | 'TLS' | 'GRPC';
  hostname?: string;
  routes: number;
  status: 'Programmed' | 'Pending' | 'Error';
}

export interface GatewayRouteSummary {
  kind: 'HTTPRoute' | 'GRPCRoute' | 'TCPRoute' | 'TLSRoute' | 'UDPRoute';
  name: string;
  namespace: string;
  rulesCount: number;
  backendRefsCount: number;
}

export interface GatewayInfo {
  name: string;
  namespace: string;
  className: string;
  addresses: string[];
  status: 'Programmed' | 'Pending' | 'Error';
  listeners: GatewayListener[];
  routes?: GatewayRouteSummary[];
  creationTimestamp?: string;
  generation?: number;
}

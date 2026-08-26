export interface EventInfo {
  id: string;
  clusterID?: string;
  nodeID?: string;
  namespace?: string;
  pod?: string;
  service?: string;
  endpoint?: string;
  flowID?: string;
  traceID?: string;
  policyID?: string;
  segmentID?: number;
  gatewayID?: string;
  type: 'INFO' | 'WARN' | 'ERROR' | 'SUCCESS';
  message: string;
  component: 'CNI' | 'Dataplane' | 'ServiceLB' | 'Gateway' | 'Policy' | 'Transit' | 'BGP';
  timestamp: string;
}

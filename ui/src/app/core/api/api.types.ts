export interface PaginationParams {
  pageSize?: number;
  cursor?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface PaginatedResult<T> {
  items: T[];
  totalCount: number;
  nextCursor?: string;
  hasMore: boolean;
}

export interface CommandRequest<T = unknown> {
  commandId?: string;
  action: string;
  resourceType: string;
  resourceId: string;
  payload?: T;
}

export interface CommandResponse {
  commandId: string;
  status: 'accepted' | 'completed' | 'rejected' | 'failed';
  traceId: string;
  message?: string;
  timestamp: string;
}

export interface LatencyPercentiles {
  p50_ms?: number;
  p90_ms?: number;
  p95_ms: number;
  p99_ms: number;
  p999_ms?: number;
  max_ms?: number;
}

export interface LatencyBreakdown {
  total_ms: number;
  auth_ms?: number;
  validation_ms?: number;
  queue_ms?: number;
  controller_ms?: number;
  kubernetes_ms?: number;
  serialization_ms?: number;
  network_ms?: number;
}

export interface FlowQuery extends PaginationParams {
  clusterId?: string;
  namespace?: string;
  gatewayId?: string;
  nodeId?: string;
  protocol?: string;
  direction?: 'Ingress' | 'Egress' | string;
  policyId?: string;
  verdict?: 'FORWARDED' | 'DROPPED' | 'REDIRECTED';
  minLatencyP99?: number;
  timeRange?: '5s' | '5m' | '24h' | '7d' | '30d' | '90d';
}

export interface GatewayListRequest extends PaginationParams {
  namespace?: string;
  status?: string;
  segmentId?: number;
}

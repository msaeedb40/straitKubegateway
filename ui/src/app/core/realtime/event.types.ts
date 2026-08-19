export type RealtimeConnectionState = 'CONNECTING' | 'CONNECTED' | 'RECONNECTING' | 'RESYNCING' | 'DISCONNECTED' | 'ERROR';

export type RealtimeEventSeverity = 'INFO' | 'WARNING' | 'ERROR' | 'CRITICAL';

export interface TelemetryCorrelation {
  trace_id: string;
  request_id?: string;
  span_id?: string;
}

export interface ResourceRef {
  type: string;
  id: string;
  name: string;
  namespace?: string;
}

export interface StateTransition<T = unknown> {
  previous?: T | null;
  current: T;
}

export interface EventOperationalContext {
  cluster_id: string;
  node_id?: string;
  namespace?: string;
}

export interface EventLatencyMetrics {
  latency_p95_ms?: number;
  latency_p99_ms?: number;
  error_rate?: number;
  packet_drop_rate?: number;
}

export interface RealtimeDomainEvent<T = unknown> {
  event_id: string;
  sequence: number;
  timestamp: string;
  type: string;
  severity?: RealtimeEventSeverity;
  resource: ResourceRef;
  state?: StateTransition<T>;
  context?: EventOperationalContext;
  metrics?: EventLatencyMetrics;
  correlation?: TelemetryCorrelation;
}

export interface ReplayRequest {
  lastSequence: number;
  streamId?: string;
  requestedAt: string;
}

export interface ReplayResponse<T = unknown> {
  replayAvailable: boolean;
  events?: RealtimeDomainEvent<T>[];
  snapshotRequired?: boolean;
  latestSequence: number;
}

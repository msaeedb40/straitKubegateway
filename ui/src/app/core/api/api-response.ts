export interface ApiResponse<T = unknown> {
  data: T;
  traceId?: string;
  requestId?: string;
  timestamp: string;
  durationMs?: number;
}

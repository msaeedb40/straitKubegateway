export class ApiError extends Error {
  readonly code: string;
  readonly type: string;
  readonly retryable: boolean;
  readonly traceId?: string;
  readonly requestId?: string;
  readonly status?: number;

  constructor(options: {
    code: string;
    type: string;
    message: string;
    retryable?: boolean;
    traceId?: string;
    requestId?: string;
    status?: number;
  }) {
    super(options.message);
    this.name = 'ApiError';
    this.code = options.code;
    this.type = options.type;
    this.retryable = options.retryable ?? false;
    this.traceId = options.traceId;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

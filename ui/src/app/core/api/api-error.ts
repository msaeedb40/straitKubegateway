export class ApiError extends Error {
  constructor(
    public readonly statusCode: number,
    public readonly statusText: string,
    public readonly endpoint: string,
    public override readonly message: string,
    public readonly details?: unknown
  ) {
    super(`API Error [${statusCode}] ${statusText} at ${endpoint}: ${message}`);
    this.name = 'ApiError';
  }

  static fromHttp(response: Response, endpoint: string, bodyText?: string): ApiError {
    return new ApiError(
      response.status,
      response.statusText,
      endpoint,
      bodyText || `HTTP request failed with status ${response.status}`
    );
  }
}

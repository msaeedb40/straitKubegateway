import { inject } from '@angular/core';
import { HttpInterceptorFn, HttpResponse, HttpErrorResponse } from '@angular/common/http';
import { tap, catchError, throwError } from 'rxjs';
import { TelemetryService } from '../logging/telemetry.service';
import { LoggerService } from '../logging/logger.service';
import { ApiError } from './api-error';

export const apiInterceptor: HttpInterceptorFn = (req, next) => {
  const telemetry = inject(TelemetryService);
  const logger = inject(LoggerService);

  const traceId = telemetry.currentTraceId();
  const requestId = telemetry.generateRequestId();
  const startTime = performance.now();

  const cloned = req.clone({
    setHeaders: {
      'X-Trace-Id': traceId,
      'X-Request-Id': requestId,
      'X-Client-Service': 'straitKubegateway-ui',
      'X-Client-Version': '1.0.0'
    }
  });

  return next(cloned).pipe(
    tap(event => {
      if (event instanceof HttpResponse) {
        const durationMs = performance.now() - startTime;
        logger.debug('ApiInterceptor', 'http_request_success', {
          trace_id: traceId,
          request_id: requestId,
          duration_ms: durationMs,
          status: `${event.status} ${event.statusText}`,
          metadata: { url: req.url, method: req.method }
        });
      }
    }),
    catchError((error: HttpErrorResponse) => {
      const durationMs = performance.now() - startTime;
      const normalizedError = new ApiError({
        code: error.error?.code || `HTTP_${error.status}`,
        type: error.name || 'HttpError',
        message: error.error?.message || error.message,
        retryable: error.status >= 500 || error.status === 429,
        traceId,
        requestId,
        status: error.status
      });

      logger.error('ApiInterceptor', 'http_request_failed', {
        trace_id: traceId,
        request_id: requestId,
        duration_ms: durationMs,
        status: `${error.status}`,
        error: {
          code: normalizedError.code,
          type: normalizedError.type,
          message: normalizedError.message,
          retryable: normalizedError.retryable
        },
        metadata: { url: req.url, method: req.method }
      });

      return throwError(() => normalizedError);
    })
  );
};

import { ErrorHandler, Injectable, inject, NgZone } from '@angular/core';
import { LoggerService } from '../logging/logger.service';
import { TelemetryService } from '../logging/telemetry.service';

@Injectable({
  providedIn: 'root'
})
export class GlobalErrorHandler implements ErrorHandler {
  private readonly logger = inject(LoggerService);
  private readonly telemetry = inject(TelemetryService);
  private readonly zone = inject(NgZone);

  handleError(error: unknown): void {
    const err = error instanceof Error ? error : new Error(String(error));
    const traceId = this.telemetry.currentTraceId();

    this.logger.error('GlobalErrorHandler', 'unhandled_exception', {
      trace_id: traceId,
      error: {
        code: 'UNHANDLED_UI_EXCEPTION',
        type: err.name || 'Error',
        message: err.message,
        retryable: false
      },
      metadata: {
        stack: err.stack
      }
    });
  }
}

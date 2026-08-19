import { Injectable, signal } from '@angular/core';

export interface TelemetrySpan {
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  name: string;
  startTime: number;
  tags: Record<string, string | number | boolean>;
}

@Injectable({
  providedIn: 'root'
})
export class TelemetryService {
  private readonly activeTraceId = signal<string>(this.generateId(16));

  readonly currentTraceId = this.activeTraceId.asReadonly();

  generateId(length: number = 16): string {
    const chars = '0123456789abcdef';
    let result = '';
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  }

  generateRequestId(): string {
    return `req-${this.generateId(8)}`;
  }

  generateCommandId(): string {
    return `cmd-${this.generateId(8)}`;
  }

  startSpan(name: string, tags: Record<string, string | number | boolean> = {}): TelemetrySpan {
    return {
      traceId: this.activeTraceId(),
      spanId: this.generateId(16),
      name,
      startTime: performance.now(),
      tags
    };
  }

  endSpan(span: TelemetrySpan, extraTags: Record<string, string | number | boolean> = {}): number {
    const durationMs = performance.now() - span.startTime;
    const completedTags = { ...span.tags, ...extraTags, durationMs };
    return durationMs;
  }

  setTraceId(traceId: string): void {
    this.activeTraceId.set(traceId);
  }
}

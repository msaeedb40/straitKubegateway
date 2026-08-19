import { Injectable, inject } from '@angular/core';
import { TelemetryService } from './telemetry.service';
import { ConfigService } from '../config/config.service';

export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

export interface StructuredLogMessage {
  timestamp: string;
  level: LogLevel;
  service: string;
  component: string;
  event: string;
  cluster_id?: string;
  namespace?: string;
  gateway_id?: string;
  gateway_name?: string;
  actor_id?: string;
  command_id?: string;
  request_id?: string;
  trace_id?: string;
  duration_ms?: number;
  status?: string;
  error?: {
    code: string;
    type: string;
    message: string;
    retryable: boolean;
  };
  metadata?: Record<string, unknown>;
}

@Injectable({
  providedIn: 'root'
})
export class LoggerService {
  private readonly telemetry = inject(TelemetryService);
  private readonly configService = inject(ConfigService);

  private log(level: LogLevel, component: string, event: string, details: Partial<StructuredLogMessage> = {}): void {
    const config = this.configService.config();
    const logObject: StructuredLogMessage = {
      timestamp: new Date().toISOString(),
      level,
      service: 'straitKubegateway-ui',
      component,
      event,
      cluster_id: config.cluster.name,
      trace_id: this.telemetry.currentTraceId(),
      ...details
    };

    if (level === 'ERROR') {
      console.error(JSON.stringify(logObject));
    } else if (level === 'WARN') {
      console.warn(JSON.stringify(logObject));
    } else if (level === 'DEBUG') {
      console.debug(JSON.stringify(logObject));
    } else {
      console.info(JSON.stringify(logObject));
    }
  }

  debug(component: string, event: string, details?: Partial<StructuredLogMessage>): void {
    this.log('DEBUG', component, event, details);
  }

  info(component: string, event: string, details?: Partial<StructuredLogMessage>): void {
    this.log('INFO', component, event, details);
  }

  warn(component: string, event: string, details?: Partial<StructuredLogMessage>): void {
    this.log('WARN', component, event, details);
  }

  error(component: string, event: string, details?: Partial<StructuredLogMessage>): void {
    this.log('ERROR', component, event, details);
  }
}

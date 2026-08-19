import { describe, it, expect, beforeAll, beforeEach } from 'vitest';
import { TestBed } from '@angular/core/testing';
import { provideZonelessChangeDetection } from '@angular/core';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { App } from './app';
import { routes } from './app.routes';
import { AuthService } from './core/auth/auth.service';
import { ConfigService } from './core/config/config.service';
import { TelemetryService } from './core/logging/telemetry.service';
import { calculatePercentiles } from './shared/utils/latency-calculator';

describe('StraitKubeGateway UI Suite', () => {
  it('should compute latency percentiles accurately', () => {
    const latencies = [10, 20, 30, 40, 50, 60, 70, 80, 90, 100];
    const p = calculatePercentiles(latencies);
    expect(p.p50).toBe(50);
    expect(p.p95).toBe(100);
    expect(p.max).toBe(100);
  });

  it('should generate valid OpenTelemetry trace IDs in TelemetryService', () => {
    const telemetry = new TelemetryService();
    const traceId = telemetry.currentTraceId();
    expect(traceId).toBeDefined();
    expect(traceId.length).toBe(16);

    const requestId = telemetry.generateRequestId();
    expect(requestId.startsWith('req-')).toBe(true);
  });

  it('should initialize AuthService with default admin profile', () => {
    const auth = new AuthService();
    expect(auth.isAuthenticated()).toBe(true);
    expect(auth.user()?.username).toBe('cluster-admin');
    expect(auth.hasRole('ClusterAdmin')).toBe(true);
  });
});

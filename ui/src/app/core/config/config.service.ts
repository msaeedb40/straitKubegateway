import { Injectable, signal } from '@angular/core';
import { RuntimeConfig } from '../models/models';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  private readonly configSignal = signal<RuntimeConfig>({
    api: {
      baseUrl: '/api',
      grpcWebUrl: '/grpc',
      websocketUrl: '/ws',
      sseUrl: '/events'
    },
    features: {
      topology: true,
      flows: true,
      observability: true,
      policies: true,
      gateways: true,
      transit: true,
      bgp: true
    },
    cluster: {
      name: 'primary-cluster',
      region: 'default-region',
      zone: 'zone-1'
    },
    refreshIntervalMs: 5000
  });

  readonly config = this.configSignal.asReadonly();

  async loadConfig(): Promise<void> {
    try {
      const res = await fetch('/config/runtime-config.json');
      if (res.ok) {
        const data: RuntimeConfig = await res.json();
        this.configSignal.set(data);
      }
    } catch {
      // Fall back to default runtime configuration
    }
  }
}

import { InjectionToken } from '@angular/core';
import { RuntimeConfig } from './runtime-config.model';

export type { RuntimeConfig };

export const defaultRuntimeConfig: RuntimeConfig = {
  apiEndpoint: '/api/v1',
  metricsEndpoint: '/metrics',
  wsEndpoint: '/ws',
  clusterName: '', // Dynamically discovered from API server / cluster state
  refreshIntervalMs: 5000,
  enableD3Topology: true,
  theme: 'dark',
  version: 'v1.0.0',
  features: {
    cniMetrics: true,
    ebpfMapInspector: true,
    wireguardTransit: true,
    bgpPeering: true
  }
};

export const RUNTIME_CONFIG = new InjectionToken<RuntimeConfig>('RUNTIME_CONFIG', {
  providedIn: 'root',
  factory: () => defaultRuntimeConfig
});

export async function loadRuntimeConfig(): Promise<RuntimeConfig> {
  let config = { ...defaultRuntimeConfig };
  try {
    const res = await fetch('/runtime-config.json');
    if (res.ok) {
      const data = await res.json();
      config = { ...config, ...data };
    }
  } catch (err) {
    console.warn('Using default runtime config:', err);
  }

  // If clusterName is not explicitly set, dynamically discover it from the controller API
  if (!config.clusterName) {
    try {
      const statusRes = await fetch(`${config.apiEndpoint}/status`);
      if (statusRes.ok) {
        const statusData = await statusRes.json();
        if (statusData?.clusterName) {
          config.clusterName = statusData.clusterName;
        }
      }
    } catch {
      // Retain empty string for late dynamic resolution by ApiClientService
    }
  }

  return config;
}

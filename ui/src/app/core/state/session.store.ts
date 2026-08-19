import { Injectable, signal, computed, inject } from '@angular/core';
import { ConfigService } from '../config/config.service';

@Injectable({
  providedIn: 'root'
})
export class SessionStore {
  private readonly configService = inject(ConfigService);

  readonly activeCluster = signal<string>('production-cluster-01');
  readonly activeNamespace = signal<string>('all');
  readonly selectedRegion = signal<string>('region-01');
  readonly timeRange = signal<'5s' | '5m' | '24h' | '7d' | '30d' | '90d'>('5m');

  readonly clusterDetails = computed(() => this.configService.config().cluster);

  setCluster(cluster: string): void {
    this.activeCluster.set(cluster);
  }

  setNamespace(namespace: string): void {
    this.activeNamespace.set(namespace);
  }

  setTimeRange(range: '5s' | '5m' | '24h' | '7d' | '30d' | '90d'): void {
    this.timeRange.set(range);
  }
}

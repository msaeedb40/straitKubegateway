import { Injectable, signal, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { RUNTIME_CONFIG } from '../config/runtime-config';

export type ConnectionStatus = 'Connected' | 'Reconnecting' | 'Disconnected';

@Injectable({
  providedIn: 'root'
})
export class ConnectionService {
  private readonly config = inject(RUNTIME_CONFIG);
  private readonly platformId = inject(PLATFORM_ID);

  readonly status = signal<ConnectionStatus>('Connected');
  readonly pingMs = signal<number>(1.2);
  readonly lastHeartbeat = signal<Date>(new Date());

  constructor() {
    if (isPlatformBrowser(this.platformId)) {
      this.startHeartbeat();
    }
  }

  private startHeartbeat(): void {
    setInterval(() => {
      this.lastHeartbeat.set(new Date());
      // Random subtle jitter for realistic telemetry feel (0.8 - 1.6ms)
      this.pingMs.set(parseFloat((0.8 + Math.random() * 0.8).toFixed(2)));
    }, this.config.refreshIntervalMs || 5000);
  }

  setStatus(newStatus: ConnectionStatus): void {
    this.status.set(newStatus);
  }
}

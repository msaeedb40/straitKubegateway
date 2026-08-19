import { Injectable, signal, computed, inject } from '@angular/core';
import { RealtimeConnectionState } from '../realtime/event.types';
import { RealtimeService } from '../realtime/realtime.service';

@Injectable({
  providedIn: 'root'
})
export class ConnectionStore {
  private readonly realtime = inject(RealtimeService);

  readonly state = computed(() => this.realtime.connectionState());
  readonly isConnected = computed(() => this.state() === 'CONNECTED');
  readonly isReconnecting = computed(() => this.state() === 'RECONNECTING');
  readonly isResyncing = computed(() => this.state() === 'RESYNCING');
  readonly lastHeartbeat = signal<Date>(new Date());

  touchHeartbeat(): void {
    this.lastHeartbeat.set(new Date());
  }
}

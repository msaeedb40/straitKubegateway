import { Injectable, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { Observable, Subject, EMPTY } from 'rxjs';
import { RealtimeDomainEvent } from './event.types';
import { ConfigService } from '../config/config.service';

@Injectable({
  providedIn: 'root'
})
export class WebSocketService {
  private readonly platformId = inject(PLATFORM_ID);
  private readonly configService = inject(ConfigService);
  private socket: WebSocket | null = null;
  private readonly eventSubject = new Subject<RealtimeDomainEvent>();
  private readonly statusSubject = new Subject<'OPEN' | 'CLOSED' | 'ERROR'>();

  readonly events$: Observable<RealtimeDomainEvent> = this.eventSubject.asObservable();
  readonly status$: Observable<'OPEN' | 'CLOSED' | 'ERROR'> = this.statusSubject.asObservable();

  connect(url?: string): void {
    if (!isPlatformBrowser(this.platformId)) {
      return;
    }

    this.disconnect();

    const targetUrl = url || this.configService.config().api.websocketUrl;
    try {
      // In local dev/mock or when full ws backend is not active, handle connection
      const fullUrl = targetUrl.startsWith('ws')
        ? targetUrl
        : `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}${targetUrl}`;

      this.socket = new WebSocket(fullUrl);

      this.socket.onopen = () => {
        this.statusSubject.next('OPEN');
      };

      this.socket.onmessage = (event: MessageEvent) => {
        try {
          const parsed: RealtimeDomainEvent = JSON.parse(event.data);
          this.eventSubject.next(parsed);
        } catch {
          // Non-JSON message or ping
        }
      };

      this.socket.onerror = () => {
        this.statusSubject.next('ERROR');
      };

      this.socket.onclose = () => {
        this.statusSubject.next('CLOSED');
      };
    } catch {
      this.statusSubject.next('ERROR');
    }
  }

  send(data: unknown): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(data));
    }
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }
}

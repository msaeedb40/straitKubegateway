import { Injectable, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { Observable, Subject } from 'rxjs';
import { RealtimeDomainEvent } from './event.types';
import { ConfigService } from '../config/config.service';

@Injectable({
  providedIn: 'root'
})
export class SseService {
  private readonly platformId = inject(PLATFORM_ID);
  private readonly configService = inject(ConfigService);
  private eventSource: EventSource | null = null;
  private readonly eventSubject = new Subject<RealtimeDomainEvent>();
  private readonly statusSubject = new Subject<'OPEN' | 'CLOSED' | 'ERROR'>();

  readonly events$: Observable<RealtimeDomainEvent> = this.eventSubject.asObservable();
  readonly status$: Observable<'OPEN' | 'CLOSED' | 'ERROR'> = this.statusSubject.asObservable();

  connect(url?: string): void {
    if (!isPlatformBrowser(this.platformId)) {
      return;
    }

    this.disconnect();

    const targetUrl = url || this.configService.config().api.sseUrl;
    try {
      this.eventSource = new EventSource(targetUrl);

      this.eventSource.onopen = () => {
        this.statusSubject.next('OPEN');
      };

      this.eventSource.onmessage = (event: MessageEvent) => {
        try {
          const parsed: RealtimeDomainEvent = JSON.parse(event.data);
          this.eventSubject.next(parsed);
        } catch {
          // Non-JSON SSE
        }
      };

      this.eventSource.onerror = () => {
        this.statusSubject.next('ERROR');
      };
    } catch {
      this.statusSubject.next('ERROR');
    }
  }

  disconnect(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
      this.statusSubject.next('CLOSED');
    }
  }
}

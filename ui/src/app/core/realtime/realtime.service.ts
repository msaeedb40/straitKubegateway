import { Injectable, inject, signal, DestroyRef } from '@angular/core';
import { Observable, Subject, filter } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RealtimeDomainEvent, RealtimeConnectionState } from './event.types';
import { WebSocketService } from './websocket.service';
import { SseService } from './sse.service';
import { ReconnectService } from './reconnect.service';
import { ResyncService } from './resync.service';

@Injectable({
  providedIn: 'root'
})
export class RealtimeService {
  private readonly ws = inject(WebSocketService);
  private readonly sse = inject(SseService);
  private readonly reconnect = inject(ReconnectService);
  private readonly resync = inject(ResyncService);
  private readonly destroyRef = inject(DestroyRef);

  private readonly connectionStateSignal = signal<RealtimeConnectionState>('DISCONNECTED');
  private readonly eventBusSubject = new Subject<RealtimeDomainEvent>();

  readonly connectionState = this.connectionStateSignal.asReadonly();
  readonly events$: Observable<RealtimeDomainEvent> = this.eventBusSubject.asObservable();

  private snapshotLoader: (() => Promise<void>) | null = null;

  registerSnapshotLoader(loader: () => Promise<void>): void {
    this.snapshotLoader = loader;
  }

  initialize(): void {
    this.connectionStateSignal.set('CONNECTING');

    // Subscribe to WebSocket events
    this.ws.events$
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(event => this.handleIncomingEvent(event));

    this.ws.status$
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(status => {
        if (status === 'OPEN') {
          this.reconnect.reset();
          this.connectionStateSignal.set('CONNECTED');
          if (this.resync.lastSequence() > 0) {
            this.triggerResync();
          }
        } else if (status === 'CLOSED' || status === 'ERROR') {
          this.connectionStateSignal.set('RECONNECTING');
          this.reconnect.scheduleReconnect(() => this.ws.connect());
        }
      });

    this.ws.connect();
  }

  private handleIncomingEvent(event: RealtimeDomainEvent): void {
    if (event.sequence) {
      this.resync.recordSequence(event.sequence);
    }
    this.eventBusSubject.next(event);
  }

  private async triggerResync(): Promise<void> {
    this.connectionStateSignal.set('RESYNCING');
    await this.resync.requestReplay(
      (replayedEvents) => {
        replayedEvents.forEach(e => this.handleIncomingEvent(e));
        this.connectionStateSignal.set('CONNECTED');
      },
      async () => {
        if (this.snapshotLoader) {
          await this.snapshotLoader();
        }
        this.connectionStateSignal.set('CONNECTED');
      }
    );
  }

  filterEventsByResource(resourceType: string): Observable<RealtimeDomainEvent> {
    return this.events$.pipe(
      filter(event => event.resource.type === resourceType)
    );
  }

  filterEventsByType(eventType: string): Observable<RealtimeDomainEvent> {
    return this.events$.pipe(
      filter(event => event.type === eventType || event.type.startsWith(eventType))
    );
  }

  // Emulate/Dispatch simulated event in browser development
  dispatchLocalEvent(event: RealtimeDomainEvent): void {
    this.handleIncomingEvent(event);
  }
}

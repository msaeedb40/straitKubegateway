import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { RealtimeDomainEvent, ReplayResponse } from './event.types';
import { ConfigService } from '../config/config.service';

@Injectable({
  providedIn: 'root'
})
export class ResyncService {
  private readonly http = inject(HttpClient);
  private readonly configService = inject(ConfigService);

  private readonly lastSequenceSignal = signal<number>(0);
  private readonly isResyncingSignal = signal<boolean>(false);

  readonly lastSequence = this.lastSequenceSignal.asReadonly();
  readonly isResyncing = this.isResyncingSignal.asReadonly();

  recordSequence(sequence: number): void {
    if (sequence > this.lastSequenceSignal()) {
      this.lastSequenceSignal.set(sequence);
    }
  }

  setSequence(sequence: number): void {
    this.lastSequenceSignal.set(sequence);
  }

  async requestReplay(onEvents: (events: RealtimeDomainEvent[]) => void, onSnapshotFallback: () => Promise<void>): Promise<void> {
    this.isResyncingSignal.set(true);
    const lastSeq = this.lastSequenceSignal();
    const baseUrl = this.configService.config().api.baseUrl;

    try {
      if (lastSeq > 0) {
        // Attempt replay request from sg-controller
        const replay = await firstValueFrom(
          this.http.post<ReplayResponse>(`${baseUrl}/v1/realtime/replay`, {
            lastSequence: lastSeq,
            requestedAt: new Date().toISOString()
          })
        ).catch(() => ({ replayAvailable: false, latestSequence: 0 } as ReplayResponse));

        if (replay && replay.replayAvailable && replay.events && replay.events.length > 0) {
          onEvents(replay.events);
          this.lastSequenceSignal.set(replay.latestSequence);
          this.isResyncingSignal.set(false);
          return;
        }
      }

      // Replay unavailable or first sync: Fallback to full snapshot
      await onSnapshotFallback();
    } catch {
      await onSnapshotFallback();
    } finally {
      this.isResyncingSignal.set(false);
    }
  }
}

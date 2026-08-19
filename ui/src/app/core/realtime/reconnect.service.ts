import { Injectable, signal } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ReconnectService {
  private retryCount = signal<number>(0);
  private isReconnecting = signal<boolean>(false);
  private timer: any = null;

  readonly retries = this.retryCount.asReadonly();
  readonly reconnecting = this.isReconnecting.asReadonly();

  scheduleReconnect(onReconnect: () => void, baseDelayMs: number = 1000, maxDelayMs: number = 30000): void {
    if (this.timer) {
      clearTimeout(this.timer);
    }

    this.isReconnecting.set(true);
    const count = this.retryCount();
    const delay = Math.min(baseDelayMs * Math.pow(1.5, count) + Math.random() * 500, maxDelayMs);

    this.timer = setTimeout(() => {
      this.retryCount.update(c => c + 1);
      onReconnect();
    }, delay);
  }

  reset(): void {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    this.retryCount.set(0);
    this.isReconnecting.set(false);
  }
}

import { Injectable, signal } from '@angular/core';

export interface AppNotification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: Date;
  read: boolean;
  autoCloseMs?: number;
}

@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  readonly notifications = signal<AppNotification[]>([
    {
      id: 'notif-1',
      type: 'info',
      title: 'eBPF Dataplane Ready',
      message: 'NetKit driver and sockops connect4 hooks loaded across all 3 nodes.',
      timestamp: new Date(),
      read: false
    }
  ]);

  readonly unreadCount = signal<number>(1);

  show(notif: Omit<AppNotification, 'id' | 'timestamp' | 'read'>): void {
    const newNotif: AppNotification = {
      ...notif,
      id: `notif-${Date.now()}-${Math.random().toString(36).substring(2, 6)}`,
      timestamp: new Date(),
      read: false
    };

    this.notifications.update(current => [newNotif, ...current]);
    this.unreadCount.update(c => c + 1);

    if (notif.autoCloseMs) {
      setTimeout(() => this.dismiss(newNotif.id), notif.autoCloseMs);
    }
  }

  markAllAsRead(): void {
    this.notifications.update(items => items.map(n => ({ ...n, read: true })));
    this.unreadCount.set(0);
  }

  dismiss(id: string): void {
    this.notifications.update(items => items.filter(n => n.id !== id));
  }
}

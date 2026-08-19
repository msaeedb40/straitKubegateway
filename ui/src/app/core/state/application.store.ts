import { Injectable, signal } from '@angular/core';

export interface ToastNotification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: Date;
  traceId?: string;
}

@Injectable({
  providedIn: 'root'
})
export class ApplicationStore {
  readonly isSidebarCollapsed = signal<boolean>(false);
  readonly isCommandPaletteOpen = signal<boolean>(false);
  readonly notifications = signal<ToastNotification[]>([]);
  readonly activeModal = signal<string | null>(null);

  toggleSidebar(): void {
    this.isSidebarCollapsed.update(v => !v);
  }

  setCommandPalette(open: boolean): void {
    this.isCommandPaletteOpen.set(open);
  }

  addNotification(notif: Omit<ToastNotification, 'id' | 'timestamp'>): void {
    const item: ToastNotification = {
      ...notif,
      id: Math.random().toString(36).substring(2, 9),
      timestamp: new Date()
    };
    this.notifications.update(list => [item, ...list].slice(0, 10));

    setTimeout(() => {
      this.dismissNotification(item.id);
    }, 6000);
  }

  dismissNotification(id: string): void {
    this.notifications.update(list => list.filter(n => n.id !== id));
  }
}

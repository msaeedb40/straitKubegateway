import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'sg-toast-overlay',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="fixed bottom-5 right-5 z-50 flex flex-col gap-2 max-w-sm w-full pointer-events-none">
      @for (notif of notifService.notifications(); track notif.id) {
        <div 
          class="pointer-events-auto p-3.5 rounded-xl border shadow-xl backdrop-blur-lg flex items-start justify-between gap-3 animate-slide-in transition-all"
          [ngClass]="bgColorClass(notif.type)">
          <div class="flex items-start gap-2.5">
            <span class="text-sm mt-0.5">{{ iconFor(notif.type) }}</span>
            <div>
              <h4 class="text-xs font-semibold text-slate-100">{{ notif.title }}</h4>
              <p class="text-[11px] text-slate-300 mt-0.5 leading-relaxed">{{ notif.message }}</p>
            </div>
          </div>
          <button 
            (click)="notifService.dismiss(notif.id)"
            class="text-slate-400 hover:text-slate-200 transition text-xs">
            ✕
          </button>
        </div>
      }
    </div>
  `
})
export class ToastOverlayComponent {
  readonly notifService = inject(NotificationService);

  bgColorClass(type: string): string {
    switch (type) {
      case 'success':
        return 'bg-emerald-950/90 border-emerald-500/40 text-emerald-200';
      case 'warning':
        return 'bg-amber-950/90 border-amber-500/40 text-amber-200';
      case 'error':
        return 'bg-rose-950/90 border-rose-500/40 text-rose-200';
      default:
        return 'bg-slate-900/90 border-slate-700/60 text-slate-200';
    }
  }

  iconFor(type: string): string {
    switch (type) {
      case 'success': return '✓';
      case 'warning': return '⚠️';
      case 'error': return '✕';
      default: return 'ℹ️';
    }
  }
}

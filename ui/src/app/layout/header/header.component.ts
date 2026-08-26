import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { ConnectionService } from '../../core/services/connection.service';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'sg-header',
  standalone: true,
  imports: [CommonModule],
  template: `
    <header class="h-16 bg-slate-950/80 backdrop-blur-md border-b border-slate-800/80 flex items-center justify-between px-6 z-20 shrink-0 select-none">
      <!-- Left side: Cluster Context & Title -->
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2">
          <span class="text-xs text-slate-400 font-medium">Cluster:</span>
          <span class="text-xs font-mono font-semibold px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-sky-400">
            {{ api.clusterName() || 'Auto-Discovered' }}
          </span>
        </div>
        <div class="hidden md:flex items-center gap-2 text-xs text-slate-400 border-l border-slate-800 pl-4">
          <span>Mode:</span>
          <span class="text-emerald-400 font-mono text-[11px] font-semibold bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded">
            kube-proxy=none
          </span>
        </div>
      </div>

      <!-- Right side: Actions & Status -->
      <div class="flex items-center gap-3">
        <!-- Quick search prompt -->
        <div class="hidden sm:flex items-center gap-2 text-xs text-slate-400 bg-slate-900/80 border border-slate-800 px-3 py-1 rounded-lg">
          <span>Search</span>
          <kbd class="text-[10px] font-mono bg-slate-800 text-slate-300 px-1.5 py-0.5 rounded border border-slate-700">⌘K</kbd>
        </div>

        <!-- Connection status dot -->
        <div class="flex items-center gap-2 px-3 py-1 rounded-full bg-slate-900 border border-slate-800 text-xs">
          <span class="w-2 h-2 rounded-full" [ngClass]="conn.status() === 'Connected' ? 'bg-emerald-400 animate-pulse' : 'bg-rose-500'"></span>
          <span class="text-slate-200 font-medium text-[11px]">{{ conn.status() }}</span>
          <span class="text-[10px] text-slate-500 font-mono">({{ conn.pingMs() }}ms)</span>
        </div>

        <!-- Notifications button -->
        <button 
          (click)="notifService.markAllAsRead()"
          class="relative p-2 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-slate-300 hover:text-slate-100 transition cursor-pointer"
          title="Notifications">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path>
          </svg>
          @if (notifService.unreadCount() > 0) {
            <span class="absolute top-1 right-1 w-2 h-2 rounded-full bg-sky-400 animate-ping"></span>
            <span class="absolute top-1 right-1 w-2 h-2 rounded-full bg-sky-400"></span>
          }
        </button>
      </div>
    </header>
  `
})
export class HeaderComponent {
  readonly api = inject(ApiClientService);
  readonly conn = inject(ConnectionService);
  readonly notifService = inject(NotificationService);
}

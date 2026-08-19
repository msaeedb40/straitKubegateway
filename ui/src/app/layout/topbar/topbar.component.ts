import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { ConnectionStore } from '../../core/state/connection.store';
import { SessionStore } from '../../core/state/session.store';
import { ApplicationStore } from '../../core/state/application.store';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';

@Component({
  selector: 'app-topbar',
  standalone: true,
  imports: [CommonModule, StatusBadgeComponent],
  template: `
    <header class="h-16 flex-shrink-0 flex items-center justify-between px-8 border-b border-slate-800/80 bg-slate-900/40 backdrop-blur-xl">
      <!-- Search and Quick Action -->
      <div class="flex items-center gap-4 flex-1 max-w-lg">
        <button
          (click)="appStore.setCommandPalette(true)"
          class="w-full flex items-center justify-between bg-slate-950/60 text-xs text-slate-400 rounded-xl px-3.5 py-2 border border-slate-800 hover:border-indigo-500/50 transition-all text-left group"
        >
          <div class="flex items-center gap-2.5">
            <svg class="w-4 h-4 text-slate-500 group-hover:text-indigo-400 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <span>Search Gateways, Flows, Tunnels...</span>
          </div>
          <kbd class="px-2 py-0.5 text-[10px] font-mono bg-slate-800 text-slate-400 rounded border border-slate-700">⌘K</kbd>
        </button>
      </div>

      <!-- Quick Metrics & Live State Indicators -->
      <div class="flex items-center gap-6 text-xs">
        <div class="flex items-center gap-2 text-slate-300">
          <span class="text-slate-500">eBPF Nodes:</span>
          <span class="font-mono font-semibold text-emerald-400">{{ state.healthyNodesCount() }}/{{ state.nodes().length }}</span>
        </div>
        <div class="h-4 w-px bg-slate-800"></div>
        <div class="flex items-center gap-2 text-slate-300">
          <span class="text-slate-500">Active VIPs:</span>
          <span class="font-mono font-semibold text-indigo-300">{{ state.services().length }}</span>
        </div>
        <div class="h-4 w-px bg-slate-800"></div>
        <div class="flex items-center gap-2">
          <span class="text-slate-500">Realtime:</span>
          <app-status-badge [status]="connectionStore.state()">
            {{ connectionStore.state() }}
          </app-status-badge>
        </div>
      </div>
    </header>
  `
})
export class TopbarComponent {
  readonly state = inject(StateService);
  readonly connectionStore = inject(ConnectionStore);
  readonly sessionStore = inject(SessionStore);
  readonly appStore = inject(ApplicationStore);
}

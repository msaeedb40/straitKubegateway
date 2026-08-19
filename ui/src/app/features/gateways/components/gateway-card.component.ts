import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { GatewayItem } from '../../../core/models/models';
import { StatusBadgeComponent } from '../../../shared/components/status-badge/status-badge.component';
import { FormatBytesPipe } from '../../../shared/pipes/format-bytes.pipe';

@Component({
  selector: 'app-gateway-card',
  standalone: true,
  imports: [CommonModule, StatusBadgeComponent, FormatBytesPipe],
  template: `
    <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4 hover:border-indigo-500/40 transition-all">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2.5 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h2 class="text-base font-semibold text-white">{{ gateway.name }}</h2>
              <span class="text-xs px-2 py-0.5 rounded-md bg-slate-800 text-slate-400 font-mono">{{ gateway.namespace }}</span>
              <span class="text-xs px-2 py-0.5 rounded-md bg-indigo-500/10 text-indigo-400 font-mono border border-indigo-500/20">Segment: {{ gateway.segmentId }}</span>
            </div>
            <p class="text-xs text-slate-400 mt-0.5">Mode: {{ gateway.mode }} &bull; {{ gateway.routesCount }} Attached Routes &bull; RX: {{ gateway.rxBytes | formatBytes }} | TX: {{ gateway.txBytes | formatBytes }}</p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <app-status-badge [status]="gateway.status"></app-status-badge>
          <button
            (click)="restart.emit(gateway.id)"
            class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-xs text-slate-300 border border-slate-700 transition-all"
            title="Restart Gateway Pods"
          >
            Restart
          </button>
        </div>
      </div>

      <!-- Listeners Grid -->
      <div class="rounded-xl bg-slate-950/60 border border-slate-800/80 p-4">
        <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider block mb-3">Configured Port Listeners</span>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          @for (l of gateway.listeners; track l.name) {
            <div class="p-3 rounded-lg bg-slate-900/60 border border-slate-800 flex items-center justify-between text-xs">
              <div>
                <div class="font-medium text-slate-200">{{ l.name }} ({{ l.protocol }})</div>
                <div class="text-slate-500 font-mono mt-0.5">{{ l.hostname || '*' }}:{{ l.port }}</div>
              </div>
              @if (l.tls) {
                <span class="px-2 py-0.5 rounded text-[10px] bg-sky-500/10 text-sky-400 border border-sky-500/20 font-mono">
                  TLS {{ l.tls.mode }}
                </span>
              }
            </div>
          }
        </div>
      </div>
    </div>
  `
})
export class GatewayCardComponent {
  @Input({ required: true }) gateway!: GatewayItem;
  @Output() restart = new EventEmitter<string>();
}

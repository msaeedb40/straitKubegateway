import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';

@Component({
  selector: 'app-transit',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Transit Gateways & Multi-Cluster Fabric</h1>
          <p class="text-xs text-slate-400 mt-1">Inter-cluster mesh routing, Transit Attachments, and automatic route propagation across segments.</p>
        </div>
      </div>

      <div class="space-y-4">
        @for (tg of state.transits(); track tg.id) {
          <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="p-2.5 rounded-xl bg-purple-500/10 text-purple-400 border border-purple-500/20">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" /></svg>
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <h2 class="text-base font-semibold text-white">{{ tg.name }}</h2>
                    <span class="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">ASN: {{ tg.asn }}</span>
                    <span class="text-xs px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-300 font-mono border border-indigo-500/20">Segment: {{ tg.segmentId }}</span>
                  </div>
                  <p class="text-xs text-slate-400 mt-0.5">Topology: {{ tg.topology }} &bull; Encapsulation: {{ tg.tunnelType }}</p>
                </div>
              </div>
              <span class="px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> {{ tg.status }}
              </span>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
              <div class="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800 flex items-center justify-between">
                <span class="text-slate-400">Transit Attachments:</span>
                <span class="font-mono text-white font-semibold">{{ tg.attachmentsCount }} Attached</span>
              </div>
              <div class="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800 flex items-center justify-between">
                <span class="text-slate-400">Active Overlay Tunnels:</span>
                <span class="font-mono text-emerald-400 font-semibold">{{ tg.activeTunnels }} Connected</span>
              </div>
            </div>
          </div>
        }
      </div>
    </div>
  `
})
export class TransitComponent {
  readonly state = inject(StateService);
}

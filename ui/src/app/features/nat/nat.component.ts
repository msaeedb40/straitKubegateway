import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';

@Component({
  selector: 'app-nat',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">NAT & Connection Tracking</h1>
          <p class="text-xs text-slate-400 mt-1">Stateful 5-tuple connection tracking, SNAT ephemeral port allocation, Masquerading, and RFC 6052 NAT64 translation.</p>
        </div>
      </div>

      <!-- Conntrack Table -->
      <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-semibold text-white">Active Conntrack Table (131,072 LRU Slots)</h2>
          <span class="text-xs text-slate-500 font-mono">{{ state.conntrack().length }} active tracked sessions</span>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="text-slate-400 border-b border-slate-800">
              <tr>
                <th class="pb-2 font-medium">State</th>
                <th class="pb-2 font-medium">Type</th>
                <th class="pb-2 font-medium">Source (Original)</th>
                <th class="pb-2 font-medium">Destination</th>
                <th class="pb-2 font-medium">Translated (NAT)</th>
                <th class="pb-2 font-medium">TTL</th>
                <th class="pb-2 font-medium">Packets / Bytes</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 text-slate-300 font-mono">
              @for (c of state.conntrack(); track $index) {
                <tr class="hover:bg-slate-800/30 transition-colors">
                  <td class="py-3">
                    <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      {{ c.state }}
                    </span>
                  </td>
                  <td class="py-3 text-indigo-400 font-bold">{{ c.natType }}</td>
                  <td class="py-3 text-slate-200">{{ c.srcIP }}:{{ c.srcPort }} [{{ c.protocol }}]</td>
                  <td class="py-3 text-slate-400">{{ c.dstIP }}:{{ c.dstPort }}</td>
                  <td class="py-3 text-sky-300">{{ c.translatedIP }}:{{ c.translatedPort }}</td>
                  <td class="py-3 text-amber-400">{{ c.ttlRemainingSec }}s</td>
                  <td class="py-3 text-slate-400">{{ c.packets }} pkts ({{ c.bytes | number }} B)</td>
                </tr>
              }
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `
})
export class NatComponent {
  readonly state = inject(StateService);
}

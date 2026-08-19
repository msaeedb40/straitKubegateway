import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';

@Component({
  selector: 'app-bgp',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">BGP Routing & Peering</h1>
          <p class="text-xs text-slate-400 mt-1">BGP-4 and MP-BGP control plane, Top-of-Rack (ToR) switch peering, BFD fast-failover, and Loc-RIB best-path selection.</p>
        </div>
      </div>

      <div class="space-y-4">
        @for (peer of state.bgpPeers(); track peer.id) {
          <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="p-2.5 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <h2 class="text-base font-semibold text-white">{{ peer.name }}</h2>
                    <span class="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">Peer ASN: {{ peer.peerAsn }}</span>
                    <span class="text-xs px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-300 font-mono border border-indigo-500/20">Local ASN: {{ peer.localAsn }}</span>
                  </div>
                  <p class="text-xs text-slate-400 mt-0.5">Peer Address: <span class="text-indigo-300 font-mono">{{ peer.peerAddress }}</span> &bull; Uptime: {{ peer.uptime }}</p>
                </div>
              </div>
              <div class="flex items-center gap-3">
                @if (peer.bfdEnabled) {
                  <span class="px-2.5 py-1 rounded-full text-[11px] font-mono font-medium bg-sky-500/10 text-sky-400 border border-sky-500/20">
                    BFD Active
                  </span>
                }
                <span class="px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5">
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> {{ peer.state }}
                </span>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs font-mono">
              <div class="p-3 rounded-xl bg-slate-950/60 border border-slate-800 flex items-center justify-between">
                <span class="text-slate-500">Advertised Prefixes:</span>
                <span class="text-indigo-400 font-bold">{{ peer.advertisedPrefixes }} Routes</span>
              </div>
              <div class="p-3 rounded-xl bg-slate-950/60 border border-slate-800 flex items-center justify-between">
                <span class="text-slate-500">Received Prefixes:</span>
                <span class="text-emerald-400 font-bold">{{ peer.receivedPrefixes }} Routes</span>
              </div>
              <div class="p-3 rounded-xl bg-slate-950/60 border border-slate-800 flex items-center justify-between">
                <span class="text-slate-500">Hold / Keepalive:</span>
                <span class="text-slate-300">{{ peer.holdTime }}s / {{ peer.keepaliveInterval }}s</span>
              </div>
            </div>
          </div>
        }
      </div>
    </div>
  `
})
export class BgpComponent {
  readonly state = inject(StateService);
}

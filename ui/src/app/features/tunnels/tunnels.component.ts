import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';
import { FormatBytesPipe } from '../../shared/pipes/format-bytes.pipe';

@Component({
  selector: 'app-tunnels',
  standalone: true,
  imports: [CommonModule, StatusBadgeComponent, FormatBytesPipe],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Encrypted Overlay Tunnels</h1>
          <p class="text-xs text-slate-400 mt-1">WireGuard Curve25519 peers and IPsec Security Associations for encrypted inter-cluster transit.</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        @for (tun of state.tunnels(); track tun.id) {
          <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4 hover:border-indigo-500/40 transition-all">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <span class="px-2.5 py-1 rounded-md text-xs font-mono font-bold"
                  [ngClass]="{
                    'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20': tun.type === 'WireGuard',
                    'bg-purple-500/10 text-purple-400 border border-purple-500/20': tun.type === 'IPsec'
                  }">
                  {{ tun.type }}
                </span>
                <div>
                  <h2 class="text-base font-semibold text-white">{{ tun.name }}</h2>
                  <p class="text-xs text-slate-400 mt-0.5">Remote: <span class="text-indigo-300 font-mono">{{ tun.remoteCluster }}</span></p>
                </div>
              </div>
              <app-status-badge [status]="tun.status"></app-status-badge>
            </div>

            <div class="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800 text-xs space-y-2 font-mono">
              <div class="flex justify-between">
                <span class="text-slate-500">Endpoint:</span>
                <span class="text-slate-200">{{ tun.endpoint }}</span>
              </div>
              @if (tun.publicKey) {
                <div class="flex justify-between">
                  <span class="text-slate-500">Public Key:</span>
                  <span class="text-slate-400 truncate max-w-[200px]">{{ tun.publicKey }}</span>
                </div>
              }
              <div class="flex justify-between">
                <span class="text-slate-500">Handshake:</span>
                <span class="text-emerald-400">{{ tun.lastHandshake }}</span>
              </div>
              <div class="flex justify-between pt-2 border-t border-slate-800 text-[11px]">
                <span class="text-slate-500">Transferred:</span>
                <span class="text-slate-300">{{ tun.rxBytes | formatBytes }} RX &bull; {{ tun.txBytes | formatBytes }} TX</span>
              </div>
            </div>
          </div>
        }
      </div>
    </div>
  `
})
export class TunnelsComponent {
  readonly state = inject(StateService);
}

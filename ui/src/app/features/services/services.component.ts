import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';

@Component({
  selector: 'app-services',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Services & kube-proxy Replacement</h1>
          <p class="text-xs text-slate-400 mt-1">Kernel-native L4 load balancing with 128-slot Maglev consistent hashing, Direct Server Return (DSR), and socket-level connect4 routing.</p>
        </div>
      </div>

      <div class="space-y-4">
        @for (s of state.services(); track s.id) {
          <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <span class="px-2.5 py-1 rounded-md text-xs font-mono font-bold"
                  [ngClass]="{
                    'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20': s.type === 'ClusterIP',
                    'bg-amber-500/10 text-amber-400 border border-amber-500/20': s.type === 'NodePort',
                    'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20': s.type === 'LoadBalancer'
                  }">
                  {{ s.type }}
                </span>
                <div>
                  <div class="flex items-center gap-2">
                    <h2 class="text-base font-semibold text-white">{{ s.name }}</h2>
                    <span class="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">{{ s.namespace }}</span>
                  </div>
                  <p class="text-xs text-slate-400 mt-0.5">VIP: <span class="text-indigo-300 font-mono">{{ s.clusterIP }}</span> &bull; Algorithm: <span class="text-slate-300 font-mono">{{ s.algorithm }}</span></p>
                </div>
              </div>
              <div class="flex items-center gap-3">
                @if (s.dsr) {
                  <span class="px-2.5 py-1 rounded-full text-xs font-medium bg-purple-500/10 text-purple-300 border border-purple-500/20">
                    DSR Enabled
                  </span>
                }
                <span class="px-2.5 py-1 rounded-full text-xs font-medium bg-slate-800 text-slate-300 border border-slate-700">
                  {{ s.backendsCount }} Healthy Backends
                </span>
              </div>
            </div>

            <!-- Ports Mapping -->
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              @for (p of s.ports; track p.port) {
                <div class="p-3 rounded-xl bg-slate-950/60 border border-slate-800 text-xs flex items-center justify-between">
                  <div>
                    <span class="text-slate-400">Port / Target:</span>
                    <span class="font-mono text-white font-semibold ml-1.5">{{ p.port }} &rarr; {{ p.targetPort }}</span>
                  </div>
                  @if (p.nodePort) {
                    <span class="px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 font-mono text-[10px] border border-amber-500/20">
                      NodePort: {{ p.nodePort }}
                    </span>
                  }
                </div>
              }
            </div>
          </div>
        }
      </div>
    </div>
  `
})
export class ServicesComponent {
  readonly state = inject(StateService);
}

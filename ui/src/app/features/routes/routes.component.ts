import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';

@Component({
  selector: 'app-routes',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">L7 & L4 Routes</h1>
          <p class="text-xs text-slate-400 mt-1">HTTPRoute, GRPCRoute, and TCP/UDP traffic routing with canary weighted splits and header mutation.</p>
        </div>
      </div>

      <div class="space-y-4">
        @for (r of state.routes(); track r.id) {
          <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <span class="px-2.5 py-1 rounded-md text-xs font-mono font-bold"
                  [ngClass]="{
                    'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20': r.protocol === 'HTTP',
                    'bg-sky-500/10 text-sky-400 border border-sky-500/20': r.protocol === 'gRPC',
                    'bg-purple-500/10 text-purple-400 border border-purple-500/20': r.protocol === 'TCP'
                  }">
                  {{ r.protocol }}
                </span>
                <div>
                  <div class="flex items-center gap-2">
                    <h2 class="text-base font-semibold text-white">{{ r.name }}</h2>
                    <span class="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">{{ r.namespace }}</span>
                  </div>
                  <p class="text-xs text-slate-400 mt-0.5">Parent Gateway: <span class="text-indigo-300 font-mono">{{ r.gatewayRef }}</span></p>
                </div>
              </div>
              <div class="flex items-center gap-2">
                @for (h of r.hostnames; track h) {
                  <span class="text-xs px-2 py-1 rounded-lg bg-slate-950 border border-slate-800 text-slate-300 font-mono">{{ h }}</span>
                }
              </div>
            </div>

            <!-- Rules & Weighted Backends -->
            <div class="space-y-3">
              @for (rule of r.rules; track $index) {
                <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 text-xs space-y-3">
                  <!-- Matcher -->
                  <div class="flex items-center gap-4">
                    <span class="text-slate-400 font-semibold uppercase text-[10px] tracking-wider">Matching:</span>
                    @for (m of rule.matches; track $index) {
                      <div class="flex items-center gap-2">
                        @if (m.path) {
                          <span class="px-2 py-0.5 rounded bg-slate-900 border border-slate-800 font-mono text-indigo-300">
                            Path: {{ m.path.type }} {{ m.path.value }}
                          </span>
                        }
                        @if (m.method) {
                          <span class="px-2 py-0.5 rounded bg-slate-900 border border-slate-800 font-mono text-amber-300">
                            {{ m.method }}
                          </span>
                        }
                      </div>
                    }
                  </div>

                  <!-- Backends & Canary Weights -->
                  <div>
                    <span class="text-slate-400 font-semibold uppercase text-[10px] tracking-wider block mb-2">Canary Backend Split:</span>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                      @for (b of rule.backends; track b.ip) {
                        <div class="p-3 rounded-lg bg-slate-900/60 border border-slate-800/80 flex items-center justify-between">
                          <div>
                            <div class="font-mono text-slate-200">{{ b.ip }}:{{ b.port }}</div>
                            <div class="w-32 bg-slate-800 h-1.5 rounded-full mt-2 overflow-hidden">
                              <div class="bg-gradient-to-r from-indigo-500 to-sky-400 h-full rounded-full" [style.width.%]="b.weight"></div>
                            </div>
                          </div>
                          <span class="text-xs font-mono font-bold text-indigo-400 bg-indigo-500/10 px-2 py-1 rounded-md border border-indigo-500/20">
                            {{ b.weight }}% Weight
                          </span>
                        </div>
                      }
                    </div>
                  </div>
                </div>
              }
            </div>
          </div>
        }
      </div>
    </div>
  `
})
export class RoutesComponent {
  readonly state = inject(StateService);
}

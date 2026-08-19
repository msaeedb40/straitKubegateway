import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';

@Component({
  selector: 'app-policies',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">StraitNetworkPolicies</h1>
          <p class="text-xs text-slate-400 mt-1">Deterministic policy compiler with Scope Hierarchy (Cluster > Segment > Namespace) and Priority + RuleNo ordering.</p>
        </div>
        <div class="flex items-center gap-2 text-xs">
          <span class="px-2.5 py-1 rounded-lg bg-rose-500/10 text-rose-400 border border-rose-500/20 font-medium">
            Default Ingress: Deny-all
          </span>
          <span class="px-2.5 py-1 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium">
            Default Egress: Allow-all
          </span>
        </div>
      </div>

      <!-- Scope Hierarchy Banner -->
      <div class="p-4 rounded-2xl bg-gradient-to-r from-indigo-950/40 via-purple-950/30 to-slate-900/40 border border-indigo-500/20 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-lg bg-indigo-500/20 border border-indigo-500/40 flex items-center justify-center text-indigo-300">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
          </div>
          <div>
            <div class="text-xs font-semibold text-white">Enforcement Hierarchy Order</div>
            <div class="text-[11px] text-slate-400">Cluster Policies (Rank 1) &gt; Segment Policies (Rank 2) &gt; Namespace Policies (Rank 3)</div>
          </div>
        </div>
        <div class="flex items-center gap-2 text-[11px] font-mono">
          <span class="px-2 py-0.5 rounded bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">1. Cluster</span>
          <span class="text-slate-600">&gt;</span>
          <span class="px-2 py-0.5 rounded bg-purple-500/20 text-purple-300 border border-purple-500/30">2. Segment</span>
          <span class="text-slate-600">&gt;</span>
          <span class="px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">3. Namespace</span>
        </div>
      </div>

      <!-- Policies Table -->
      <div class="space-y-4">
        @for (p of state.policies(); track p.id) {
          <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <span class="px-2.5 py-1 rounded-md text-xs font-mono font-bold"
                  [ngClass]="{
                    'bg-indigo-500/10 text-indigo-300 border border-indigo-500/30': p.scope === 'Cluster',
                    'bg-purple-500/10 text-purple-300 border border-purple-500/30': p.scope === 'Segment',
                    'bg-slate-800 text-slate-300 border border-slate-700': p.scope === 'Namespace'
                  }">
                  {{ p.scope }} Scope
                </span>
                <div>
                  <div class="flex items-center gap-2">
                    <h2 class="text-base font-semibold text-white">{{ p.name }}</h2>
                    <span class="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">{{ p.namespace }}</span>
                  </div>
                  <p class="text-xs text-slate-400 mt-0.5">Priority: <span class="text-amber-400 font-mono font-bold">{{ p.priority }}</span> &bull; {{ p.hitCount | number }} Dataplane Hits</p>
                </div>
              </div>
            </div>

            <!-- Ingress & Egress Rules Matrix -->
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 text-xs">
              <!-- Ingress Rules -->
              <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
                <span class="text-slate-400 font-semibold uppercase text-[10px] tracking-wider block mb-2">Ingress Rules (Ordered by RuleNo):</span>
                <div class="space-y-2">
                  @for (r of p.ingressRules; track r.ruleNo) {
                    <div class="p-2.5 rounded-lg bg-slate-900/60 border border-slate-800 flex items-center justify-between">
                      <div class="flex items-center gap-2">
                        <span class="w-5 h-5 rounded-full bg-slate-800 flex items-center justify-center font-mono text-[10px] text-slate-400">#{{ r.ruleNo }}</span>
                        <span class="px-2 py-0.5 rounded font-mono font-bold text-[10px]"
                          [ngClass]="{
                            'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20': r.action === 'Allow',
                            'bg-rose-500/10 text-rose-400 border border-rose-500/20': r.action === 'Deny'
                          }">
                          {{ r.action }}
                        </span>
                        <span class="text-slate-300 font-mono">
                          @for (pt of r.ports; track pt.port) {
                            {{ pt.protocol }}:{{ pt.port === 0 ? '*' : pt.port }}
                          }
                        </span>
                      </div>
                      @if (r.log) {
                        <span class="px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 text-[10px] font-mono border border-amber-500/20">Audit Log</span>
                      }
                    </div>
                  }
                </div>
              </div>

              <!-- Egress Rules -->
              <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
                <span class="text-slate-400 font-semibold uppercase text-[10px] tracking-wider block mb-2">Egress Rules (Ordered by RuleNo):</span>
                <div class="space-y-2">
                  @for (r of p.egressRules; track r.ruleNo) {
                    <div class="p-2.5 rounded-lg bg-slate-900/60 border border-slate-800 flex items-center justify-between">
                      <div class="flex items-center gap-2">
                        <span class="w-5 h-5 rounded-full bg-slate-800 flex items-center justify-center font-mono text-[10px] text-slate-400">#{{ r.ruleNo }}</span>
                        <span class="px-2 py-0.5 rounded font-mono font-bold text-[10px]"
                          [ngClass]="{
                            'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20': r.action === 'Allow',
                            'bg-rose-500/10 text-rose-400 border border-rose-500/20': r.action === 'Deny'
                          }">
                          {{ r.action }}
                        </span>
                        <span class="text-slate-300 font-mono">
                          @for (pt of r.ports; track pt.port) {
                            {{ pt.protocol }}:{{ pt.port === 0 ? '*' : pt.port }}
                          }
                        </span>
                      </div>
                    </div>
                  }
                </div>
              </div>
            </div>
          </div>
        }
      </div>
    </div>
  `
})
export class PoliciesComponent {
  readonly state = inject(StateService);
}

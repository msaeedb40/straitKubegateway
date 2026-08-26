import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';

@Component({
  selector: 'app-policy',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Network Policies & Security Identities</h1>
          <p class="text-xs text-slate-400 mt-1">Identity-based L3/L4/L7 policies enforced via eBPF LSM and NetKit hooks</p>
        </div>
        <button class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition shadow-md">
          + Create Policy
        </button>
      </div>

      <div class="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-950/80 text-slate-400 uppercase text-[10px] tracking-wider border-b border-slate-800">
            <tr>
              <th class="px-6 py-3.5">Policy Name</th>
              <th class="px-6 py-3.5">Namespace</th>
              <th class="px-6 py-3.5">Priority (0-255)</th>
              <th class="px-6 py-3.5">Direction</th>
              <th class="px-6 py-3.5">Target Selectors</th>
              <th class="px-6 py-3.5">Verdict Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60">
            <tr class="hover:bg-slate-800/40 transition">
              <td class="px-6 py-4 font-semibold text-slate-200">default-deny-ingress</td>
              <td class="px-6 py-4 text-slate-400">production</td>
              <td class="px-6 py-4 font-mono text-amber-400 font-semibold">255</td>
              <td class="px-6 py-4 text-slate-300">Ingress</td>
              <td class="px-6 py-4 font-mono text-slate-400">podSelector: &#123;&#125;</td>
              <td class="px-6 py-4">
                <span class="px-2.5 py-1 rounded-full text-[10px] font-medium bg-rose-950 text-rose-400 border border-rose-800/50">
                  Deny
                </span>
              </td>
            </tr>
            <tr class="hover:bg-slate-800/40 transition">
              <td class="px-6 py-4 font-semibold text-slate-200">allow-gateway-to-backend</td>
              <td class="px-6 py-4 text-slate-400">production</td>
              <td class="px-6 py-4 font-mono text-emerald-400 font-semibold">50</td>
              <td class="px-6 py-4 text-slate-300">Ingress</td>
              <td class="px-6 py-4 font-mono text-sky-400">app=api, role=backend</td>
              <td class="px-6 py-4">
                <span class="px-2.5 py-1 rounded-full text-[10px] font-medium bg-emerald-950 text-emerald-400 border border-emerald-800/50">
                  Allow
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  `
})
export class PolicyComponent {
  readonly api = inject(ApiClientService);
}

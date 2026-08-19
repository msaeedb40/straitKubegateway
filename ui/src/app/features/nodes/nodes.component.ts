import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';

@Component({
  selector: 'app-nodes',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Cluster Nodes & eBPF Dataplane</h1>
          <p class="text-xs text-slate-400 mt-1">Node health, Linux kernel versions (6.7+/6.12+ LTS), bpffs map mounts, and CNI agent status.</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        @for (node of state.nodes(); track node.id) {
          <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
            <div class="flex items-center justify-between">
              <div>
                <h2 class="text-base font-semibold text-white">{{ node.name }}</h2>
                <p class="text-xs text-slate-400 mt-0.5">Host IP: <span class="text-indigo-300 font-mono">{{ node.ip }}</span> &bull; {{ node.podCIDR }}</p>
              </div>
              <span class="px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> {{ node.ebpfStatus }}
              </span>
            </div>

            <div class="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800 text-xs space-y-2 font-mono">
              <div class="flex justify-between">
                <span class="text-slate-500">Kernel Version:</span>
                <span class="text-slate-200">{{ node.kernelVersion }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">CNI Agent:</span>
                <span class="text-indigo-400">{{ node.cniVersion }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">bpffs Mounted:</span>
                <span class="text-emerald-400">{{ node.bpffsMounted ? 'Yes (/sys/fs/bpf)' : 'No' }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">Cgroup v2:</span>
                <span class="text-emerald-400">{{ node.cgroupV2 ? 'Active' : 'Disabled' }}</span>
              </div>
              <div class="flex justify-between pt-2 border-t border-slate-800 text-[11px]">
                <span class="text-slate-500">Active Endpoints:</span>
                <span class="text-slate-300">{{ node.activeEndpoints }} Pod NetKits</span>
              </div>
            </div>
          </div>
        }
      </div>
    </div>
  `
})
export class NodesComponent {
  readonly state = inject(StateService);
}

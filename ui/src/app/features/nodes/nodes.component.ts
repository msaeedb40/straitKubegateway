import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { NodeInfo } from '../../core/models/node.model';
import { CardComponent } from '../../shared/components/card/card.component';
import { StatusIndicatorComponent } from '../../shared/status/status-indicator.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';

@Component({
  selector: 'app-nodes',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    StatusIndicatorComponent,
    DataTableComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Node Agents & CNI Runtime</h1>
          <p class="text-xs text-slate-400 mt-1">
            straitd node daemonset health, independent readiness conditions, and NetKit driver bindings
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 font-mono">
            3/3 Nodes Reconciled
          </span>
        </div>
      </div>

      <!-- Node Condition Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
        @for (node of api.nodes(); track node.name) {
          <sg-card [title]="node.name" [subtitle]="node.internalIP">
            <div card-action>
              <sg-status-indicator [status]="node.status === 'Ready' ? 'ready' : 'warning'" [label]="node.status" />
            </div>

            <div class="mt-4 space-y-4 text-xs">
              <!-- Pod CIDR & Mode -->
              <div class="grid grid-cols-2 gap-2 bg-slate-950/70 p-3 rounded-lg border border-slate-800">
                <div>
                  <span class="text-[11px] text-slate-400">Pod CIDR:</span>
                  <div class="font-mono text-sky-400 font-bold mt-0.5">{{ node.podCIDR }}</div>
                </div>
                <div>
                  <span class="text-[11px] text-slate-400">eBPF Hook:</span>
                  <div class="font-mono text-indigo-400 font-bold mt-0.5">{{ node.ebpfMode || 'NetKit' }}</div>
                </div>
              </div>

              <!-- Independent Readiness Matrix (Architectural Invariant 12) -->
              <div>
                <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">
                  Independent Readiness Conditions
                </span>
                <div class="grid grid-cols-2 gap-2 font-mono text-[11px]">
                  <div class="flex items-center justify-between p-2 rounded bg-slate-950 border border-slate-800/80">
                    <span class="text-slate-400">CNI:</span>
                    <span [ngClass]="node.cniReady ? 'text-emerald-400' : 'text-rose-400'">
                      {{ node.cniReady ? '● READY' : '○ PENDING' }}
                    </span>
                  </div>
                  <div class="flex items-center justify-between p-2 rounded bg-slate-950 border border-slate-800/80">
                    <span class="text-slate-400">SERVICE:</span>
                    <span [ngClass]="node.serviceReady ? 'text-emerald-400' : 'text-rose-400'">
                      {{ node.serviceReady ? '● READY' : '○ PENDING' }}
                    </span>
                  </div>
                  <div class="flex items-center justify-between p-2 rounded bg-slate-950 border border-slate-800/80">
                    <span class="text-slate-400">POLICY:</span>
                    <span [ngClass]="node.policyReady ? 'text-emerald-400' : 'text-rose-400'">
                      {{ node.policyReady ? '● READY' : '○ PENDING' }}
                    </span>
                  </div>
                  <div class="flex items-center justify-between p-2 rounded bg-slate-950 border border-slate-800/80">
                    <span class="text-slate-400">GATEWAY:</span>
                    <span [ngClass]="node.gatewayReady ? 'text-emerald-400' : 'text-rose-400'">
                      {{ node.gatewayReady ? '● READY' : '○ PENDING' }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- CPU & Memory Footprint -->
              <div class="flex items-center justify-between text-slate-400 pt-1">
                <span>CPU: <strong class="text-slate-200 font-mono">{{ node.cpuUsagePct }}%</strong></span>
                <span>Memory: <strong class="text-slate-200 font-mono">{{ node.memoryUsageMb }} MB</strong></span>
                <span>BPF IDs: <strong class="text-sky-400 font-mono">{{ node.allocatedIdentities }}</strong></span>
              </div>
            </div>

            <div card-footer class="flex items-center justify-between text-[11px] font-mono">
              <span class="text-slate-400">Kernel: {{ node.kernelVersion }}</span>
              <span class="text-emerald-400">straitd: active</span>
            </div>
          </sg-card>
        }
      </div>

      <!-- Node Inventory Table -->
      <sg-card title="Node Agent Inventory" subtitle="Complete node networking state and capabilities">
        <sg-data-table 
          [columns]="nodeColumns" 
          [data]="api.nodes()"
          searchPlaceholder="Search nodes..." />
      </sg-card>
    </div>
  `
})
export class NodesComponent {
  readonly api = inject(ApiClientService);

  readonly nodeColumns: TableColumn<NodeInfo>[] = [
    { key: 'name', header: 'Node Name', sortable: true },
    { key: 'internalIP', header: 'Internal IP', sortable: true },
    { key: 'podCIDR', header: 'Pod CIDR', sortable: true },
    { key: 'status', header: 'Status', sortable: true },
    { key: 'kernelVersion', header: 'Linux Kernel', sortable: true },
    { key: 'allocatedIdentities', header: 'Security IDs', sortable: true }
  ];
}

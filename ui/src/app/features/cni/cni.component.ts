import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { CardComponent } from '../../shared/components/card/card.component';
import { BandwidthMeterComponent } from '../../shared/charts/bandwidth-meter.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';

@Component({
  selector: 'app-cni',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    BandwidthMeterComponent,
    DataTableComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">CNI Plugin & IPAM Pool</h1>
          <p class="text-xs text-slate-400 mt-1">
            Production-grade CNI v1.1+ lifecycle, NetKit namespace attachment, and zero-dependency IPAM
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs px-3 py-1 rounded-full bg-sky-500/10 border border-sky-500/20 text-sky-400 font-mono">
            CNI Spec: 1.1.0
          </span>
        </div>
      </div>

      <!-- IPAM Pool Statistics -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
        <sg-card title="Cluster Pod Subnet" subtitle="CIDR range & utilization">
          <div class="mt-2 space-y-3">
            <div class="flex justify-between items-center text-xs">
              <span class="text-slate-400">Subnet:</span>
              <span class="font-mono text-sky-400 font-bold text-sm">{{ api.cniPool().subnet }}</span>
            </div>
            <sg-bandwidth-meter 
              label="IP Address Allocation"
              [current]="api.cniPool().allocatedIPs + ' IPs'"
              [max]="api.cniPool().totalIPs + ' IPs'"
              [percentage]="1" />
          </div>
        </sg-card>

        <sg-card title="CNI Lifecycle Verbs" subtitle="Standard compliance">
          <div class="grid grid-cols-2 gap-2 mt-2 font-mono text-xs">
            <div class="p-2 rounded bg-slate-950 border border-slate-800 text-emerald-400">✓ ADD</div>
            <div class="p-2 rounded bg-slate-950 border border-slate-800 text-emerald-400">✓ DEL</div>
            <div class="p-2 rounded bg-slate-950 border border-slate-800 text-emerald-400">✓ CHECK</div>
            <div class="p-2 rounded bg-slate-950 border border-slate-800 text-emerald-400">✓ GC / VERSION</div>
          </div>
        </sg-card>

        <sg-card title="Driver Architecture" subtitle="Linux Kernel 6.6+">
          <div class="space-y-2 mt-2 text-xs">
            <div class="flex justify-between text-slate-400">
              <span>Primary Datapath:</span>
              <span class="font-mono text-indigo-400 font-semibold">NetKit (bpf_redirect)</span>
            </div>
            <div class="flex justify-between text-slate-400">
              <span>Fallback Datapath:</span>
              <span class="font-mono text-slate-300">TCX veth</span>
            </div>
            <div class="flex justify-between text-slate-400">
              <span>Zero Copy:</span>
              <span class="text-emerald-400 font-mono">Enabled</span>
            </div>
          </div>
        </sg-card>
      </div>

      <!-- Allocated Pod IPs Table -->
      <sg-card title="Active Pod IP Allocations" subtitle="Container network namespace attachments and NetKit interfaces">
        <sg-data-table 
          [columns]="cniColumns" 
          [data]="api.cniPool().allocations"
          searchPlaceholder="Search pod IP allocations..." />
      </sg-card>
    </div>
  `
})
export class CniComponent {
  readonly api = inject(ApiClientService);

  readonly cniColumns: TableColumn[] = [
    { key: 'podName', header: 'Pod Name', sortable: true },
    { key: 'namespace', header: 'Namespace', sortable: true },
    { key: 'ip', header: 'Allocated IP', sortable: true },
    { key: 'interface', header: 'Host Interface', sortable: true },
    { key: 'netns', header: 'Network Namespace', sortable: true },
    { key: 'allocatedAt', header: 'Allocated At', sortable: true }
  ];
}

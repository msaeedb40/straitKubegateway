import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { EBPFMapSummary } from '../../core/api/api.types';
import { CardComponent } from '../../shared/components/card/card.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';

@Component({
  selector: 'app-ebpf',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    DataTableComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">eBPF Dataplane & Map Inspector</h1>
          <p class="text-xs text-slate-400 mt-1">
            Pinned BPF maps in /sys/fs/bpf/strait, CO-RE kernel objects, and hook attachment diagnostics
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 font-mono">
            CO-RE Relocations: OK
          </span>
        </div>
      </div>

      <!-- Hook Attachments Overview -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4 text-xs">
        <sg-card title="NetKit (Pod Hook)" subtitle="Fast-path container veth replacement">
          <div class="mt-2 flex items-center justify-between font-mono">
            <span class="text-slate-400">Attached:</span>
            <span class="text-emerald-400 font-bold">● Active</span>
          </div>
          <div card-footer class="font-mono text-[10px] text-slate-400">
            netkit_pod.c | 0-copy
          </div>
        </sg-card>

        <sg-card title="TCX (Node Ingress/Egress)" subtitle="Host packet classifier">
          <div class="mt-2 flex items-center justify-between font-mono">
            <span class="text-slate-400">Attached:</span>
            <span class="text-emerald-400 font-bold">● Active</span>
          </div>
          <div card-footer class="font-mono text-[10px] text-slate-400">
            service_lb.c | TC_INGRESS
          </div>
        </sg-card>

        <sg-card title="sockops / cgroup" subtitle="Socket-level VIP rewrite on connect()">
          <div class="mt-2 flex items-center justify-between font-mono">
            <span class="text-slate-400">Attached:</span>
            <span class="text-emerald-400 font-bold">● Active</span>
          </div>
          <div card-footer class="font-mono text-[10px] text-slate-400">
            cgroup/connect4 | bypass IP stack
          </div>
        </sg-card>

        <sg-card title="XDP (DDoS & Ingress)" subtitle="Earliest NIC driver layer">
          <div class="mt-2 flex items-center justify-between font-mono">
            <span class="text-slate-400">Attached:</span>
            <span class="text-emerald-400 font-bold">● Active</span>
          </div>
          <div card-footer class="font-mono text-[10px] text-slate-400">
            xdp_ingress.c | DRIVER
          </div>
        </sg-card>
      </div>

      <!-- Pinned Maps Table Card -->
      <sg-card title="Pinned Kernel eBPF Maps" subtitle="Direct memory maps populated by Dataplane Compiler">
        <sg-data-table 
          [columns]="mapColumns" 
          [data]="api.ebpfMaps()"
          searchPlaceholder="Search BPF maps..." />
      </sg-card>
    </div>
  `
})
export class EbpfComponent {
  readonly api = inject(ApiClientService);

  readonly mapColumns: TableColumn<EBPFMapSummary>[] = [
    { key: 'name', header: 'Map Name', sortable: true },
    { key: 'type', header: 'Map Type', sortable: true },
    { 
      key: 'currentEntries', 
      header: 'Entries (Used / Max)', 
      sortable: true,
      render: (m) => `${m.currentEntries} / ${m.maxEntries}`
    },
    { 
      key: 'keySize', 
      header: 'Key / Val Size', 
      sortable: true,
      render: (m) => `${m.keySize}B / ${m.valueSize}B`
    },
    { key: 'pinnedPath', header: 'BPF FS Path', sortable: true }
  ];
}

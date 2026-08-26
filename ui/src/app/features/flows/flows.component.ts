import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { FlowInfo } from '../../core/models/flow.model';
import { CardComponent } from '../../shared/components/card/card.component';
import { ButtonComponent } from '../../shared/components/button/button.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';

@Component({
  selector: 'app-flows',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    ButtonComponent,
    DataTableComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Live eBPF Flow Monitor</h1>
          <p class="text-xs text-slate-400 mt-1">
            Real-time 5-tuple kernel flow visibility captured directly from NetKit and TCX hooks
          </p>
        </div>
        <div class="flex items-center gap-2">
          <sg-button [variant]="isStreaming() ? 'primary' : 'secondary'" (clicked)="toggleStreaming()">
            <span class="w-2 h-2 rounded-full" [ngClass]="isStreaming() ? 'bg-emerald-300 animate-ping' : 'bg-slate-400'"></span>
            <span>{{ isStreaming() ? 'Streaming Live' : 'Paused' }}</span>
          </sg-button>
        </div>
      </div>

      <!-- Flow Table Card -->
      <sg-card title="Active Kernel Flow Events" subtitle="5-tuple L3/L4 identity, policy verdicts, and byte counters">
        <sg-data-table 
          [columns]="flowColumns" 
          [data]="api.flows()"
          searchPlaceholder="Filter flows by IP, port, policy..." />
      </sg-card>
    </div>
  `
})
export class FlowsComponent {
  readonly api = inject(ApiClientService);
  readonly isStreaming = signal<boolean>(true);

  readonly flowColumns: TableColumn<FlowInfo>[] = [
    { key: 'flowID', header: 'Flow ID', sortable: true },
    { 
      key: 'srcIP', 
      header: 'Source (5-Tuple)', 
      sortable: true,
      render: (f) => `${f.srcIP}:${f.srcPort}` 
    },
    { 
      key: 'dstIP', 
      header: 'Destination (5-Tuple)', 
      sortable: true,
      render: (f) => `${f.dstIP}:${f.dstPort}` 
    },
    { key: 'protocol', header: 'Protocol', sortable: true },
    { key: 'action', header: 'Verdict', sortable: true },
    { 
      key: 'bytes', 
      header: 'Traffic', 
      sortable: true,
      render: (f) => `${f.bytes} B (${f.packets} pkts)` 
    },
    { key: 'policyID', header: 'Enforced Policy', sortable: true },
    { key: 'timestamp', header: 'Captured', sortable: true }
  ];

  toggleStreaming(): void {
    this.isStreaming.update(s => !s);
  }
}

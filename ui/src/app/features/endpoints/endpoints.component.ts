import { Component, inject, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { CardComponent } from '../../shared/components/card/card.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';

interface EndpointRow {
  service: string;
  namespace: string;
  backendIP: string;
  port: number;
  protocol: string;
  nodeName: string;
  ready: boolean;
}

@Component({
  selector: 'app-endpoints',
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
          <h1 class="text-xl font-bold text-slate-100">EndpointSlices & Backend Pools</h1>
          <p class="text-xs text-slate-400 mt-1">
            Dynamic discovery and direct synchronization between EndpointSlices and eBPF backend_map
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs px-3 py-1 rounded-full bg-sky-500/10 border border-sky-500/20 text-sky-400 font-mono">
            {{ flattenedEndpoints().length }} Backends Active
          </span>
        </div>
      </div>

      <!-- Endpoint Table Card -->
      <sg-card title="Endpoint Directory" subtitle="All registered pod endpoints compiled into kernel memory">
        <sg-data-table 
          [columns]="endpointColumns" 
          [data]="flattenedEndpoints()"
          searchPlaceholder="Search endpoints by IP, service, node..." />
      </sg-card>
    </div>
  `
})
export class EndpointsComponent {
  readonly api = inject(ApiClientService);

  readonly flattenedEndpoints = computed<EndpointRow[]>(() => {
    const list: EndpointRow[] = [];
    for (const svc of this.api.services()) {
      for (const be of svc.backends) {
        list.push({
          service: svc.name,
          namespace: svc.namespace,
          backendIP: be.ip,
          port: be.port,
          protocol: be.protocol,
          nodeName: be.nodeName || 'unknown-node',
          ready: be.ready
        });
      }
    }
    return list;
  });

  readonly endpointColumns: TableColumn<EndpointRow>[] = [
    { key: 'service', header: 'Service', sortable: true },
    { key: 'namespace', header: 'Namespace', sortable: true },
    { 
      key: 'backendIP', 
      header: 'Backend (IP:Port)', 
      sortable: true,
      render: (e) => `${e.backendIP}:${e.port} (${e.protocol})`
    },
    { key: 'nodeName', header: 'Host Node', sortable: true },
    { 
      key: 'ready', 
      header: 'Health', 
      sortable: true,
      render: (e) => e.ready ? '● HEALTHY' : '○ DEGRADED'
    }
  ];
}

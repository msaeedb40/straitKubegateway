import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { ServiceInfo } from '../../core/models/service.model';
import { CardComponent } from '../../shared/components/card/card.component';
import { BadgeComponent } from '../../shared/components/badge/badge.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';

@Component({
  selector: 'app-services',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    BadgeComponent,
    DataTableComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Kernel Service Load Balancer</h1>
          <p class="text-xs text-slate-400 mt-1">
            Complete kube-proxy replacement: Maglev consistent hashing, socket connect4 load balancing, and backend health
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 font-mono">
            Maglev Table Size: 128
          </span>
        </div>
      </div>

      <!-- Service Cards Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
        @for (svc of api.services(); track svc.name) {
          <sg-card [title]="svc.name" [subtitle]="svc.namespace + ' / ' + svc.type">
            <div card-action>
              <sg-badge variant="primary">{{ svc.lbAlgorithm || 'maglev' }}</sg-badge>
            </div>

            <div class="mt-4 space-y-3 text-xs">
              <div class="flex justify-between items-center bg-slate-950/70 p-2.5 rounded-lg border border-slate-800">
                <span class="text-slate-400">ClusterIP VIP:</span>
                <span class="font-mono text-sky-400 font-bold">{{ svc.clusterIP }}</span>
              </div>

              <!-- Ports -->
              <div>
                <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-1.5">Ports</span>
                <div class="space-y-1">
                  @for (p of svc.ports; track p.port) {
                    <div class="flex items-center justify-between p-1.5 rounded bg-slate-950 border border-slate-800/80 font-mono text-[11px]">
                      <span class="text-slate-300">{{ p.name || 'port' }}</span>
                      <span class="text-indigo-400 font-bold">{{ p.port }} / {{ p.protocol }} → :{{ p.targetPort }}</span>
                    </div>
                  }
                </div>
              </div>

              <!-- Backends preview -->
              <div>
                <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-1.5">
                  Backends ({{ svc.backendCount }})
                </span>
                <div class="space-y-1">
                  @for (be of svc.backends; track be.ip) {
                    <div class="flex items-center justify-between p-1.5 rounded bg-slate-950/60 border border-slate-800/50 font-mono text-[11px]">
                      <div class="flex items-center gap-1.5">
                        <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
                        <span class="text-slate-300">{{ be.ip }}:{{ be.port }}</span>
                      </div>
                      <span class="text-slate-500 text-[10px]">{{ be.nodeName }}</span>
                    </div>
                  }
                </div>
              </div>
            </div>

            <div card-footer class="flex items-center justify-between text-[11px]">
              <span class="text-slate-400">Affinity: {{ svc.sessionAffinity ? 'Enabled' : 'None' }}</span>
              <span class="text-emerald-400 font-mono">DSR: ready</span>
            </div>
          </sg-card>
        }
      </div>

      <!-- Full Service Table -->
      <sg-card title="Services & VIP Mapping" subtitle="All compiled ClusterIP and NodePort mappings in eBPF service_map">
        <sg-data-table 
          [columns]="serviceColumns" 
          [data]="api.services()"
          searchPlaceholder="Search services..." />
      </sg-card>
    </div>
  `
})
export class ServicesComponent {
  readonly api = inject(ApiClientService);

  readonly serviceColumns: TableColumn<ServiceInfo>[] = [
    { key: 'name', header: 'Service Name', sortable: true },
    { key: 'namespace', header: 'Namespace', sortable: true },
    { key: 'type', header: 'Type', sortable: true },
    { key: 'clusterIP', header: 'ClusterIP VIP', sortable: true },
    { key: 'backendCount', header: 'Healthy Backends', sortable: true },
    { key: 'lbAlgorithm', header: 'LB Algorithm', sortable: true }
  ];
}

import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { GatewayInfo } from '../../core/models/gateway.model';
import { CardComponent } from '../../shared/components/card/card.component';
import { BadgeComponent } from '../../shared/components/badge/badge.component';
import { ButtonComponent } from '../../shared/components/button/button.component';
import { StatusIndicatorComponent } from '../../shared/status/status-indicator.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';

@Component({
  selector: 'app-gateways',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    BadgeComponent,
    ButtonComponent,
    StatusIndicatorComponent,
    DataTableComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Gateway API Controller</h1>
          <p class="text-xs text-slate-400 mt-1">Conformant Gateway API v1.6.1 management, listeners, and route bindings</p>
        </div>
        <div class="flex items-center gap-2">
          <sg-button variant="primary" (clicked)="showNewGatewayModal.set(true)">
            + New Gateway
          </sg-button>
        </div>
      </div>

      <!-- Gateway Cards Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
        @for (gw of api.gateways(); track gw.name) {
          <sg-card [title]="gw.name" [subtitle]="gw.namespace + ' / ' + gw.className">
            <div card-action>
              <sg-status-indicator [status]="gw.status === 'Programmed' ? 'ready' : 'pending'" [label]="gw.status" />
            </div>

            <div class="mt-4 space-y-4">
              <!-- Addresses & Classes -->
              <div class="grid grid-cols-2 gap-3 text-xs bg-slate-950/60 p-3 rounded-lg border border-slate-800/80">
                <div>
                  <span class="text-slate-400 text-[11px]">VIP Addresses:</span>
                  <div class="font-mono text-slate-200 font-medium mt-0.5">
                    {{ gw.addresses.join(', ') }}
                  </div>
                </div>
                <div>
                  <span class="text-slate-400 text-[11px]">GatewayClass:</span>
                  <div class="font-mono text-sky-400 font-medium mt-0.5">
                    {{ gw.className }}
                  </div>
                </div>
              </div>

              <!-- Listeners List -->
              <div>
                <h4 class="text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">Programmed Listeners</h4>
                <div class="space-y-1.5">
                  @for (lis of gw.listeners; track lis.name) {
                    <div class="flex items-center justify-between p-2 rounded-lg bg-slate-950/80 border border-slate-800 text-xs font-mono">
                      <div class="flex items-center gap-2">
                        <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
                        <span class="text-slate-200 font-semibold">{{ lis.name }}</span>
                        <span class="text-slate-500">:</span>
                        <span class="text-sky-400 font-bold">{{ lis.port }}</span>
                        <sg-badge variant="info">{{ lis.protocol }}</sg-badge>
                      </div>
                      <span class="text-slate-400 text-[11px] font-sans">{{ lis.routes }} attached routes</span>
                    </div>
                  }
                </div>
              </div>

              <!-- Attached Routes Preview -->
              @if (gw.routes && gw.routes.length > 0) {
                <div>
                  <h4 class="text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">Attached Routes</h4>
                  <div class="space-y-1.5">
                    @for (r of gw.routes; track r.name) {
                      <div class="flex items-center justify-between p-2 rounded-lg bg-slate-950/50 border border-slate-800/60 text-xs">
                        <div class="flex items-center gap-2">
                          <sg-badge [variant]="r.kind === 'HTTPRoute' ? 'primary' : 'secondary'">{{ r.kind }}</sg-badge>
                          <span class="font-mono text-slate-200">{{ r.namespace }}/{{ r.name }}</span>
                        </div>
                        <span class="text-slate-400 font-mono text-[11px]">{{ r.rulesCount }} rules | {{ r.backendRefsCount }} backends</span>
                      </div>
                    }
                  </div>
                </div>
              }
            </div>

            <div card-footer class="flex items-center justify-between">
              <span class="font-mono text-[11px]">Gen: {{ gw.generation || 1 }}</span>
              <span class="text-[11px] text-slate-400">Created: {{ gw.creationTimestamp || 'recently' }}</span>
            </div>
          </sg-card>
        }
      </div>

      <!-- Full Gateway Table -->
      <sg-card title="Gateway Inventory" subtitle="All registered GatewayClass and Gateway instances">
        <sg-data-table 
          [columns]="gatewayColumns" 
          [data]="api.gateways()"
          searchPlaceholder="Search gateways..." />
      </sg-card>
    </div>
  `
})
export class GatewaysComponent {
  readonly api = inject(ApiClientService);
  readonly showNewGatewayModal = signal<boolean>(false);

  readonly gatewayColumns: TableColumn<GatewayInfo>[] = [
    { key: 'name', header: 'Gateway Name', sortable: true },
    { key: 'namespace', header: 'Namespace', sortable: true },
    { key: 'className', header: 'Gateway Class', sortable: true },
    { key: 'status', header: 'Status', sortable: true },
    { 
      key: 'addresses', 
      header: 'Addresses', 
      render: (gw) => gw.addresses.join(', ') 
    }
  ];
}

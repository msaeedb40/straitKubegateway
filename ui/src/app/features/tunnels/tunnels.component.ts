import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { TunnelInfo } from '../../core/models/tunnel.model';
import { CardComponent } from '../../shared/components/card/card.component';
import { BadgeComponent } from '../../shared/components/badge/badge.component';
import { StatusIndicatorComponent } from '../../shared/status/status-indicator.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';
import { BytesPipe } from '../../shared/utilities/bytes.pipe';

@Component({
  selector: 'app-tunnels',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    BadgeComponent,
    StatusIndicatorComponent,
    DataTableComponent,
    BytesPipe
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Multi-Cluster Transit & Mesh Tunnels</h1>
          <p class="text-xs text-slate-400 mt-1">
            Segment 0 default backbone, WireGuard / VXLAN mesh tunnels, and inter-cluster route exchange
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 font-mono">
            Segment 0 Backbone Active
          </span>
        </div>
      </div>

      <!-- Tunnel KPI Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        @for (tun of api.tunnels(); track tun.id) {
          <sg-card [title]="tun.peerCluster || tun.peerNode" [subtitle]="tun.id">
            <div card-action>
              <sg-status-indicator [status]="tun.status === 'Active' ? 'ready' : 'warning'" [label]="tun.status" />
            </div>

            <div class="mt-4 space-y-3 text-xs">
              <div class="flex justify-between items-center bg-slate-950/70 p-2 rounded-lg border border-slate-800">
                <span class="text-slate-400">Segment:</span>
                <span class="font-mono text-amber-400 font-bold">Segment {{ tun.segmentID }}</span>
              </div>
              <div class="flex justify-between items-center bg-slate-950/70 p-2 rounded-lg border border-slate-800">
                <span class="text-slate-400">Encapsulation:</span>
                <sg-badge variant="info">{{ tun.mode | uppercase }}</sg-badge>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-slate-400">Latency:</span>
                <span class="font-mono text-emerald-400 font-bold">{{ tun.latencyMs }} ms</span>
              </div>
              <div class="flex justify-between items-center text-[11px] text-slate-400 pt-1 border-t border-slate-800/60 font-mono">
                <span>TX: {{ tun.txBytes | bytes }}</span>
                <span>RX: {{ tun.rxBytes | bytes }}</span>
              </div>
            </div>

            <div card-footer class="text-[10px] text-slate-500 font-mono">
              Endpoint: {{ tun.endpoint }}
            </div>
          </sg-card>
        }
      </div>

      <!-- Transit Tunnel Table -->
      <sg-card title="Transit Tunnels Inventory" subtitle="Complete peer mesh connection status and metrics">
        <sg-data-table 
          [columns]="tunnelColumns" 
          [data]="api.tunnels()"
          searchPlaceholder="Filter tunnels..." />
      </sg-card>
    </div>
  `
})
export class TunnelsComponent {
  readonly api = inject(ApiClientService);

  readonly tunnelColumns: TableColumn<TunnelInfo>[] = [
    { key: 'id', header: 'Tunnel ID', sortable: true },
    { key: 'peerNode', header: 'Peer Node / Cluster', sortable: true },
    { key: 'endpoint', header: 'Remote Endpoint', sortable: true },
    { key: 'podCIDR', header: 'Remote Pod CIDR', sortable: true },
    { key: 'mode', header: 'Mode', sortable: true },
    { key: 'status', header: 'Status', sortable: true },
    { key: 'latencyMs', header: 'Latency (ms)', sortable: true }
  ];
}

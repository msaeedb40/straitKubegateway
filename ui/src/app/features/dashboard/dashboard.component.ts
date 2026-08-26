import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ApiClientService } from '../../core/api/api-client';
import { CardComponent } from '../../shared/components/card/card.component';
import { StatusIndicatorComponent } from '../../shared/status/status-indicator.component';
import { TimeseriesChartComponent } from '../../shared/charts/timeseries-chart.component';
import { DonutChartComponent } from '../../shared/charts/donut-chart.component';
import { BandwidthMeterComponent } from '../../shared/charts/bandwidth-meter.component';
import { D3TopologyGraphComponent } from '../../shared/topology/d3-topology-graph.component';
import { DurationPipe } from '../../shared/utilities/duration.pipe';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    CardComponent,
    StatusIndicatorComponent,
    TimeseriesChartComponent,
    DonutChartComponent,
    BandwidthMeterComponent,
    D3TopologyGraphComponent,
    DurationPipe
  ],
  template: `
    <div class="space-y-6">
      <!-- Top KPI Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Nodes KPI -->
        <sg-card title="Cluster Nodes" subtitle="straitd node agents">
          <div class="flex items-baseline justify-between mt-2">
            <span class="text-3xl font-bold font-mono text-slate-100">{{ api.overviewStats().nodesHealthy }} / {{ api.overviewStats().nodesCount }}</span>
            <sg-status-indicator status="ready" label="All Ready" />
          </div>
          <div card-footer class="flex justify-between items-center">
            <span>NetKit CNI active</span>
            <a routerLink="/nodes" class="text-sky-400 hover:underline">View Nodes →</a>
          </div>
        </sg-card>

        <!-- Gateways KPI -->
        <sg-card title="Gateways" subtitle="Gateway API v1.6.1">
          <div class="flex items-baseline justify-between mt-2">
            <span class="text-3xl font-bold font-mono text-sky-400">{{ api.gateways().length }}</span>
            <span class="text-xs px-2 py-0.5 rounded bg-sky-500/10 text-sky-400 font-mono">Programmed</span>
          </div>
          <div card-footer class="flex justify-between items-center">
            <span>18 active listeners</span>
            <a routerLink="/gateways" class="text-sky-400 hover:underline">View Gateways →</a>
          </div>
        </sg-card>

        <!-- Transit Tunnels KPI -->
        <sg-card title="Transit Tunnels" subtitle="WireGuard / Segment 0">
          <div class="flex items-baseline justify-between mt-2">
            <span class="text-3xl font-bold font-mono text-indigo-400">{{ api.tunnels().length }}</span>
            <span class="text-xs text-indigo-300 font-mono">0.28ms latency</span>
          </div>
          <div card-footer class="flex justify-between items-center">
            <span>Mesh Backbone</span>
            <a routerLink="/tunnels" class="text-sky-400 hover:underline">View Tunnels →</a>
          </div>
        </sg-card>

        <!-- Health / Replacement KPI -->
        <sg-card title="Dataplane Health" subtitle="kube-proxy replacement">
          <div class="flex items-baseline justify-between mt-2">
            <span class="text-3xl font-bold font-mono text-emerald-400">{{ api.overviewStats().healthPercentage }}%</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 font-mono">sockops Fastpath</span>
          </div>
          <div card-footer class="flex justify-between items-center">
            <span>Uptime: {{ api.overviewStats().uptimeSeconds | duration }}</span>
            <a routerLink="/ebpf" class="text-sky-400 hover:underline">Inspect BPF →</a>
          </div>
        </sg-card>
      </div>

      <!-- Telemetry Charts & Gauges -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
        <sg-card title="Throughput & Forwarding" subtitle="Real-time eBPF packet rate">
          <div class="mt-2 space-y-4">
            <sg-timeseries-chart 
              [data]="[14200, 16800, 15400, 19200, 22100, 18450, 24100, 21800, 26400, 28900]"
              color="#38bdf8"
              label="Packet Rate"
              currentValue="28.9k pkts/sec" />
            <sg-bandwidth-meter 
              label="Segment 0 Backbone Utilization"
              current="3.8 Gbps"
              max="10 Gbps"
              [percentage]="38" />
          </div>
        </sg-card>

        <sg-card title="Policy Evaluation Engine" subtitle="Deterministic verdicts">
          <div class="mt-3 flex justify-center">
            <sg-donut-chart 
              [segments]="[
                { label: 'Allowed', value: 98.4, color: '#10b981' },
                { label: 'Denied', value: 1.4, color: '#f43f5e' },
                { label: 'Dropped', value: 0.2, color: '#f59e0b' }
              ]"
              centerText="98.4%" />
          </div>
          <div card-footer class="text-center">
            <span class="text-[11px] text-slate-400">Zero packet drops on fast-path NetKit</span>
          </div>
        </sg-card>

        <sg-card title="Service Load Balancer" subtitle="Maglev consistent hashing (128)">
          <div class="space-y-3 mt-2">
            <div class="flex items-center justify-between text-xs p-2 rounded-lg bg-slate-950/60 border border-slate-800">
              <span class="text-slate-400">Total VIPs / Services:</span>
              <span class="font-mono text-slate-100 font-bold">{{ api.services().length }}</span>
            </div>
            <div class="flex items-center justify-between text-xs p-2 rounded-lg bg-slate-950/60 border border-slate-800">
              <span class="text-slate-400">Active Endpoints:</span>
              <span class="font-mono text-emerald-400 font-bold">32 Healthy</span>
            </div>
            <div class="flex items-center justify-between text-xs p-2 rounded-lg bg-slate-950/60 border border-slate-800">
              <span class="text-slate-400">Conntrack LRU Entries:</span>
              <span class="font-mono text-indigo-400 font-bold">4,182</span>
            </div>
          </div>
        </sg-card>
      </div>

      <!-- Live Interactive Topology Map Preview -->
      <sg-card title="Live Multi-Cluster & Transit Topology" subtitle="Interactive D3.js Force Simulation">
        <div class="mt-3">
          <sg-d3-topology-graph />
        </div>
      </sg-card>

      <!-- Bottom Split: Real-time Flows & System Events -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <!-- Live Flows Stream -->
        <sg-card title="Live eBPF Flow Stream" subtitle="Kernel socket & NetKit events">
          <div class="space-y-2 mt-2">
            @for (flow of api.flows(); track flow.flowID) {
              <div class="flex items-center justify-between p-2.5 bg-slate-950/80 rounded-lg border border-slate-800/80 text-xs font-mono">
                <div class="flex items-center gap-2">
                  <span class="text-sky-400">{{ flow.srcIP }}:{{ flow.srcPort }}</span>
                  <span class="text-slate-500">→</span>
                  <span class="text-indigo-400">{{ flow.dstIP }}:{{ flow.dstPort }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <span class="px-1.5 py-0.5 rounded text-[10px] uppercase font-bold"
                        [ngClass]="flow.action === 'Allowed' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'">
                    {{ flow.action }}
                  </span>
                  <span class="text-[10px] text-slate-500">{{ flow.timestamp }}</span>
                </div>
              </div>
            }
          </div>
          <div card-footer>
            <a routerLink="/flows" class="text-sky-400 hover:underline">Explore full flow analytics →</a>
          </div>
        </sg-card>

        <!-- System & Transit Events -->
        <sg-card title="Unified Telemetry Events" subtitle="11-field metadata tuple tracking">
          <div class="space-y-2 mt-2">
            @for (ev of api.events(); track ev.id) {
              <div class="flex items-start justify-between p-2.5 bg-slate-950/80 rounded-lg border border-slate-800/80 text-xs">
                <div class="flex items-start gap-2">
                  <span class="text-[10px] px-1.5 py-0.5 rounded uppercase font-mono font-bold mt-0.5"
                        [ngClass]="ev.type === 'SUCCESS' ? 'bg-emerald-500/10 text-emerald-400' : (ev.type === 'WARN' ? 'bg-amber-500/10 text-amber-400' : 'bg-sky-500/10 text-sky-400')">
                    {{ ev.component }}
                  </span>
                  <p class="text-slate-200 text-xs leading-snug">{{ ev.message }}</p>
                </div>
                <span class="text-[10px] text-slate-500 font-mono shrink-0 ml-2">{{ ev.timestamp }}</span>
              </div>
            }
          </div>
          <div card-footer>
            <a routerLink="/events" class="text-sky-400 hover:underline">View all audit events →</a>
          </div>
        </sg-card>
      </div>
    </div>
  `
})
export class DashboardComponent {
  readonly api = inject(ApiClientService);
}

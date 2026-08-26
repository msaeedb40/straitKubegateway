import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '../../shared/components/card/card.component';
import { TimeseriesChartComponent } from '../../shared/charts/timeseries-chart.component';
import { BandwidthMeterComponent } from '../../shared/charts/bandwidth-meter.component';
import { DonutChartComponent } from '../../shared/charts/donut-chart.component';

@Component({
  selector: 'app-metrics',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    TimeseriesChartComponent,
    BandwidthMeterComponent,
    DonutChartComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Prometheus & Dataplane Metrics</h1>
          <p class="text-xs text-slate-400 mt-1">
            Real-time telemetry counters, latency distributions, and resource accounting
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs px-3 py-1 rounded-full bg-sky-500/10 border border-sky-500/20 text-sky-400 font-mono">
            /metrics endpoint active
          </span>
        </div>
      </div>

      <!-- Charts Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
        <sg-card title="Forwarding Throughput" subtitle="Packets per second on NetKit">
          <div class="mt-2">
            <sg-timeseries-chart 
              [data]="[21000, 24500, 22800, 28900, 31200, 29400, 34000, 32100, 36500, 38200]"
              color="#38bdf8"
              label="Current PPS"
              currentValue="38.2k pkts/s" />
          </div>
        </sg-card>

        <sg-card title="Fast-Path Forwarding Latency" subtitle="Kernel eBPF hop delay">
          <div class="mt-2">
            <sg-timeseries-chart 
              [data]="[0.22, 0.24, 0.21, 0.28, 0.23, 0.25, 0.24, 0.22, 0.21, 0.23]"
              color="#10b981"
              label="p99 Latency"
              currentValue="0.23 ms" />
          </div>
        </sg-card>

        <sg-card title="Memory Footprint per Node" subtitle="cgroup v2 resource control">
          <div class="mt-2 space-y-3">
            <sg-bandwidth-meter 
              label="straitd Daemon Memory"
              current="142 MB"
              max="512 MB"
              [percentage]="27" />
            <sg-bandwidth-meter 
              label="eBPF Maps Kernel Alloc"
              current="38 MB"
              max="256 MB"
              [percentage]="14" />
          </div>
        </sg-card>
      </div>

      <!-- Conntrack & Protocol Distribution -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
        <sg-card title="Protocol Distribution" subtitle="Active flows by transport layer">
          <div class="mt-4 flex justify-center">
            <sg-donut-chart 
              [segments]="[
                { label: 'TCP', value: 74, color: '#38bdf8' },
                { label: 'UDP', value: 22, color: '#818cf8' },
                { label: 'ICMP', value: 4, color: '#34d399' }
              ]"
              centerText="18.4k flows" />
          </div>
        </sg-card>

        <sg-card title="Conntrack Table Occupancy" subtitle="LRU Hash max 262,144 entries">
          <div class="mt-4 space-y-3">
            <sg-bandwidth-meter 
              label="Conntrack Capacity"
              current="4,182"
              max="262,144"
              [percentage]="2" />
            <div class="p-3 bg-slate-950 rounded-lg border border-slate-800 text-xs font-mono space-y-1 text-slate-300">
              <div>TCP Established: <strong class="text-sky-400">3,892</strong></div>
              <div>UDP Unreplied: <strong class="text-indigo-400">241</strong></div>
              <div>SYN_SENT: <strong class="text-amber-400">49</strong></div>
            </div>
          </div>
        </sg-card>
      </div>
    </div>
  `
})
export class MetricsComponent {}

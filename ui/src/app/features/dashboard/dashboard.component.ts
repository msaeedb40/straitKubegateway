import { Component, inject, ElementRef, viewChild, afterNextRender } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { StateService } from '../../core/state/state.service';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';
import { HealthIndicatorComponent } from '../../shared/components/health-indicator/health-indicator.component';
import * as d3 from 'd3';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, StatusBadgeComponent, HealthIndicatorComponent],
  template: `
    <div class="space-y-8 max-w-7xl mx-auto">
      <!-- Top Title Header -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Cluster Networking & Gateway Overview</h1>
          <p class="text-xs text-slate-400 mt-1">Real-time eBPF dataplane telemetry, Gateway API routes, and transit fabric status.</p>
        </div>
        <div class="flex items-center gap-3">
          <button (click)="state.refreshAll()" class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-xs font-medium text-slate-200 border border-slate-700 flex items-center gap-1.5 transition-all">
            <svg class="w-3.5 h-3.5 text-indigo-400 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"></path></svg>
            Sync Snapshot
          </button>
        </div>
      </div>

      <!-- Metric KPI Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Active Gateways -->
        <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl relative overflow-hidden group hover:border-indigo-500/50 transition-all">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">Active Gateways</span>
            <div class="p-2 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>
            </div>
          </div>
          <div class="mt-3 flex items-baseline gap-2">
            <span class="text-2xl font-bold font-mono text-white">{{ state.stats()?.activeGateways || 0 }}</span>
            <app-status-badge status="Ready">100% Ready</app-status-badge>
          </div>
          <p class="text-[11px] text-slate-500 mt-1">{{ state.totalRoutes() }} configured L7/L4 routes</p>
        </div>

        <!-- Network Policies -->
        <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl relative overflow-hidden group hover:border-sky-500/50 transition-all">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">Network Policies</span>
            <div class="p-2 rounded-lg bg-sky-500/10 text-sky-400 border border-sky-500/20">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
            </div>
          </div>
          <div class="mt-3 flex items-baseline gap-2">
            <span class="text-2xl font-bold font-mono text-white">{{ state.stats()?.activePolicies || 0 }}</span>
            <app-status-badge status="Active">Enforced</app-status-badge>
          </div>
          <p class="text-[11px] text-slate-500 mt-1">Ingress: Deny-all | Egress: Allow-all</p>
        </div>

        <!-- Dataplane Throughput -->
        <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl relative overflow-hidden group hover:border-emerald-500/50 transition-all">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">Traffic Throughput</span>
            <div class="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>
            </div>
          </div>
          <div class="mt-3 flex items-baseline gap-2">
            <span class="text-2xl font-bold font-mono text-white">{{ state.stats()?.rxThroughputMbps || 0 }}</span>
            <span class="text-xs text-slate-400">Mbps RX</span>
          </div>
          <p class="text-[11px] text-slate-500 mt-1">{{ state.stats()?.txThroughputMbps || 0 }} Mbps TX | {{ state.stats()?.activeFlowsPerSec || 0 }} flows/s</p>
        </div>

        <!-- Multi-Cluster Fabric -->
        <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl relative overflow-hidden group hover:border-purple-500/50 transition-all">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">Transit & Peering</span>
            <div class="p-2 rounded-lg bg-purple-500/10 text-purple-400 border border-purple-500/20">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
            </div>
          </div>
          <div class="mt-3 flex items-baseline gap-2">
            <span class="text-2xl font-bold font-mono text-white">{{ state.stats()?.establishedTunnels || 0 }}</span>
            <app-status-badge status="Established">WireGuard / IPsec</app-status-badge>
          </div>
          <p class="text-[11px] text-slate-500 mt-1">{{ state.stats()?.bgpPeersEstablished || 0 }} BGP Peers Established</p>
        </div>
      </div>

      <!-- D3 Interactive Throughput Chart & Top Gateways -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Live Throughput D3 Chart -->
        <div class="lg:col-span-2 p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
          <div class="flex items-center justify-between mb-4">
            <div>
              <h2 class="text-sm font-semibold text-white">Real-Time eBPF Packet Flow Rate</h2>
              <p class="text-[11px] text-slate-400">NetKit & TCX dataplane packet velocity (pps)</p>
            </div>
            <div class="flex items-center gap-4 text-xs font-mono">
              <span class="flex items-center gap-1.5 text-indigo-400"><span class="w-2 h-2 rounded-full bg-indigo-500"></span> Ingress</span>
              <span class="flex items-center gap-1.5 text-sky-400"><span class="w-2 h-2 rounded-full bg-sky-500"></span> Egress</span>
            </div>
          </div>
          <div #chartContainer class="h-64 w-full"></div>
        </div>

        <!-- Quick Status Summary -->
        <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-col justify-between">
          <div>
            <h2 class="text-sm font-semibold text-white mb-3">eBPF Dataplane Health</h2>
            <div class="space-y-3">
              <div class="flex items-center justify-between p-3 rounded-xl bg-slate-950/60 border border-slate-800">
                <app-health-indicator [isHealthy]="true" label="kube-proxy Replacement"></app-health-indicator>
                <span class="text-[11px] font-mono text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20">eBPF Native</span>
              </div>
              <div class="flex items-center justify-between p-3 rounded-xl bg-slate-950/60 border border-slate-800">
                <app-health-indicator [isHealthy]="true" label="Maglev Hash LUT"></app-health-indicator>
                <span class="text-[11px] font-mono text-indigo-400 bg-indigo-500/10 px-2 py-0.5 rounded-full border border-indigo-500/20">128 Slots</span>
              </div>
              <div class="flex items-center justify-between p-3 rounded-xl bg-slate-950/60 border border-slate-800">
                <app-health-indicator [isHealthy]="true" label="Segments Isolation"></app-health-indicator>
                <span class="text-[11px] font-mono text-purple-400 bg-purple-500/10 px-2 py-0.5 rounded-full border border-purple-500/20">32-bit (0..4B)</span>
              </div>
              <div class="flex items-center justify-between p-3 rounded-xl bg-slate-950/60 border border-slate-800">
                <app-health-indicator [isHealthy]="true" label="NAT64 Stateful IPv6"></app-health-indicator>
                <span class="text-[11px] font-mono text-sky-400 bg-sky-500/10 px-2 py-0.5 rounded-full border border-sky-500/20">64:ff9b::/96</span>
              </div>
            </div>
          </div>
          <div class="mt-4 pt-4 border-t border-slate-800/80 flex items-center justify-between text-xs">
            <span class="text-slate-400">Drop Rate:</span>
            <span class="font-mono text-emerald-400 font-semibold">{{ state.stats()?.dropRatePct }}% (Normal)</span>
          </div>
        </div>
      </div>

      <!-- Recent Audit Events -->
      <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-semibold text-white">Recent Gateway & Policy Events</h2>
          <a routerLink="/events" class="text-xs text-indigo-400 hover:text-indigo-300">View all events &rarr;</a>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="text-slate-400 border-b border-slate-800">
              <tr>
                <th class="pb-2 font-medium">Severity</th>
                <th class="pb-2 font-medium">Component</th>
                <th class="pb-2 font-medium">Message</th>
                <th class="pb-2 font-medium">Resource</th>
                <th class="pb-2 font-medium">Timestamp</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 text-slate-300">
              @for (ev of state.events(); track ev.id) {
                <tr class="hover:bg-slate-800/30 transition-colors">
                  <td class="py-3">
                    <app-status-badge [status]="ev.type">{{ ev.type }}</app-status-badge>
                  </td>
                  <td class="py-3 font-mono text-slate-400">{{ ev.component }}</td>
                  <td class="py-3 font-medium text-slate-200">{{ ev.message }}</td>
                  <td class="py-3 font-mono text-indigo-300">{{ ev.resourceRef }}</td>
                  <td class="py-3 text-slate-500 font-mono">{{ ev.timestamp | date:'HH:mm:ss' }}</td>
                </tr>
              }
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `
})
export class DashboardComponent {
  readonly state = inject(StateService);
  private readonly chartContainer = viewChild<ElementRef<HTMLDivElement>>('chartContainer');

  constructor() {
    afterNextRender(() => {
      this.initThroughputChart();
    });
  }

  private initThroughputChart(): void {
    const el = this.chartContainer()?.nativeElement;
    if (!el) return;

    d3.select(el).selectAll('*').remove();

    const width = el.clientWidth || 600;
    const height = el.clientHeight || 250;
    const margin = { top: 20, right: 20, bottom: 30, left: 50 };

    const svg = d3.select(el)
      .append('svg')
      .attr('width', width)
      .attr('height', height);

    const dataLength = 30;
    const ingressData = Array.from({ length: dataLength }, (_, i) => ({
      time: i,
      value: 1200 + Math.sin(i / 3) * 300 + Math.random() * 150
    }));

    const egressData = Array.from({ length: dataLength }, (_, i) => ({
      time: i,
      value: 950 + Math.cos(i / 3) * 250 + Math.random() * 100
    }));

    const x = d3.scaleLinear()
      .domain([0, dataLength - 1])
      .range([margin.left, width - margin.right]);

    const y = d3.scaleLinear()
      .domain([0, 2000])
      .range([height - margin.bottom, margin.top]);

    // Gridlines
    svg.append('g')
      .attr('stroke', '#334155')
      .attr('stroke-opacity', 0.3)
      .call(d3.axisLeft(y).tickSize(-width + margin.left + margin.right).tickFormat(() => ''));

    // Ingress Area
    const ingressArea = d3.area<{ time: number; value: number }>()
      .x(d => x(d.time))
      .y0(height - margin.bottom)
      .y1(d => y(d.value))
      .curve(d3.curveMonotoneX);

    svg.append('path')
      .datum(ingressData)
      .attr('fill', 'rgba(99, 102, 241, 0.15)')
      .attr('d', ingressArea);

    // Ingress Line
    const ingressLine = d3.line<{ time: number; value: number }>()
      .x(d => x(d.time))
      .y(d => y(d.value))
      .curve(d3.curveMonotoneX);

    svg.append('path')
      .datum(ingressData)
      .attr('fill', 'none')
      .attr('stroke', '#6366f1')
      .attr('stroke-width', 2)
      .attr('d', ingressLine);

    // Egress Line
    const egressLine = d3.line<{ time: number; value: number }>()
      .x(d => x(d.time))
      .y(d => y(d.value))
      .curve(d3.curveMonotoneX);

    svg.append('path')
      .datum(egressData)
      .attr('fill', 'none')
      .attr('stroke', '#38bdf8')
      .attr('stroke-width', 2)
      .attr('d', egressLine);

    // X Axis
    svg.append('g')
      .attr('transform', `translate(0,${height - margin.bottom})`)
      .attr('color', '#64748b')
      .call(d3.axisBottom(x).ticks(5).tickFormat(d => `-${dataLength - Number(d)}s`));

    // Y Axis
    svg.append('g')
      .attr('transform', `translate(${margin.left},0)`)
      .attr('color', '#64748b')
      .call(d3.axisLeft(y).ticks(4));
  }
}

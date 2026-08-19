import { Component, inject, ElementRef, viewChild, afterNextRender } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { SessionStore } from '../../core/state/session.store';
import * as d3 from 'd3';

@Component({
  selector: 'app-observability',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <!-- Title & Time Resolution Selector -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Observability & Distributed Tracing</h1>
          <p class="text-xs text-slate-400 mt-1">L4/L7 latency distribution, eBPF packet histograms, and Prometheus telemetry metrics.</p>
        </div>
        <!-- Multi-Resolution Time Selector -->
        <div class="flex items-center gap-1 p-1 rounded-xl bg-slate-900 border border-slate-800 text-xs">
          @for (range of timeRanges; track range) {
            <button
              (click)="setTimeRange(range)"
              class="px-2.5 py-1 rounded-lg font-mono transition-all"
              [ngClass]="sessionStore.timeRange() === range ? 'bg-indigo-600 text-white font-bold shadow-sm' : 'text-slate-400 hover:text-slate-200'"
            >
              {{ range }}
            </button>
          }
        </div>
      </div>

      <!-- Latency Percentiles KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
        <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col">
          <span class="text-xs text-slate-400 font-mono">P50 Latency</span>
          <span class="text-xl font-bold font-mono text-white mt-1">45 µs</span>
          <span class="text-[10px] text-emerald-400 mt-0.5">&plusmn; 2 µs jitter</span>
        </div>
        <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col">
          <span class="text-xs text-slate-400 font-mono">P90 Latency</span>
          <span class="text-xl font-bold font-mono text-white mt-1">120 µs</span>
          <span class="text-[10px] text-emerald-400 mt-0.5">Optimal</span>
        </div>
        <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col">
          <span class="text-xs text-slate-400 font-mono">P95 Latency</span>
          <span class="text-xl font-bold font-mono text-indigo-300 mt-1">240 µs</span>
          <span class="text-[10px] text-indigo-400 mt-0.5">SLO Met</span>
        </div>
        <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col">
          <span class="text-xs text-slate-400 font-mono">P99 Latency</span>
          <span class="text-xl font-bold font-mono text-amber-300 mt-1">580 µs</span>
          <span class="text-[10px] text-amber-400 mt-0.5">&lt; 1.0 ms Target</span>
        </div>
        <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col">
          <span class="text-xs text-slate-400 font-mono">P99.9 Latency</span>
          <span class="text-xl font-bold font-mono text-purple-300 mt-1">1.12 ms</span>
          <span class="text-[10px] text-purple-400 mt-0.5">Max Tail</span>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Latency Percentiles Chart -->
        <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
          <div class="flex items-center justify-between mb-4">
            <div>
              <h2 class="text-sm font-semibold text-white">L4/L7 Latency Distribution</h2>
              <p class="text-[11px] text-slate-400">Resolution window: {{ sessionStore.timeRange() }}</p>
            </div>
            <span class="text-xs text-slate-400 font-mono">Microseconds (µs)</span>
          </div>
          <div #latencyChart class="h-64 w-full"></div>
        </div>

        <!-- Request Latency Breakdown -->
        <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-col justify-between">
          <div>
            <h2 class="text-sm font-semibold text-white mb-1">Request Latency Breakdown</h2>
            <p class="text-[11px] text-slate-400 mb-4">Detailed phase-by-phase execution duration</p>
            <div class="space-y-2.5 text-xs">
              <div class="flex items-center justify-between p-2.5 rounded-lg bg-slate-950/60 border border-slate-800">
                <span class="text-slate-400">eBPF TCX / NetKit Dataplane:</span>
                <span class="font-mono text-emerald-400 font-bold">18 µs</span>
              </div>
              <div class="flex items-center justify-between p-2.5 rounded-lg bg-slate-950/60 border border-slate-800">
                <span class="text-slate-400">Maglev Hash LUT Evaluation:</span>
                <span class="font-mono text-emerald-400 font-bold">12 µs</span>
              </div>
              <div class="flex items-center justify-between p-2.5 rounded-lg bg-slate-950/60 border border-slate-800">
                <span class="text-slate-400">Network Policy Filter Rule Match:</span>
                <span class="font-mono text-sky-400 font-bold">25 µs</span>
              </div>
              <div class="flex items-center justify-between p-2.5 rounded-lg bg-slate-950/60 border border-slate-800">
                <span class="text-slate-400">Gateway TLS Handshake / Termination:</span>
                <span class="font-mono text-indigo-300 font-bold">185 µs</span>
              </div>
              <div class="flex items-center justify-between p-2.5 rounded-lg bg-slate-950/60 border border-slate-800">
                <span class="text-slate-400">sg-controller Reconciliation Latency:</span>
                <span class="font-mono text-purple-300 font-bold">340 µs</span>
              </div>
            </div>
          </div>
          <div class="mt-4 pt-4 border-t border-slate-800/80 flex items-center justify-between text-[11px] text-slate-500">
            <span>Trace sampling: 100% Errors / 1% Normal</span>
            <code class="text-indigo-400 font-mono">GET /metrics</code>
          </div>
        </div>
      </div>
    </div>
  `
})
export class ObservabilityComponent {
  readonly state = inject(StateService);
  readonly sessionStore = inject(SessionStore);
  private readonly latencyChart = viewChild<ElementRef<HTMLDivElement>>('latencyChart');

  readonly timeRanges: ('5s' | '5m' | '24h' | '7d' | '30d' | '90d')[] = ['5s', '5m', '24h', '7d', '30d', '90d'];

  constructor() {
    afterNextRender(() => {
      this.initLatencyChart();
    });
  }

  setTimeRange(range: '5s' | '5m' | '24h' | '7d' | '30d' | '90d'): void {
    this.sessionStore.setTimeRange(range);
    this.initLatencyChart();
  }

  private initLatencyChart(): void {
    const el = this.latencyChart()?.nativeElement;
    if (!el) return;

    d3.select(el).selectAll('*').remove();

    const width = el.clientWidth || 500;
    const height = el.clientHeight || 250;
    const margin = { top: 20, right: 20, bottom: 30, left: 40 };

    const svg = d3.select(el)
      .append('svg')
      .attr('width', width)
      .attr('height', height);

    const categories = ['P50', 'P90', 'P95', 'P99', 'P99.9'];
    const values = [45, 120, 240, 580, 1120];

    const x = d3.scaleBand()
      .domain(categories)
      .range([margin.left, width - margin.right])
      .padding(0.4);

    const y = d3.scaleLinear()
      .domain([0, 1300])
      .range([height - margin.bottom, margin.top]);

    // Bars
    svg.selectAll('.bar')
      .data(values)
      .enter()
      .append('rect')
      .attr('class', 'bar')
      .attr('x', (_, i) => x(categories[i]) || 0)
      .attr('y', d => y(d))
      .attr('width', x.bandwidth())
      .attr('height', d => height - margin.bottom - y(d))
      .attr('rx', 6)
      .attr('fill', (d, i) => i >= 3 ? '#a855f7' : '#6366f1');

    // Values on top of bars
    svg.selectAll('.label')
      .data(values)
      .enter()
      .append('text')
      .attr('x', (_, i) => (x(categories[i]) || 0) + x.bandwidth() / 2)
      .attr('y', d => y(d) - 6)
      .attr('text-anchor', 'middle')
      .attr('fill', '#cbd5e1')
      .attr('font-size', '10px')
      .attr('font-family', 'monospace')
      .text(d => `${d}µs`);

    // X Axis
    svg.append('g')
      .attr('transform', `translate(0,${height - margin.bottom})`)
      .attr('color', '#64748b')
      .call(d3.axisBottom(x));

    // Y Axis
    svg.append('g')
      .attr('transform', `translate(${margin.left},0)`)
      .attr('color', '#64748b')
      .call(d3.axisLeft(y).ticks(4));
  }
}

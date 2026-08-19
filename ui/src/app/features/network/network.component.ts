import { Component, inject, ElementRef, viewChild, afterNextRender, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { DataTableComponent, TableColumn } from '../../shared/components/data-table/data-table.component';
import * as d3 from 'd3';

interface TopologyNode extends d3.SimulationNodeDatum {
  id: string;
  name: string;
  type: 'gateway' | 'node' | 'segment' | 'transit';
  group: number;
}

interface TopologyLink extends d3.SimulationLinkDatum<TopologyNode> {
  source: string | TopologyNode;
  target: string | TopologyNode;
  value: number;
}

@Component({
  selector: 'app-network',
  standalone: true,
  imports: [CommonModule, DataTableComponent],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <!-- Title & View Toggle -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Network Topology & Segments</h1>
          <p class="text-xs text-slate-400 mt-1">Multi-tenant 32-bit segment isolation, Geneve VNI mappings, and cluster mesh topology.</p>
        </div>
        <div class="flex items-center gap-2 p-1 rounded-xl bg-slate-900 border border-slate-800 text-xs">
          <button
            (click)="activeView.set('graph')"
            class="px-3 py-1.5 rounded-lg transition-all"
            [ngClass]="activeView() === 'graph' ? 'bg-indigo-600 text-white font-medium shadow-md shadow-indigo-600/30' : 'text-slate-400 hover:text-slate-200'"
          >
            D3 Visual Graph
          </button>
          <button
            (click)="activeView.set('table')"
            class="px-3 py-1.5 rounded-lg transition-all"
            [ngClass]="activeView() === 'table' ? 'bg-indigo-600 text-white font-medium shadow-md shadow-indigo-600/30' : 'text-slate-400 hover:text-slate-200'"
          >
            Accessible Table View (WCAG)
          </button>
        </div>
      </div>

      <!-- Graph View Mode -->
      @if (activeView() === 'graph') {
        <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-3">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-semibold text-white">Interactive Network Fabric Mesh</h2>
            <div class="flex items-center gap-4 text-xs font-mono">
              <span class="flex items-center gap-1.5 text-indigo-400"><span class="w-2.5 h-2.5 rounded-full bg-indigo-500"></span> Gateways</span>
              <span class="flex items-center gap-1.5 text-emerald-400"><span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span> Nodes</span>
              <span class="flex items-center gap-1.5 text-purple-400"><span class="w-2.5 h-2.5 rounded-full bg-purple-500"></span> Segments</span>
            </div>
          </div>
          <div #topologyContainer class="h-96 w-full rounded-xl bg-slate-950/60 border border-slate-800/80 overflow-hidden relative"></div>
          <p class="text-[11px] text-slate-500">Click and drag nodes to inspect connection links and segment boundaries.</p>
        </div>
      }

      <!-- Accessible Table Fallback Mode (WCAG Compliant) -->
      @if (activeView() === 'table') {
        <div class="space-y-6">
          <app-data-table
            title="Configured Segments & Subnets"
            subtitle="Accessible tabular summary of all network isolation boundaries"
            [columns]="segmentColumns"
            [data]="state.segments()"
          ></app-data-table>
        </div>
      }

      <!-- Segments Summary Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        @for (seg of state.segments(); track seg.id) {
          <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-3 hover:border-purple-500/40 transition-all">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold px-2 py-0.5 rounded-md bg-purple-500/10 text-purple-400 border border-purple-500/20 font-mono">
                Segment {{ seg.id }}
              </span>
              <span class="text-xs font-mono text-slate-400">VNI: {{ seg.vni }}</span>
            </div>
            <h3 class="text-base font-semibold text-white">{{ seg.name }}</h3>
            <div class="text-xs text-slate-400 space-y-1">
              <div>Endpoints: <span class="font-mono text-slate-200">{{ seg.endpointsCount }}</span></div>
              <div>Subnets: <span class="font-mono text-indigo-300">{{ seg.subnets.join(', ') }}</span></div>
            </div>
          </div>
        }
      </div>
    </div>
  `
})
export class NetworkComponent {
  readonly state = inject(StateService);
  readonly activeView = signal<'graph' | 'table'>('graph');
  private readonly topologyContainer = viewChild<ElementRef<HTMLDivElement>>('topologyContainer');

  readonly segmentColumns: TableColumn[] = [
    { key: 'id', header: 'Segment ID', sortable: true, mono: true, width: '120px' },
    { key: 'name', header: 'Segment Name', sortable: true },
    { key: 'vni', header: 'Geneve VNI', sortable: true, mono: true },
    { key: 'endpointsCount', header: 'Active Endpoints', sortable: true, mono: true, align: 'center' },
    { key: 'isolated', header: 'Isolation Enforced', sortable: true }
  ];

  constructor() {
    afterNextRender(() => {
      this.initTopologyGraph();
    });
  }

  private initTopologyGraph(): void {
    const el = this.topologyContainer()?.nativeElement;
    if (!el) return;

    d3.select(el).selectAll('*').remove();

    const width = el.clientWidth || 800;
    const height = el.clientHeight || 384;

    const svg = d3.select(el)
      .append('svg')
      .attr('width', width)
      .attr('height', height)
      .attr('viewBox', [0, 0, width, height]);

    const nodes: TopologyNode[] = [
      { id: 'gw-01', name: 'edge-gateway-prod', type: 'gateway', group: 1 },
      { id: 'gw-02', name: 'internal-api-gateway', type: 'gateway', group: 1 },
      { id: 'node-01', name: 'worker-01 (eBPF)', type: 'node', group: 2 },
      { id: 'node-02', name: 'worker-02 (eBPF)', type: 'node', group: 2 },
      { id: 'seg-0', name: 'Global Segment (0)', type: 'segment', group: 3 },
      { id: 'seg-100', name: 'Prod Segment (100)', type: 'segment', group: 3 },
      { id: 'tgw-01', name: 'Transit Mesh Hub', type: 'transit', group: 4 }
    ];

    const links: TopologyLink[] = [
      { source: 'gw-01', target: 'seg-0', value: 2 },
      { source: 'gw-02', target: 'seg-100', value: 2 },
      { source: 'gw-01', target: 'node-01', value: 1 },
      { source: 'gw-02', target: 'node-02', value: 1 },
      { source: 'node-01', target: 'tgw-01', value: 3 },
      { source: 'node-02', target: 'tgw-01', value: 3 }
    ];

    const simulation = d3.forceSimulation<TopologyNode>(nodes)
      .force('link', d3.forceLink<TopologyNode, TopologyLink>(links).id(d => d.id).distance(90))
      .force('charge', d3.forceManyBody().strength(-200))
      .force('center', d3.forceCenter(width / 2, height / 2));

    const link = svg.append('g')
      .attr('stroke', '#475569')
      .attr('stroke-opacity', 0.6)
      .selectAll('line')
      .data(links)
      .join('line')
      .attr('stroke-width', d => Math.sqrt(d.value) * 1.5);

    const node = svg.append('g')
      .selectAll('g')
      .data(nodes)
      .join('g')
      .call(d3.drag<any, TopologyNode>()
        .on('start', (event, d) => {
          if (!event.active) simulation.alphaTarget(0.3).restart();
          d.fx = d.x;
          d.fy = d.y;
        })
        .on('drag', (event, d) => {
          d.fx = event.x;
          d.fy = event.y;
        })
        .on('end', (event, d) => {
          if (!event.active) simulation.alphaTarget(0);
          d.fx = null;
          d.fy = null;
        }));

    // Circle color per group
    const color = (type: string) => {
      switch (type) {
        case 'gateway': return '#6366f1';
        case 'node': return '#10b981';
        case 'segment': return '#a855f7';
        case 'transit': return '#38bdf8';
        default: return '#94a3b8';
      }
    };

    node.append('circle')
      .attr('r', 12)
      .attr('fill', d => color(d.type))
      .attr('stroke', '#0f172a')
      .attr('stroke-width', 2);

    node.append('text')
      .attr('x', 16)
      .attr('y', 4)
      .text(d => d.name)
      .attr('fill', '#f1f5f9')
      .attr('font-size', '10px')
      .attr('font-family', 'sans-serif');

    simulation.on('tick', () => {
      link
        .attr('x1', d => (d.source as TopologyNode).x!)
        .attr('y1', d => (d.source as TopologyNode).y!)
        .attr('x2', d => (d.target as TopologyNode).x!)
        .attr('y2', d => (d.target as TopologyNode).y!);

      node
        .attr('transform', d => `translate(${d.x},${d.y})`);
    });
  }
}

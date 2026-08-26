import { Component, ElementRef, ViewChild, AfterViewInit, OnDestroy, input, output, inject, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import * as d3 from 'd3';

export interface TopologyNode extends d3.SimulationNodeDatum {
  id: string;
  name: string;
  type: 'gateway' | 'node' | 'service' | 'tunnel' | 'pod';
  status: 'active' | 'warning' | 'error';
  ip?: string;
  segment?: number;
}

export interface TopologyLink extends d3.SimulationLinkDatum<TopologyNode> {
  source: string | TopologyNode;
  target: string | TopologyNode;
  type: 'wireguard' | 'vxlan' | 'service' | 'netkit';
  active?: boolean;
}

@Component({
  selector: 'sg-d3-topology-graph',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="relative w-full h-[540px] bg-slate-950/80 rounded-xl border border-slate-800/80 overflow-hidden backdrop-blur-md">
      <!-- Toolbar controls -->
      <div class="absolute top-4 right-4 z-10 flex items-center gap-2 bg-slate-900/90 border border-slate-800 rounded-lg p-1.5 backdrop-blur-md shadow-lg">
        <button 
          (click)="zoomIn()" 
          title="Zoom In"
          class="w-7 h-7 flex items-center justify-center rounded bg-slate-800 hover:bg-slate-700 text-slate-200 font-bold transition text-xs">
          +
        </button>
        <button 
          (click)="zoomOut()" 
          title="Zoom Out"
          class="w-7 h-7 flex items-center justify-center rounded bg-slate-800 hover:bg-slate-700 text-slate-200 font-bold transition text-xs">
          −
        </button>
        <button 
          (click)="resetZoom()" 
          title="Reset Zoom"
          class="px-2 h-7 flex items-center justify-center rounded bg-slate-800 hover:bg-slate-700 text-slate-200 text-[11px] font-medium transition">
          Reset
        </button>
      </div>

      <!-- Legend overlay -->
      <div class="absolute bottom-4 left-4 z-10 flex flex-wrap items-center gap-3 bg-slate-900/90 border border-slate-800/80 rounded-lg px-3 py-2 text-[11px] text-slate-300 backdrop-blur-md">
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-indigo-500 shadow-sm shadow-indigo-500/50"></span>
          <span>Gateway</span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-sky-400 shadow-sm shadow-sky-400/50"></span>
          <span>Node</span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 shadow-sm shadow-emerald-400/50"></span>
          <span>Service</span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-purple-400 shadow-sm shadow-purple-400/50"></span>
          <span>Transit Tunnel</span>
        </div>
      </div>

      <!-- SVG Canvas -->
      <svg #svgCanvas class="w-full h-full cursor-grab active:cursor-grabbing select-none"></svg>
    </div>
  `
})
export class D3TopologyGraphComponent implements AfterViewInit, OnDestroy {
  @ViewChild('svgCanvas', { static: true }) svgCanvas!: ElementRef<SVGSVGElement>;

  readonly nodes = input<TopologyNode[]>([
    { id: 'gw-1', name: 'strait-transit-gw', type: 'gateway', status: 'active', ip: '10.0.0.1', segment: 0 },
    { id: 'node-cp', name: 'sg-control-plane', type: 'node', status: 'active', ip: '192.168.1.10' },
    { id: 'node-w1', name: 'sg-worker-01', type: 'node', status: 'active', ip: '192.168.1.11' },
    { id: 'node-w2', name: 'sg-worker-02', type: 'node', status: 'active', ip: '192.168.1.12' },
    { id: 'svc-dns', name: 'kube-dns', type: 'service', status: 'active', ip: '10.96.0.10' },
    { id: 'svc-pay', name: 'payment-service', type: 'service', status: 'active', ip: '10.96.142.18' },
    { id: 'tun-cl-b', name: 'cluster-b-mesh', type: 'tunnel', status: 'active', ip: '198.51.100.25' }
  ]);

  readonly links = input<TopologyLink[]>([
    { source: 'gw-1', target: 'node-cp', type: 'netkit', active: true },
    { source: 'gw-1', target: 'node-w1', type: 'netkit', active: true },
    { source: 'gw-1', target: 'node-w2', type: 'netkit', active: true },
    { source: 'node-w1', target: 'svc-pay', type: 'service', active: true },
    { source: 'node-w2', target: 'svc-pay', type: 'service', active: true },
    { source: 'node-cp', target: 'svc-dns', type: 'service', active: true },
    { source: 'gw-1', target: 'tun-cl-b', type: 'wireguard', active: true }
  ]);

  readonly nodeSelected = output<TopologyNode>();

  private readonly platformId = inject(PLATFORM_ID);
  private simulation?: d3.Simulation<TopologyNode, TopologyLink>;
  private zoomBehavior?: d3.ZoomBehavior<SVGSVGElement, unknown>;
  private gContainer?: d3.Selection<SVGGElement, unknown, null, undefined>;

  ngAfterViewInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      this.initGraph();
    }
  }

  ngOnDestroy(): void {
    this.simulation?.stop();
  }

  private initGraph(): void {
    const svg = d3.select(this.svgCanvas.nativeElement);
    svg.selectAll('*').remove();

    const width = this.svgCanvas.nativeElement.clientWidth || 800;
    const height = this.svgCanvas.nativeElement.clientHeight || 540;

    this.gContainer = svg.append('g').attr('class', 'topology-viewport');

    this.zoomBehavior = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.3, 3])
      .on('zoom', (event) => {
        this.gContainer?.attr('transform', event.transform);
      });

    svg.call(this.zoomBehavior);

    const nodesData = JSON.parse(JSON.stringify(this.nodes())) as TopologyNode[];
    const linksData = JSON.parse(JSON.stringify(this.links())) as TopologyLink[];

    this.simulation = d3.forceSimulation<TopologyNode>(nodesData)
      .force('link', d3.forceLink<TopologyNode, TopologyLink>(linksData).id((d) => d.id).distance(110))
      .force('charge', d3.forceManyBody().strength(-380))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(36));

    // Render Links
    const link = this.gContainer.append('g')
      .attr('class', 'links')
      .selectAll('line')
      .data(linksData)
      .enter()
      .append('line')
      .attr('stroke', (d) => d.type === 'wireguard' ? '#a855f7' : (d.type === 'service' ? '#10b981' : '#38bdf8'))
      .attr('stroke-width', 1.8)
      .attr('stroke-dasharray', (d) => d.type === 'wireguard' ? '4 3' : null)
      .attr('stroke-opacity', 0.65);

    // Render Nodes Group
    const nodeGroup = this.gContainer.append('g')
      .attr('class', 'nodes')
      .selectAll('g')
      .data(nodesData)
      .enter()
      .append('g')
      .attr('class', 'node cursor-pointer')
      .on('click', (_, d) => this.nodeSelected.emit(d))
      .call(
        d3.drag<SVGGElement, TopologyNode>()
          .on('start', (event, d) => {
            if (!event.active) this.simulation?.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on('drag', (event, d) => {
            d.fx = event.x;
            d.fy = event.y;
          })
          .on('end', (event, d) => {
            if (!event.active) this.simulation?.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          })
      );

    // Outer glow circle
    nodeGroup.append('circle')
      .attr('r', 20)
      .attr('fill', (d) => this.getNodeColor(d.type))
      .attr('fill-opacity', 0.15)
      .attr('stroke', (d) => this.getNodeColor(d.type))
      .attr('stroke-width', 1.5);

    // Core icon circle
    nodeGroup.append('circle')
      .attr('r', 12)
      .attr('fill', (d) => this.getNodeColor(d.type))
      .attr('stroke', '#0f172a')
      .attr('stroke-width', 2);

    // Node Labels
    nodeGroup.append('text')
      .text((d) => d.name)
      .attr('y', 28)
      .attr('text-anchor', 'middle')
      .attr('fill', '#e2e8f0')
      .attr('font-size', '10px')
      .attr('font-family', 'ui-monospace, monospace')
      .attr('font-weight', '500');

    // IP label
    nodeGroup.append('text')
      .text((d) => d.ip || '')
      .attr('y', 39)
      .attr('text-anchor', 'middle')
      .attr('fill', '#94a3b8')
      .attr('font-size', '8.5px')
      .attr('font-family', 'ui-monospace, monospace');

    // Simulation Tick
    this.simulation.on('tick', () => {
      link
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y);

      nodeGroup.attr('transform', (d) => `translate(${d.x}, ${d.y})`);
    });
  }

  private getNodeColor(type: string): string {
    switch (type) {
      case 'gateway': return '#6366f1';
      case 'node': return '#38bdf8';
      case 'service': return '#10b981';
      case 'tunnel': return '#c084fc';
      default: return '#94a3b8';
    }
  }

  zoomIn(): void {
    if (this.zoomBehavior) {
      d3.select(this.svgCanvas.nativeElement).transition().call(this.zoomBehavior.scaleBy, 1.3);
    }
  }

  zoomOut(): void {
    if (this.zoomBehavior) {
      d3.select(this.svgCanvas.nativeElement).transition().call(this.zoomBehavior.scaleBy, 0.77);
    }
  }

  resetZoom(): void {
    if (this.zoomBehavior) {
      d3.select(this.svgCanvas.nativeElement).transition().call(this.zoomBehavior.transform, d3.zoomIdentity);
    }
  }
}

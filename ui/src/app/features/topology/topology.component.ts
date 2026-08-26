import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '../../shared/components/card/card.component';
import { BadgeComponent } from '../../shared/components/badge/badge.component';
import { D3TopologyGraphComponent, TopologyNode } from '../../shared/topology/d3-topology-graph.component';

@Component({
  selector: 'app-topology',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    BadgeComponent,
    D3TopologyGraphComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Multi-Cluster & Gateway Topology</h1>
          <p class="text-xs text-slate-400 mt-1">
            Dynamic D3 force simulation of Gateways, Nodes, Transit Mesh Tunnels, and Services
          </p>
        </div>
      </div>

      <!-- Graph & Details Split Layout -->
      <div class="grid grid-cols-1 lg:grid-cols-4 gap-5">
        <div class="lg:col-span-3">
          <sg-card title="Interactive Graph View" subtitle="Drag nodes, scroll to zoom, click to inspect details">
            <sg-d3-topology-graph (nodeSelected)="selectedNode.set($event)" />
          </sg-card>
        </div>

        <!-- Node Inspector Drawer -->
        <div class="lg:col-span-1">
          <sg-card title="Resource Inspector" subtitle="Selected topology element">
            @if (selectedNode()) {
              <div class="space-y-4 mt-2 text-xs">
                <div class="p-3 rounded-lg bg-slate-950/80 border border-slate-800 space-y-2">
                  <div class="flex justify-between items-center">
                    <span class="text-slate-400 text-[11px]">Type:</span>
                    <sg-badge variant="primary">{{ selectedNode()?.type | uppercase }}</sg-badge>
                  </div>
                  <div class="flex justify-between items-center">
                    <span class="text-slate-400 text-[11px]">Name:</span>
                    <span class="font-mono text-slate-100 font-bold">{{ selectedNode()?.name }}</span>
                  </div>
                  <div class="flex justify-between items-center">
                    <span class="text-slate-400 text-[11px]">IP Address:</span>
                    <span class="font-mono text-sky-400 font-bold">{{ selectedNode()?.ip || '-' }}</span>
                  </div>
                  @if (selectedNode()?.segment !== undefined) {
                    <div class="flex justify-between items-center">
                      <span class="text-slate-400 text-[11px]">Segment:</span>
                      <span class="font-mono text-amber-400 font-bold">Segment {{ selectedNode()?.segment }}</span>
                    </div>
                  }
                  <div class="flex justify-between items-center">
                    <span class="text-slate-400 text-[11px]">Status:</span>
                    <span class="font-mono text-emerald-400 font-semibold">{{ selectedNode()?.status | uppercase }}</span>
                  </div>
                </div>

                <p class="text-[11px] text-slate-400 leading-relaxed">
                  Traffic forwarding to this entity is handled directly by kernel eBPF fastpath with zero userspace context switching.
                </p>
              </div>
            } @else {
              <div class="py-12 text-center text-slate-500 text-xs italic">
                Click any node in the topology canvas to inspect metadata
              </div>
            }
          </sg-card>
        </div>
      </div>
    </div>
  `
})
export class TopologyComponent {
  readonly selectedNode = signal<TopologyNode | null>(null);
}

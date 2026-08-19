import { Component, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { FlowItem } from '../../core/models/models';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';
import { VirtualListComponent } from '../../shared/components/virtual-list/virtual-list.component';
import { FormatBytesPipe } from '../../shared/pipes/format-bytes.pipe';

@Component({
  selector: 'app-flows',
  standalone: true,
  imports: [CommonModule, StatusBadgeComponent, VirtualListComponent, FormatBytesPipe],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <!-- Title & Live Stream Controls -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Live Flow Observer</h1>
          <p class="text-xs text-slate-400 mt-1">Real-time eBPF ring-buffer network packet stream with identity tracking and drop analysis.</p>
        </div>
        <div class="flex items-center gap-3">
          <button
            (click)="toggleStream()"
            class="px-3.5 py-1.5 rounded-xl text-xs font-semibold border transition-all flex items-center gap-1.5"
            [ngClass]="isStreaming() ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 hover:bg-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border-amber-500/30 hover:bg-amber-500/20'"
          >
            <span class="w-2 h-2 rounded-full" [ngClass]="isStreaming() ? 'bg-emerald-400 animate-ping' : 'bg-amber-400'"></span>
            {{ isStreaming() ? 'Live Ingest Active' : 'Stream Paused' }}
          </button>
        </div>
      </div>

      <!-- Flow Filter Bar -->
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-wrap items-center gap-4 text-xs">
        <div class="flex items-center gap-2">
          <span class="text-slate-400">Verdict:</span>
          <select
            [value]="selectedVerdict()"
            (change)="onVerdictChange($event)"
            class="bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1 text-slate-200 focus:outline-none"
          >
            <option value="ALL">All Verdicts</option>
            <option value="FORWARDED">FORWARDED</option>
            <option value="DROPPED">DROPPED</option>
            <option value="REDIRECTED">REDIRECTED</option>
          </select>
        </div>

        <div class="flex items-center gap-2">
          <span class="text-slate-400">Direction:</span>
          <select
            [value]="selectedDirection()"
            (change)="onDirectionChange($event)"
            class="bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1 text-slate-200 focus:outline-none"
          >
            <option value="ALL">All Directions</option>
            <option value="Ingress">Ingress</option>
            <option value="Egress">Egress</option>
          </select>
        </div>

        <div class="ml-auto text-slate-400 font-mono">
          Showing <span class="text-indigo-400 font-bold">{{ filteredFlows().length }}</span> / {{ state.flows().length }} flows
        </div>
      </div>

      <!-- High-Performance Virtualized Flow Table -->
      <app-virtual-list
        [items]="filteredFlows()"
        [itemSize]="48"
        viewportHeight="520px"
        headerTitle="eBPF Packet Ring Buffer (Virtual Scroller)"
      >
        <ng-template #itemTemplate let-f>
          <div class="w-full grid grid-cols-12 gap-2 text-xs font-mono items-center">
            <div class="col-span-2">
              <app-status-badge [status]="f.verdict">{{ f.verdict }}</app-status-badge>
            </div>
            <div class="col-span-1 text-slate-400">{{ f.direction }}</div>
            <div class="col-span-3 text-slate-200 truncate">{{ f.srcIP }}:{{ f.srcPort }} <span class="text-slate-500">[{{ f.protocol }}]</span></div>
            <div class="col-span-3 text-slate-200 truncate">&rarr; {{ f.dstIP }}:{{ f.dstPort }}</div>
            <div class="col-span-1 text-indigo-300 text-[11px] truncate">ID:{{ f.srcIdentity }}</div>
            <div class="col-span-1 text-slate-400">{{ f.bytes | formatBytes }}</div>
            <div class="col-span-1 text-right text-slate-500 text-[10px]">{{ f.timestamp | date:'HH:mm:ss' }}</div>
          </div>
        </ng-template>
      </app-virtual-list>
    </div>
  `
})
export class FlowsComponent {
  readonly state = inject(StateService);
  readonly isStreaming = signal<boolean>(true);
  readonly selectedVerdict = signal<string>('ALL');
  readonly selectedDirection = signal<string>('ALL');

  readonly filteredFlows = computed(() => {
    let list = this.state.flows();
    const v = this.selectedVerdict();
    const d = this.selectedDirection();

    if (v !== 'ALL') {
      list = list.filter(f => f.verdict === v);
    }
    if (d !== 'ALL') {
      list = list.filter(f => f.direction === d);
    }
    return list;
  });

  toggleStream(): void {
    this.isStreaming.update(s => !s);
  }

  onVerdictChange(event: Event): void {
    const val = (event.target as HTMLSelectElement).value;
    this.selectedVerdict.set(val);
  }

  onDirectionChange(event: Event): void {
    const val = (event.target as HTMLSelectElement).value;
    this.selectedDirection.set(val);
  }
}

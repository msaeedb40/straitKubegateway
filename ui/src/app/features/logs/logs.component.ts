import { Component, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '../../shared/components/card/card.component';
import { SearchInputComponent } from '../../shared/components/search-input/search-input.component';

interface LogEntry {
  id: string;
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  component: string;
  msg: string;
  traceID?: string;
  timestamp: string;
}

@Component({
  selector: 'app-logs',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    SearchInputComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Structured Telemetry Logs</h1>
          <p class="text-xs text-slate-400 mt-1">
            Real-time structured logs from sg-controller, straitd, and compiler workers
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button 
            (click)="selectedLevel.set('ALL')"
            class="px-2.5 py-1 text-xs rounded-lg transition cursor-pointer"
            [ngClass]="selectedLevel() === 'ALL' ? 'bg-sky-500 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'">
            ALL
          </button>
          <button 
            (click)="selectedLevel.set('INFO')"
            class="px-2.5 py-1 text-xs rounded-lg transition cursor-pointer"
            [ngClass]="selectedLevel() === 'INFO' ? 'bg-sky-500 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'">
            INFO
          </button>
          <button 
            (click)="selectedLevel.set('WARN')"
            class="px-2.5 py-1 text-xs rounded-lg transition cursor-pointer"
            [ngClass]="selectedLevel() === 'WARN' ? 'bg-amber-500 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'">
            WARN
          </button>
          <button 
            (click)="selectedLevel.set('ERROR')"
            class="px-2.5 py-1 text-xs rounded-lg transition cursor-pointer"
            [ngClass]="selectedLevel() === 'ERROR' ? 'bg-rose-500 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'">
            ERROR
          </button>
        </div>
      </div>

      <!-- Logs Terminal Viewer -->
      <sg-card title="Live Log Terminal" subtitle="Standard output from straitd daemon and controllers">
        <div card-action class="w-64">
          <sg-search-input placeholder="Filter log text..." (searchChange)="searchQuery.set($event)" />
        </div>

        <div class="mt-3 bg-slate-950 p-4 rounded-xl border border-slate-800/80 font-mono text-xs max-h-[500px] overflow-y-auto space-y-2 select-text">
          @for (log of filteredLogs(); track log.id) {
            <div class="flex items-start gap-3 hover:bg-slate-900/60 p-1.5 rounded transition">
              <span class="text-slate-500 text-[11px] shrink-0">{{ log.timestamp }}</span>
              <span class="px-1.5 py-0.5 rounded text-[10px] uppercase font-bold shrink-0"
                    [ngClass]="levelClass(log.level)">
                {{ log.level }}
              </span>
              <span class="text-indigo-400 font-semibold shrink-0">[{{ log.component }}]</span>
              <span class="text-slate-200 flex-1 leading-relaxed">{{ log.msg }}</span>
              @if (log.traceID) {
                <span class="text-[10px] text-slate-500 font-mono shrink-0">trace={{ log.traceID }}</span>
              }
            </div>
          } @empty {
            <div class="py-12 text-center text-slate-500 italic">No logs matching filter criteria</div>
          }
        </div>
      </sg-card>
    </div>
  `
})
export class LogsComponent {
  readonly selectedLevel = signal<string>('ALL');
  readonly searchQuery = signal<string>('');

  readonly logs = signal<LogEntry[]>([
    { id: '1', level: 'INFO', component: 'straitd', msg: 'Discovered dynamically: clusterCIDR=10.244.0.0/16 serviceCIDR=10.96.0.0/12', traceID: 'tr-091a', timestamp: '2026-08-26T10:14:00Z' },
    { id: '2', level: 'INFO', component: 'bpf-loader', msg: 'Loaded CO-RE BPF object netkit_pod.o into kernel 6.12.8 (verifier verified in 42ms)', traceID: 'tr-091b', timestamp: '2026-08-26T10:14:02Z' },
    { id: '3', level: 'INFO', component: 'service-lb', msg: 'Compiled 14 Services into service_map with Maglev hash table size 128', traceID: 'tr-091c', timestamp: '2026-08-26T10:14:05Z' },
    { id: '4', level: 'WARN', component: 'policy-eval', msg: 'LPM trie hit deny rule for segment 50 cross-traffic attempt to node-internal port 6443', traceID: 'tr-092x', timestamp: '2026-08-26T10:14:15Z' },
    { id: '5', level: 'INFO', component: 'transit-gw', msg: 'Segment 0 backbone heartbeat OK with peer 198.51.100.25:51820 (latency 14.2ms)', traceID: 'tr-093a', timestamp: '2026-08-26T10:14:20Z' }
  ]);

  readonly filteredLogs = computed(() => {
    let list = this.logs();
    const lvl = this.selectedLevel();
    if (lvl !== 'ALL') {
      list = list.filter(l => l.level === lvl);
    }
    const q = this.searchQuery().toLowerCase().trim();
    if (q) {
      list = list.filter(l => l.msg.toLowerCase().includes(q) || l.component.toLowerCase().includes(q));
    }
    return list;
  });

  levelClass(level: string): string {
    switch (level) {
      case 'INFO': return 'bg-sky-500/10 text-sky-400 border border-sky-500/20';
      case 'WARN': return 'bg-amber-500/10 text-amber-400 border border-amber-500/20';
      case 'ERROR': return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
      default: return 'bg-slate-800 text-slate-400';
    }
  }
}

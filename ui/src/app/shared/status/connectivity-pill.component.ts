import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-connectivity-pill',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="inline-flex items-center gap-2 px-2.5 py-1 rounded-full text-xs font-mono border bg-slate-900/80 backdrop-blur"
         [ngClass]="borderClass()">
      <span class="w-1.5 h-1.5 rounded-full" [ngClass]="dotClass()"></span>
      <span class="text-slate-300 font-sans">{{ label() }}</span>
      @if (ping() !== undefined) {
        <span class="text-[10px] text-slate-400 font-mono">{{ ping() }}ms</span>
      }
    </div>
  `
})
export class ConnectivityPillComponent {
  readonly connected = input<boolean>(true);
  readonly label = input<string>('Connected');
  readonly ping = input<number | undefined>(undefined);

  dotClass(): string {
    return this.connected() ? 'bg-emerald-400' : 'bg-rose-500 animate-pulse';
  }

  borderClass(): string {
    return this.connected() ? 'border-emerald-500/30' : 'border-rose-500/40';
  }
}

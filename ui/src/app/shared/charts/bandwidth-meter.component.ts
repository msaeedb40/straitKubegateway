import { Component, input, computed } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-bandwidth-meter',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="flex flex-col gap-1.5 w-full">
      <div class="flex justify-between items-center text-xs">
        <span class="text-slate-400 font-medium text-[11px]">{{ label() }}</span>
        <span class="font-mono text-slate-200 font-semibold">{{ current() }} / {{ max() }}</span>
      </div>
      <div class="h-2 w-full bg-slate-800/80 rounded-full overflow-hidden p-0.5 border border-slate-700/40">
        <div 
          class="h-full rounded-full transition-all duration-500"
          [style.width.%]="percentage()"
          [ngClass]="colorClass()">
        </div>
      </div>
    </div>
  `
})
export class BandwidthMeterComponent {
  readonly label = input<string>('Bandwidth');
  readonly current = input<string>('4.2 Gbps');
  readonly max = input<string>('10 Gbps');
  readonly percentage = input<number>(42);

  readonly colorClass = computed(() => {
    const p = this.percentage();
    if (p > 85) return 'bg-gradient-to-r from-amber-500 to-rose-500';
    if (p > 60) return 'bg-gradient-to-r from-sky-500 to-indigo-500';
    return 'bg-gradient-to-r from-emerald-500 to-sky-500';
  });
}

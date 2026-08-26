import { Component, input, computed } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-timeseries-chart',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="w-full flex flex-col gap-1">
      <div class="h-20 w-full relative">
        <svg class="w-full h-full overflow-visible" preserveAspectRatio="none" viewBox="0 0 100 40">
          <defs>
            <linearGradient [id]="gradientId()" x1="0%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" [attr.stop-color]="color()" stop-opacity="0.35" />
              <stop offset="100%" [attr.stop-color]="color()" stop-opacity="0.0" />
            </linearGradient>
          </defs>
          <!-- Area fill -->
          <path [attr.d]="areaPath()" [attr.fill]="'url(#' + gradientId() + ')'" />
          <!-- Main Line -->
          <path [attr.d]="linePath()" fill="none" [attr.stroke]="color()" stroke-width="1.8" stroke-linecap="round" />
        </svg>
      </div>
      @if (label()) {
        <div class="flex justify-between items-center text-[10px] text-slate-400">
          <span>{{ label() }}</span>
          <span class="font-mono text-slate-300 font-semibold">{{ currentValue() }}</span>
        </div>
      }
    </div>
  `
})
export class TimeseriesChartComponent {
  readonly data = input<number[]>([12, 18, 15, 24, 28, 22, 35, 42, 38, 45, 52, 48]);
  readonly color = input<string>('#38bdf8');
  readonly label = input<string>('');
  readonly currentValue = input<string>('');
  readonly gradientId = computed(() => `grad-${Math.random().toString(36).substring(2, 8)}`);

  readonly linePath = computed(() => {
    const pts = this.data();
    if (pts.length < 2) return '';
    const min = Math.min(...pts);
    const max = Math.max(...pts) || 1;
    const range = max - min || 1;

    return pts.map((val, idx) => {
      const x = (idx / (pts.length - 1)) * 100;
      const y = 38 - ((val - min) / range) * 34;
      return `${idx === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`;
    }).join(' ');
  });

  readonly areaPath = computed(() => {
    const line = this.linePath();
    if (!line) return '';
    return `${line} L 100 40 L 0 40 Z`;
  });
}

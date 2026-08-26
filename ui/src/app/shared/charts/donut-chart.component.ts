import { Component, input, computed } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface DonutSegment {
  label: string;
  value: number;
  color: string;
}

@Component({
  selector: 'sg-donut-chart',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="flex items-center gap-4">
      <div class="relative w-20 h-20 shrink-0">
        <svg class="w-full h-full transform -rotate-90" viewBox="0 0 36 36">
          <!-- Background circle -->
          <circle cx="18" cy="18" r="14" fill="none" stroke="rgba(255,255,255,0.06)" stroke-width="3.5" />
          <!-- Slices -->
          @for (s of calculatedSegments(); track s.label) {
            <circle
              cx="18"
              cy="18"
              r="14"
              fill="none"
              [attr.stroke]="s.color"
              stroke-width="3.5"
              [attr.stroke-dasharray]="s.dashArray"
              [attr.stroke-dashoffset]="s.dashOffset"
              stroke-linecap="round"
              class="transition-all duration-700" />
          }
        </svg>
        <div class="absolute inset-0 flex flex-col items-center justify-center text-center">
          <span class="text-xs font-bold text-slate-100 font-mono">{{ centerText() }}</span>
        </div>
      </div>
      <div class="flex flex-col gap-1 text-xs">
        @for (s of segments(); track s.label) {
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full shrink-0" [style.background-color]="s.color"></span>
            <span class="text-slate-400 text-[11px]">{{ s.label }}:</span>
            <span class="text-slate-200 font-mono font-medium text-[11px]">{{ s.value }}</span>
          </div>
        }
      </div>
    </div>
  `
})
export class DonutChartComponent {
  readonly segments = input<DonutSegment[]>([
    { label: 'Allowed', value: 85, color: '#10b981' },
    { label: 'Denied', value: 12, color: '#f43f5e' },
    { label: 'Dropped', value: 3, color: '#f59e0b' }
  ]);
  readonly centerText = input<string>('99.9%');

  readonly calculatedSegments = computed(() => {
    const total = this.segments().reduce((acc, s) => acc + s.value, 0) || 1;
    const circumference = 2 * Math.PI * 14; // ~87.96
    let currentOffset = 0;

    return this.segments().map(s => {
      const pct = s.value / total;
      const strokeLen = pct * circumference;
      const dashArray = `${strokeLen} ${circumference - strokeLen}`;
      const dashOffset = -currentOffset;
      currentOffset += strokeLen;

      return {
        ...s,
        dashArray,
        dashOffset
      };
    });
  });
}

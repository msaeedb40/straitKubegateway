import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-badge',
  standalone: true,
  imports: [CommonModule],
  template: `
    <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-medium border"
          [ngClass]="colorClasses()">
      <ng-content></ng-content>
    </span>
  `
})
export class BadgeComponent {
  readonly variant = input<'primary' | 'secondary' | 'success' | 'warning' | 'danger' | 'info'>('secondary');

  colorClasses(): string {
    switch (this.variant()) {
      case 'primary':
        return 'bg-sky-500/10 text-sky-400 border-sky-500/30';
      case 'success':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30';
      case 'warning':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/30';
      case 'danger':
        return 'bg-rose-500/10 text-rose-400 border-rose-500/30';
      case 'info':
        return 'bg-indigo-500/10 text-indigo-400 border-indigo-500/30';
      default:
        return 'bg-slate-800 text-slate-300 border-slate-700/60';
    }
  }
}

import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-status-badge',
  standalone: true,
  imports: [CommonModule],
  template: `
    <span
      class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-medium border tracking-wide"
      [ngClass]="getClasses()"
      [attr.aria-label]="'Status: ' + status"
    >
      <span class="w-1.5 h-1.5 rounded-full" [ngClass]="getDotClass()"></span>
      <ng-content></ng-content>
      @if (!hasContent) {
        {{ status }}
      }
    </span>
  `
})
export class StatusBadgeComponent {
  @Input() status: 'Ready' | 'Established' | 'Active' | 'Normal' | 'Degraded' | 'Warning' | 'Connecting' | 'Pending' | 'Down' | 'Error' | 'FORWARDED' | 'DROPPED' | 'REDIRECTED' | string = 'Ready';
  @Input() hasContent: boolean = false;

  getClasses(): string {
    const s = this.status.toLowerCase();
    if (['ready', 'established', 'active', 'normal', 'forwarded'].includes(s)) {
      return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    }
    if (['degraded', 'warning', 'connecting', 'redirected'].includes(s)) {
      return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    }
    if (['down', 'error', 'dropped'].includes(s)) {
      return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
    }
    return 'bg-slate-500/10 text-slate-400 border-slate-500/20';
  }

  getDotClass(): string {
    const s = this.status.toLowerCase();
    if (['ready', 'established', 'active', 'normal', 'forwarded'].includes(s)) {
      return 'bg-emerald-400';
    }
    if (['degraded', 'warning', 'connecting', 'redirected'].includes(s)) {
      return 'bg-amber-400';
    }
    if (['down', 'error', 'dropped'].includes(s)) {
      return 'bg-rose-400';
    }
    return 'bg-slate-400';
  }
}

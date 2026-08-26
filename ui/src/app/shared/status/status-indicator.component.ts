import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-status-indicator',
  standalone: true,
  imports: [CommonModule],
  template: `
    <span class="inline-flex items-center gap-1.5 font-medium text-xs">
      <span class="relative flex h-2 w-2">
        @if (pulse()) {
          <span 
            class="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75"
            [ngClass]="dotColorClass()">
          </span>
        }
        <span 
          class="relative inline-flex rounded-full h-2 w-2"
          [ngClass]="dotColorClass()">
        </span>
      </span>
      @if (label()) {
        <span [ngClass]="textColorClass()">{{ label() }}</span>
      }
    </span>
  `
})
export class StatusIndicatorComponent {
  readonly status = input<'ready' | 'success' | 'warning' | 'error' | 'pending' | 'active'>('ready');
  readonly label = input<string>('');
  readonly pulse = input<boolean>(true);

  dotColorClass(): string {
    switch (this.status()) {
      case 'ready':
      case 'success':
      case 'active':
        return 'bg-emerald-400';
      case 'warning':
      case 'pending':
        return 'bg-amber-400';
      case 'error':
        return 'bg-rose-500';
      default:
        return 'bg-slate-400';
    }
  }

  textColorClass(): string {
    switch (this.status()) {
      case 'ready':
      case 'success':
      case 'active':
        return 'text-emerald-400';
      case 'warning':
      case 'pending':
        return 'text-amber-400';
      case 'error':
        return 'text-rose-400';
      default:
        return 'text-slate-400';
    }
  }
}

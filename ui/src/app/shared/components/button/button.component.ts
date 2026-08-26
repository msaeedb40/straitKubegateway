import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-button',
  standalone: true,
  imports: [CommonModule],
  template: `
    <button
      [type]="type()"
      [disabled]="disabled() || loading()"
      (click)="clicked.emit($event)"
      class="inline-flex items-center justify-center gap-2 font-medium text-xs rounded-lg transition active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
      [ngClass]="[variantClasses(), sizeClasses()]">
      @if (loading()) {
        <span class="w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin"></span>
      }
      <ng-content></ng-content>
    </button>
  `
})
export class ButtonComponent {
  readonly variant = input<'primary' | 'secondary' | 'outline' | 'danger' | 'ghost'>('secondary');
  readonly size = input<'sm' | 'md' | 'lg'>('md');
  readonly type = input<'button' | 'submit' | 'reset'>('button');
  readonly disabled = input<boolean>(false);
  readonly loading = input<boolean>(false);
  readonly clicked = output<MouseEvent>();

  variantClasses(): string {
    switch (this.variant()) {
      case 'primary':
        return 'bg-gradient-to-r from-sky-500 to-indigo-600 hover:from-sky-400 hover:to-indigo-500 text-white shadow-md shadow-sky-500/20';
      case 'outline':
        return 'border border-slate-700 bg-transparent hover:bg-slate-800 text-slate-200';
      case 'danger':
        return 'bg-rose-600 hover:bg-rose-500 text-white shadow-sm';
      case 'ghost':
        return 'bg-transparent hover:bg-slate-800/60 text-slate-300 hover:text-slate-100';
      default:
        return 'bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700/60';
    }
  }

  sizeClasses(): string {
    switch (this.size()) {
      case 'sm':
        return 'px-2.5 py-1 text-xs';
      case 'lg':
        return 'px-5 py-2.5 text-sm';
      default:
        return 'px-3.5 py-1.5 text-xs';
    }
  }
}

import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-error-state',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="flex flex-col items-center justify-center p-12 text-center rounded-2xl bg-rose-950/20 border border-rose-800/40" role="alert">
      <div class="w-12 h-12 rounded-xl bg-rose-500/10 border border-rose-500/20 flex items-center justify-center text-rose-400 mb-3">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h3 class="text-sm font-semibold text-rose-200">{{ title }}</h3>
      <p class="text-xs text-rose-300/80 mt-1 max-w-sm font-mono">{{ message }}</p>
      @if (traceId) {
        <span class="text-[10px] text-slate-500 font-mono mt-1">Trace ID: {{ traceId }}</span>
      }
      <button
        (click)="retry.emit()"
        class="mt-4 px-3 py-1.5 rounded-lg bg-rose-600/30 hover:bg-rose-600/50 text-rose-200 border border-rose-500/40 text-xs font-semibold transition-all flex items-center gap-1.5"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
        Retry Operation
      </button>
    </div>
  `
})
export class ErrorStateComponent {
  @Input() title: string = 'Communication Error';
  @Input() message: string = 'Unable to synchronize domain state with sg-controller.';
  @Input() traceId?: string;
  @Output() retry = new EventEmitter<void>();
}

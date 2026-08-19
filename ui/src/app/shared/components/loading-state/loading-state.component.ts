import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-loading-state',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="flex flex-col items-center justify-center p-12 text-center" role="status" aria-live="polite">
      <div class="w-8 h-8 rounded-full border-2 border-indigo-500/20 border-t-indigo-500 animate-spin mb-3"></div>
      <p class="text-xs text-slate-400 font-medium">{{ message }}</p>
      <span class="sr-only">Loading content...</span>
    </div>
  `
})
export class LoadingStateComponent {
  @Input() message: string = 'Loading real-time telemetry...';
}

import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-health-indicator',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="flex items-center gap-2" [attr.aria-label]="label + ': ' + (isHealthy ? 'Healthy' : 'Unhealthy')">
      <span class="relative flex h-2 w-2">
        @if (isHealthy && pulse) {
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75" [ngClass]="isHealthy ? 'bg-emerald-400' : 'bg-rose-400'"></span>
        }
        <span class="relative inline-flex rounded-full h-2 w-2" [ngClass]="isHealthy ? 'bg-emerald-400' : 'bg-rose-400'"></span>
      </span>
      <span class="text-xs font-mono" [ngClass]="isHealthy ? 'text-emerald-400' : 'text-rose-400'">
        {{ label }}
      </span>
    </div>
  `
})
export class HealthIndicatorComponent {
  @Input() isHealthy: boolean = true;
  @Input() label: string = 'Healthy';
  @Input() pulse: boolean = true;
}

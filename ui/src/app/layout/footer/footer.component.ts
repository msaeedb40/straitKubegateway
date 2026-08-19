import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { TelemetryService } from '../../core/logging/telemetry.service';

@Component({
  selector: 'app-footer',
  standalone: true,
  imports: [CommonModule],
  template: `
    <footer class="h-8 border-t border-slate-800/60 bg-slate-950/80 px-6 flex items-center justify-between text-[10px] text-slate-500 font-mono">
      <div class="flex items-center gap-4">
        <span>straitKubegateway UI v1.0 (Zoneless)</span>
        <span class="text-slate-700">|</span>
        <span>Trace: <span class="text-indigo-400">{{ telemetry.currentTraceId() }}</span></span>
      </div>
      <div class="flex items-center gap-4">
        <span>Last Event Sync: {{ state.lastSync() | date:'HH:mm:ss.SSS' }}</span>
      </div>
    </footer>
  `
})
export class FooterComponent {
  readonly state = inject(StateService);
  readonly telemetry = inject(TelemetryService);
}

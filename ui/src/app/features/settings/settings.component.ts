import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ConfigService } from '../../core/config/config.service';
import { AuthService } from '../../core/auth/auth.service';
import { TelemetryService } from '../../core/logging/telemetry.service';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6 max-w-4xl mx-auto">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-white">System & Controller Settings</h1>
        <p class="text-xs text-slate-400 mt-1">Runtime externalized configuration, authentication identity, and telemetry endpoints.</p>
      </div>

      <!-- Runtime Config Card -->
      <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
        <h2 class="text-sm font-semibold text-white">External Runtime Configuration (/config/runtime-config.json)</h2>
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 font-mono text-xs text-slate-300 overflow-x-auto">
          <pre>{{ configService.config() | json }}</pre>
        </div>
      </div>

      <!-- Auth Context -->
      <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
        <h2 class="text-sm font-semibold text-white">Authentication & Roles Context</h2>
        <div class="grid grid-cols-2 gap-4 text-xs">
          <div class="p-3 rounded-xl bg-slate-950/60 border border-slate-800">
            <span class="text-slate-400">Current User:</span>
            <div class="font-mono text-white mt-1">{{ authService.user()?.username }}</div>
          </div>
          <div class="p-3 rounded-xl bg-slate-950/60 border border-slate-800">
            <span class="text-slate-400">Assigned Roles:</span>
            <div class="font-mono text-indigo-300 mt-1">{{ authService.userRoles().join(', ') }}</div>
          </div>
        </div>
      </div>

      <!-- Telemetry Info -->
      <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
        <h2 class="text-sm font-semibold text-white">Distributed Tracing (OpenTelemetry)</h2>
        <div class="p-3 rounded-xl bg-slate-950/60 border border-slate-800 text-xs">
          <span class="text-slate-400">Active Trace ID:</span>
          <div class="font-mono text-emerald-400 mt-1">{{ telemetry.currentTraceId() }}</div>
        </div>
      </div>
    </div>
  `
})
export class SettingsComponent {
  readonly configService = inject(ConfigService);
  readonly authService = inject(AuthService);
  readonly telemetry = inject(TelemetryService);
}

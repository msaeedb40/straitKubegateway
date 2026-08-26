import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CardComponent } from '../../shared/components/card/card.component';
import { ButtonComponent } from '../../shared/components/button/button.component';
import { PreferencesService } from '../../core/services/preferences.service';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    CardComponent,
    ButtonComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Control Plane & Dashboard Settings</h1>
          <p class="text-xs text-slate-400 mt-1">
            Configure StraitKubeGateway runtime parameters, telemetry thresholds, and UI preferences
          </p>
        </div>
        <div class="flex items-center gap-2">
          <sg-button variant="primary" (clicked)="saveSettings()">
            Save Changes
          </sg-button>
        </div>
      </div>

      <!-- Settings Tabs -->
      <div class="flex items-center gap-2 border-b border-slate-800 pb-2 overflow-x-auto select-none">
        @for (tab of tabs; track tab.id) {
          <button 
            (click)="activeTab.set(tab.id)"
            class="px-3 py-1.5 rounded-lg text-xs font-medium transition cursor-pointer"
            [ngClass]="activeTab() === tab.id ? 'bg-sky-500/10 text-sky-400 border border-sky-500/30' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900'">
            {{ tab.label }}
          </button>
        }
      </div>

      <!-- General Settings -->
      @if (activeTab() === 'general') {
        <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
          <sg-card title="Cluster & Gateway Identity" subtitle="Basic cluster identification parameters">
            <div class="space-y-4 mt-2 text-xs">
              <div>
                <label class="block text-slate-400 mb-1">Cluster ID</label>
                <input 
                  type="text" 
                  [(ngModel)]="settings.clusterID"
                  class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:border-sky-500 outline-none font-mono" />
              </div>
              <div>
                <label class="block text-slate-400 mb-1">Transit Segment 0 Backbone IP</label>
                <input 
                  type="text" 
                  [(ngModel)]="settings.backboneIP"
                  class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:border-sky-500 outline-none font-mono" />
              </div>
            </div>
          </sg-card>

          <sg-card title="Dashboard Telemetry Preferences" subtitle="Refresh rate and local UI parameters">
            <div class="space-y-4 mt-2 text-xs">
              <div>
                <label class="block text-slate-400 mb-1">Telemetry Polling Interval (ms)</label>
                <input 
                  type="number" 
                  [(ngModel)]="settings.refreshInterval"
                  class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:border-sky-500 outline-none font-mono" />
              </div>
              <div class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800">
                <span class="text-slate-300">Enable D3 Force Simulation Animation</span>
                <input 
                  type="checkbox" 
                  [(ngModel)]="settings.d3Simulation"
                  class="w-4 h-4 rounded bg-slate-900 border-slate-700 text-sky-500 focus:ring-0 cursor-pointer" />
              </div>
            </div>
          </sg-card>
        </div>
      }

      <!-- Dataplane & eBPF Settings -->
      @if (activeTab() === 'ebpf') {
        <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
          <sg-card title="Kernel Dataplane Driver" subtitle="eBPF hook configuration">
            <div class="space-y-4 mt-2 text-xs">
              <div class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800">
                <div>
                  <div class="text-slate-200 font-semibold">kube-proxy Replacement</div>
                  <div class="text-[11px] text-slate-400 mt-0.5">Redirect socket connections at cgroup connect4 hook</div>
                </div>
                <span class="text-emerald-400 font-mono text-[11px]">Enabled (none)</span>
              </div>
              <div class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800">
                <div>
                  <div class="text-slate-200 font-semibold">NetKit Fastpath</div>
                  <div class="text-[11px] text-slate-400 mt-0.5">0-copy container namespace redirection</div>
                </div>
                <span class="text-emerald-400 font-mono text-[11px]">Active</span>
              </div>
            </div>
          </sg-card>

          <sg-card title="Load Balancing Algorithm" subtitle="Service VIP distribution">
            <div class="space-y-4 mt-2 text-xs">
              <div>
                <label class="block text-slate-400 mb-1">Hash Algorithm</label>
                <select 
                  [(ngModel)]="settings.lbAlgorithm"
                  class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:border-sky-500 outline-none font-mono">
                  <option value="maglev">Maglev Consistent Hash (Table size 128)</option>
                  <option value="round-robin">Round Robin</option>
                  <option value="least-conn">Least Connections</option>
                </select>
              </div>
              <div class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800">
                <span class="text-slate-300">Direct Server Return (DSR)</span>
                <input 
                  type="checkbox" 
                  [(ngModel)]="settings.enableDSR"
                  class="w-4 h-4 rounded bg-slate-900 border-slate-700 text-sky-500 focus:ring-0 cursor-pointer" />
              </div>
            </div>
          </sg-card>
        </div>
      }
    </div>
  `
})
export class SettingsComponent {
  private readonly prefs = inject(PreferencesService);
  private readonly notif = inject(NotificationService);

  readonly activeTab = signal<string>('general');

  readonly tabs = [
    { id: 'general', label: 'General' },
    { id: 'ebpf', label: 'eBPF & Dataplane' },
    { id: 'network', label: 'Network & CNI' },
    { id: 'security', label: 'Security & Policy' }
  ];

  settings = {
    clusterID: 'strait-cluster-01',
    backboneIP: '10.0.0.1',
    refreshInterval: 5000,
    d3Simulation: true,
    lbAlgorithm: 'maglev',
    enableDSR: true
  };

  saveSettings(): void {
    this.prefs.update({
      telemetryRefreshInterval: this.settings.refreshInterval,
      d3SimulationEnabled: this.settings.d3Simulation
    });

    this.notif.show({
      type: 'success',
      title: 'Settings Saved',
      message: 'Control plane parameters and preferences updated successfully.',
      autoCloseMs: 3000
    });
  }
}

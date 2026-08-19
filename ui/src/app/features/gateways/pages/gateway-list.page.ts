import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { GatewayStore } from '../data-access/gateway.store';
import { GatewayCardComponent } from '../components/gateway-card.component';
import { ConfirmationDialogComponent } from '../../../shared/components/confirmation-dialog/confirmation-dialog.component';

@Component({
  selector: 'app-gateway-list-page',
  standalone: true,
  imports: [CommonModule, GatewayCardComponent, ConfirmationDialogComponent],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Gateways & Listeners</h1>
          <p class="text-xs text-slate-400 mt-1">Manage Kubernetes Gateway API instances, port listeners, and TLS certificates.</p>
        </div>
        <button
          (click)="openCreateModal()"
          class="px-3.5 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-xs font-semibold text-white shadow-lg shadow-indigo-600/20 transition-all flex items-center gap-1.5"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
          Create Gateway
        </button>
      </div>

      <!-- Gateway Cards -->
      <div class="grid grid-cols-1 gap-4">
        @for (gw of store.gateways(); track gw.id) {
          <app-gateway-card [gateway]="gw" (restart)="onRestartRequested($event)"></app-gateway-card>
        }
      </div>

      <!-- Confirmation Dialog -->
      <app-confirmation-dialog
        [isOpen]="isRestartModalOpen()"
        title="Restart Gateway Pods"
        [message]="'Are you sure you want to gracefully restart gateway ' + selectedGatewayId() + '?'"
        confirmLabel="Restart"
        (confirm)="confirmRestart()"
        (cancel)="isRestartModalOpen.set(false)"
      ></app-confirmation-dialog>
    </div>
  `
})
export class GatewayListPageComponent {
  readonly store = inject(GatewayStore);
  readonly isRestartModalOpen = signal<boolean>(false);
  readonly selectedGatewayId = signal<string>('');

  onRestartRequested(id: string): void {
    this.selectedGatewayId.set(id);
    this.isRestartModalOpen.set(true);
  }

  confirmRestart(): void {
    const id = this.selectedGatewayId();
    if (id) {
      this.store.restartGateway(id);
    }
    this.isRestartModalOpen.set(false);
  }

  openCreateModal(): void {
    // Submit default demo gateway
    this.store.createGateway({
      name: `edge-gateway-${Math.floor(Math.random() * 1000)}`,
      namespace: 'default',
      segmentId: 100,
      mode: 'DirectRouting',
      listeners: [{ name: 'http-in', port: 8080, protocol: 'HTTP' }]
    });
  }
}

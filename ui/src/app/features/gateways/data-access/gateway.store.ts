import { Injectable, inject, computed } from '@angular/core';
import { StateService } from '../../../core/state/state.service';
import { GatewayApiService } from './gateway-api.service';
import { ApplicationStore } from '../../../core/state/application.store';
import { CreateGatewayPayload } from './gateway.models';

@Injectable({
  providedIn: 'root'
})
export class GatewayStore {
  private readonly stateService = inject(StateService);
  private readonly gatewayApi = inject(GatewayApiService);
  private readonly appStore = inject(ApplicationStore);

  readonly gateways = computed(() => this.stateService.gateways());
  readonly totalGateways = computed(() => this.gateways().length);
  readonly readyGateways = computed(() => this.gateways().filter(g => g.status === 'Ready').length);

  async createGateway(payload: CreateGatewayPayload): Promise<void> {
    this.gatewayApi.create(payload).subscribe({
      next: res => {
        this.appStore.addNotification({
          type: 'success',
          title: 'Gateway Creation Accepted',
          message: `Command ${res.commandId} accepted. Provisioning ${payload.name}...`,
          traceId: res.traceId
        });
      },
      error: err => {
        this.appStore.addNotification({
          type: 'error',
          title: 'Gateway Creation Failed',
          message: err.message || 'Unable to submit gateway command.'
        });
      }
    });
  }

  async restartGateway(id: string): Promise<void> {
    this.gatewayApi.restart(id).subscribe({
      next: res => {
        this.appStore.addNotification({
          type: 'info',
          title: 'Gateway Restart Initiated',
          message: `Command ${res.commandId} submitted for gateway ${id}.`,
          traceId: res.traceId
        });
      }
    });
  }
}

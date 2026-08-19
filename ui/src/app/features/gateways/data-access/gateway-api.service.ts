import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ControllerApiClient } from '../../../core/api/api-client';
import { GatewayListRequest, CommandResponse } from '../../../core/api/api.types';
import { GatewayItem } from '../../../core/models/models';
import { CreateGatewayPayload } from './gateway.models';

@Injectable({
  providedIn: 'root'
})
export class GatewayApiService {
  private readonly api = inject(ControllerApiClient);

  list(params?: GatewayListRequest): Observable<GatewayItem[]> {
    return this.api.get<GatewayItem[]>('/v1/gateways', params);
  }

  get(id: string): Observable<GatewayItem> {
    return this.api.get<GatewayItem>(`/v1/gateways/${id}`);
  }

  create(payload: CreateGatewayPayload): Observable<CommandResponse> {
    return this.api.executeCommand('/v1/gateways', payload);
  }

  restart(id: string): Observable<CommandResponse> {
    return this.api.executeCommand(`/v1/gateways/${id}/restart`, {});
  }

  delete(id: string): Observable<CommandResponse> {
    return this.api.executeCommand(`/v1/gateways/${id}/delete`, {});
  }
}

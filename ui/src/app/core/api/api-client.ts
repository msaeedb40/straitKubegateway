import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ConfigService } from '../config/config.service';
import { TelemetryService } from '../logging/telemetry.service';
import { CommandResponse } from './api.types';

@Injectable({
  providedIn: 'root'
})
export class ControllerApiClient {
  private readonly http = inject(HttpClient);
  private readonly configService = inject(ConfigService);
  private readonly telemetry = inject(TelemetryService);

  private get baseUrl(): string {
    return this.configService.config().api.baseUrl;
  }

  get<T>(path: string, queryParams?: Record<string, any>): Observable<T> {
    let params = new HttpParams();
    if (queryParams) {
      Object.keys(queryParams).forEach(key => {
        const val = queryParams[key];
        if (val !== undefined && val !== null && val !== '') {
          params = params.set(key, String(val));
        }
      });
    }
    const cleanPath = path.startsWith('/') ? path : `/${path}`;
    return this.http.get<T>(`${this.baseUrl}${cleanPath}`, { params });
  }

  post<T>(path: string, body: unknown = {}): Observable<T> {
    const cleanPath = path.startsWith('/') ? path : `/${path}`;
    return this.http.post<T>(`${this.baseUrl}${cleanPath}`, body);
  }

  put<T>(path: string, body: unknown = {}): Observable<T> {
    const cleanPath = path.startsWith('/') ? path : `/${path}`;
    return this.http.put<T>(`${this.baseUrl}${cleanPath}`, body);
  }

  delete<T>(path: string): Observable<T> {
    const cleanPath = path.startsWith('/') ? path : `/${path}`;
    return this.http.delete<T>(`${this.baseUrl}${cleanPath}`);
  }

  executeCommand(path: string, payload: unknown = {}): Observable<CommandResponse> {
    const commandId = this.telemetry.generateCommandId();
    const cleanPath = path.startsWith('/') ? path : `/${path}`;
    const envelope = {
      commandId,
      traceId: this.telemetry.currentTraceId(),
      timestamp: new Date().toISOString(),
      payload
    };
    return this.http.post<CommandResponse>(`${this.baseUrl}${cleanPath}`, envelope);
  }
}

import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, catchError, map, throwError } from 'rxjs';
import { ConfigService } from '../config/config.service';
import { TelemetryService } from '../logging/telemetry.service';
import { GrpcError, GrpcStatusCode } from './grpc-error';

@Injectable({
  providedIn: 'root'
})
export class GrpcClient {
  private readonly http = inject(HttpClient);
  private readonly configService = inject(ConfigService);
  private readonly telemetry = inject(TelemetryService);

  private get grpcEndpoint(): string {
    return this.configService.config().api.grpcWebUrl;
  }

  unaryCall<TReq, TRes>(serviceName: string, methodName: string, request: TReq): Observable<TRes> {
    const url = `${this.grpcEndpoint}/${serviceName}/${methodName}`;
    const traceId = this.telemetry.currentTraceId();

    const headers = new HttpHeaders({
      'content-type': 'application/grpc-web+proto',
      'x-grpc-web': '1',
      'x-trace-id': traceId
    });

    return this.http.post<TRes>(url, request, { headers }).pipe(
      catchError(err => {
        const grpcError = new GrpcError(
          err.status === 404 ? GrpcStatusCode.NOT_FOUND : GrpcStatusCode.UNAVAILABLE,
          err.message || 'gRPC call failed',
          traceId
        );
        return throwError(() => grpcError);
      })
    );
  }
}

import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { GrpcClient } from './grpc-client';

@Injectable({
  providedIn: 'root'
})
export class ControllerApiService {
  private readonly grpc = inject(GrpcClient);

  invokeGatewayCommand<TReq, TRes>(method: string, request: TReq): Observable<TRes> {
    return this.grpc.unaryCall<TReq, TRes>('straitgateway.v1.GatewayService', method, request);
  }

  invokeNetworkCommand<TReq, TRes>(method: string, request: TReq): Observable<TRes> {
    return this.grpc.unaryCall<TReq, TRes>('straitgateway.v1.NetworkService', method, request);
  }

  invokePolicyCommand<TReq, TRes>(method: string, request: TReq): Observable<TRes> {
    return this.grpc.unaryCall<TReq, TRes>('straitgateway.v1.PolicyService', method, request);
  }
}

export type { GatewayItem } from '../../../core/models/models';

export interface CreateGatewayPayload {
  name: string;
  namespace: string;
  segmentId: number;
  mode: string;
  listeners: {
    name: string;
    port: number;
    protocol: string;
    hostname?: string;
  }[];
}

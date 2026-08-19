import { Injectable, inject, signal, computed, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ApiService } from '../api/api.service';
import { RealtimeService } from '../realtime/realtime.service';
import { ConnectionStore } from './connection.store';
import { LoggerService } from '../logging/logger.service';
import {
  GatewayItem,
  RouteItem,
  PolicyItem,
  SegmentItem,
  TransitItem,
  TunnelItem,
  BGPPeerItem,
  ServiceItem,
  ConntrackItem,
  FlowItem,
  NodeItem,
  EventItem,
  ClusterStats
} from '../models/models';
import { firstValueFrom } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class StateService {
  private readonly api = inject(ApiService);
  private readonly realtime = inject(RealtimeService);
  private readonly connectionStore = inject(ConnectionStore);
  private readonly logger = inject(LoggerService);
  private readonly destroyRef = inject(DestroyRef);

  // State Signals
  readonly stats = signal<ClusterStats | null>(null);
  readonly gateways = signal<GatewayItem[]>([]);
  readonly routes = signal<RouteItem[]>([]);
  readonly policies = signal<PolicyItem[]>([]);
  readonly segments = signal<SegmentItem[]>([]);
  readonly transits = signal<TransitItem[]>([]);
  readonly tunnels = signal<TunnelItem[]>([]);
  readonly bgpPeers = signal<BGPPeerItem[]>([]);
  readonly services = signal<ServiceItem[]>([]);
  readonly conntrack = signal<ConntrackItem[]>([]);
  readonly flows = signal<FlowItem[]>([]);
  readonly nodes = signal<NodeItem[]>([]);
  readonly events = signal<EventItem[]>([]);
  readonly searchQuery = signal<string>('');
  readonly lastSync = signal<Date>(new Date());
  readonly isInitialLoading = signal<boolean>(true);

  // Computed signals
  readonly totalGateways = computed(() => this.gateways().length);
  readonly totalPolicies = computed(() => this.policies().length);
  readonly totalRoutes = computed(() => this.routes().length);
  readonly healthyNodesCount = computed(() => this.nodes().filter(n => n.ebpfStatus === 'Active').length);
  readonly isConnected = computed(() => this.connectionStore.isConnected());

  constructor() {
    // Register snapshot loader with realtime service
    this.realtime.registerSnapshotLoader(() => this.loadInitialSnapshot());

    // Subscribe to realtime domain events
    this.realtime.events$
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(event => this.handleRealtimeEvent(event));

    // Load initial snapshot once (no periodic polling!)
    this.loadInitialSnapshot();

    // Start realtime stream transport
    this.realtime.initialize();
  }

  async loadInitialSnapshot(): Promise<void> {
    this.isInitialLoading.set(true);
    this.logger.info('StateService', 'loading_initial_snapshot');
    try {
      const [
        stats,
        gateways,
        routes,
        policies,
        segments,
        transits,
        tunnels,
        bgpPeers,
        services,
        conntrack,
        flows,
        nodes,
        events
      ] = await Promise.all([
        firstValueFrom(this.api.getClusterStats()),
        firstValueFrom(this.api.getGateways()),
        firstValueFrom(this.api.getRoutes()),
        firstValueFrom(this.api.getPolicies()),
        firstValueFrom(this.api.getSegments()),
        firstValueFrom(this.api.getTransits()),
        firstValueFrom(this.api.getTunnels()),
        firstValueFrom(this.api.getBGPPeers()),
        firstValueFrom(this.api.getServices()),
        firstValueFrom(this.api.getConntrack()),
        firstValueFrom(this.api.getFlows()),
        firstValueFrom(this.api.getNodes()),
        firstValueFrom(this.api.getEvents())
      ]);

      this.stats.set(stats);
      this.gateways.set(gateways);
      this.routes.set(routes);
      this.policies.set(policies);
      this.segments.set(segments);
      this.transits.set(transits);
      this.tunnels.set(tunnels);
      this.bgpPeers.set(bgpPeers);
      this.services.set(services);
      this.conntrack.set(conntrack);
      this.flows.set(flows);
      this.nodes.set(nodes);
      this.events.set(events);
      this.lastSync.set(new Date());
    } catch (err) {
      this.logger.error('StateService', 'initial_snapshot_failed', { metadata: { error: String(err) } });
    } finally {
      this.isInitialLoading.set(false);
    }
  }

  refreshAll(): void {
    this.loadInitialSnapshot();
  }

  private handleRealtimeEvent(event: any): void {
    this.lastSync.set(new Date());

    if (!event || !event.type) return;

    // Incremental updates per resource type
    switch (event.resource?.type) {
      case 'gateway':
        this.handleGatewayEvent(event);
        break;
      case 'flow':
        this.handleFlowEvent(event);
        break;
      case 'tunnel':
        this.handleTunnelEvent(event);
        break;
      case 'policy':
        this.handlePolicyEvent(event);
        break;
      case 'node':
        this.handleNodeEvent(event);
        break;
      case 'event':
        if (event.state?.current) {
          this.events.update(list => [event.state.current, ...list].slice(0, 100));
        }
        break;
    }
  }

  private handleGatewayEvent(event: any): void {
    if (event.state?.current) {
      const updated: GatewayItem = event.state.current;
      this.gateways.update(list => {
        const idx = list.findIndex(g => g.id === updated.id);
        if (idx >= 0) {
          const clone = [...list];
          clone[idx] = updated;
          return clone;
        }
        return [updated, ...list];
      });
    }
  }

  private handleFlowEvent(event: any): void {
    if (event.state?.current) {
      const flow: FlowItem = event.state.current;
      this.flows.update(list => [flow, ...list].slice(0, 200));
    }
  }

  private handleTunnelEvent(event: any): void {
    if (event.state?.current) {
      const updated: TunnelItem = event.state.current;
      this.tunnels.update(list => {
        const idx = list.findIndex(t => t.id === updated.id);
        if (idx >= 0) {
          const clone = [...list];
          clone[idx] = updated;
          return clone;
        }
        return [updated, ...list];
      });
    }
  }

  private handlePolicyEvent(event: any): void {
    if (event.state?.current) {
      const updated: PolicyItem = event.state.current;
      this.policies.update(list => {
        const idx = list.findIndex(p => p.id === updated.id);
        if (idx >= 0) {
          const clone = [...list];
          clone[idx] = updated;
          return clone;
        }
        return [updated, ...list];
      });
    }
  }

  private handleNodeEvent(event: any): void {
    if (event.state?.current) {
      const updated: NodeItem = event.state.current;
      this.nodes.update(list => {
        const idx = list.findIndex(n => n.id === updated.id);
        if (idx >= 0) {
          const clone = [...list];
          clone[idx] = updated;
          return clone;
        }
        return [updated, ...list];
      });
    }
  }

  setSearchQuery(q: string): void {
    this.searchQuery.set(q);
  }
}

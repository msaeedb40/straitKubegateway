import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'dashboard',
    pathMatch: 'full'
  },
  {
    path: 'dashboard',
    loadComponent: () => import('./features/dashboard/dashboard.component').then(m => m.DashboardComponent)
  },
  {
    path: 'gateways',
    loadComponent: () => import('./features/gateways/gateways.component').then(m => m.GatewaysComponent)
  },
  {
    path: 'nodes',
    loadComponent: () => import('./features/nodes/nodes.component').then(m => m.NodesComponent)
  },
  {
    path: 'tunnels',
    loadComponent: () => import('./features/tunnels/tunnels.component').then(m => m.TunnelsComponent)
  },
  {
    path: 'flows',
    loadComponent: () => import('./features/flows/flows.component').then(m => m.FlowsComponent)
  },
  {
    path: 'topology',
    loadComponent: () => import('./features/topology/topology.component').then(m => m.TopologyComponent)
  },
  {
    path: 'services',
    loadComponent: () => import('./features/services/services.component').then(m => m.ServicesComponent)
  },
  {
    path: 'endpoints',
    loadComponent: () => import('./features/endpoints/endpoints.component').then(m => m.EndpointsComponent)
  },
  {
    path: 'ebpf',
    loadComponent: () => import('./features/ebpf/ebpf.component').then(m => m.EbpfComponent)
  },
  {
    path: 'cni',
    loadComponent: () => import('./features/cni/cni.component').then(m => m.CniComponent)
  },
  {
    path: 'events',
    loadComponent: () => import('./features/events/events.component').then(m => m.EventsComponent)
  },
  {
    path: 'logs',
    loadComponent: () => import('./features/logs/logs.component').then(m => m.LogsComponent)
  },
  {
    path: 'metrics',
    loadComponent: () => import('./features/metrics/metrics.component').then(m => m.MetricsComponent)
  },
  {
    path: 'settings',
    loadComponent: () => import('./features/settings/settings.component').then(m => m.SettingsComponent)
  },
  {
    path: '**',
    redirectTo: 'dashboard'
  }
];

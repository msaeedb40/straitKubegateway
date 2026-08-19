import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  {
    path: 'dashboard',
    loadComponent: () => import('./features/dashboard/dashboard.component').then(m => m.DashboardComponent)
  },
  {
    path: 'gateways',
    loadComponent: () => import('./features/gateways/gateways.component').then(m => m.GatewaysComponent)
  },
  {
    path: 'routes',
    loadComponent: () => import('./features/routes/routes.component').then(m => m.RoutesComponent)
  },
  {
    path: 'policies',
    loadComponent: () => import('./features/policies/policies.component').then(m => m.PoliciesComponent)
  },
  {
    path: 'namespaces',
    loadComponent: () => import('./features/namespaces/namespaces.component').then(m => m.NamespacesComponent)
  },
  {
    path: 'workloads',
    loadComponent: () => import('./features/workloads/workloads.component').then(m => m.WorkloadsComponent)
  },
  {
    path: 'services',
    loadComponent: () => import('./features/services/services.component').then(m => m.ServicesComponent)
  },
  {
    path: 'nat',
    loadComponent: () => import('./features/nat/nat.component').then(m => m.NatComponent)
  },
  {
    path: 'network',
    loadComponent: () => import('./features/network/network.component').then(m => m.NetworkComponent)
  },
  {
    path: 'transit',
    loadComponent: () => import('./features/transit/transit.component').then(m => m.TransitComponent)
  },
  {
    path: 'tunnels',
    loadComponent: () => import('./features/tunnels/tunnels.component').then(m => m.TunnelsComponent)
  },
  {
    path: 'bgp',
    loadComponent: () => import('./features/bgp/bgp.component').then(m => m.BgpComponent)
  },
  {
    path: 'flows',
    loadComponent: () => import('./features/flows/flows.component').then(m => m.FlowsComponent)
  },
  {
    path: 'nodes',
    loadComponent: () => import('./features/nodes/nodes.component').then(m => m.NodesComponent)
  },
  {
    path: 'observability',
    loadComponent: () => import('./features/observability/observability.component').then(m => m.ObservabilityComponent)
  },
  {
    path: 'events',
    loadComponent: () => import('./features/events/events.component').then(m => m.EventsComponent)
  },
  {
    path: 'settings',
    loadComponent: () => import('./features/settings/settings.component').then(m => m.SettingsComponent)
  },
  { path: '**', redirectTo: 'dashboard' }
];

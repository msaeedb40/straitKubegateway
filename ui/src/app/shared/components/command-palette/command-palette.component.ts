import { Component, inject, signal, HostListener, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { ApplicationStore } from '../../../core/state/application.store';
import { FocusTrapDirective } from '../../directives/focus-trap.directive';

interface PaletteItem {
  id: string;
  title: string;
  category: string;
  route: string;
  icon: string;
}

@Component({
  selector: 'app-command-palette',
  standalone: true,
  imports: [CommonModule, FocusTrapDirective],
  template: `
    @if (appStore.isCommandPaletteOpen()) {
      <div class="fixed inset-0 z-50 flex items-start justify-center pt-20 p-4 bg-slate-950/80 backdrop-blur-md" (click)="close()">
        <div
          appFocusTrap
          class="w-full max-w-xl rounded-2xl bg-slate-900 border border-slate-800 shadow-2xl overflow-hidden flex flex-col"
          (click)="$event.stopPropagation()"
        >
          <!-- Search Header -->
          <div class="p-4 border-b border-slate-800 flex items-center gap-3">
            <svg class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              type="text"
              placeholder="Type a command or navigate (e.g. Gateways, Flows, Tunnels, Observability)..."
              [value]="searchQuery()"
              (input)="onInput($event)"
              class="w-full bg-transparent text-sm text-slate-100 placeholder-slate-500 focus:outline-none"
              autofocus
            />
            <kbd class="px-2 py-0.5 text-[10px] font-mono bg-slate-800 text-slate-400 rounded border border-slate-700">ESC</kbd>
          </div>

          <!-- Items List -->
          <div class="max-h-80 overflow-y-auto p-2 space-y-1">
            @for (item of filteredItems(); track item.id; let i = $index) {
              <button
                (click)="navigate(item.route)"
                class="w-full flex items-center justify-between p-3 rounded-xl text-left text-xs transition-all hover:bg-indigo-600/20 hover:text-indigo-200 text-slate-300"
              >
                <div class="flex items-center gap-3">
                  <span class="text-slate-400 font-mono">{{ item.category }} &rsaquo;</span>
                  <span class="font-medium text-white">{{ item.title }}</span>
                </div>
                <span class="text-[10px] font-mono text-slate-500">{{ item.route }}</span>
              </button>
            }
            @if (filteredItems().length === 0) {
              <div class="p-6 text-center text-xs text-slate-500">
                No commands matching "{{ searchQuery() }}"
              </div>
            }
          </div>
        </div>
      </div>
    }
  `
})
export class CommandPaletteComponent {
  readonly appStore = inject(ApplicationStore);
  private readonly router = inject(Router);

  readonly searchQuery = signal<string>('');

  private readonly allItems: PaletteItem[] = [
    { id: '1', title: 'Dashboard Overview', category: 'Control', route: '/dashboard', icon: 'dashboard' },
    { id: '2', title: 'Gateways & Listeners', category: 'Control', route: '/gateways', icon: 'gateways' },
    { id: '3', title: 'L7 / L4 Routes', category: 'Control', route: '/routes', icon: 'routes' },
    { id: '4', title: 'Network Policies & Hierarchy', category: 'Control', route: '/policies', icon: 'policies' },
    { id: '5', title: 'Services & Maglev Hash VIPs', category: 'Fabric', route: '/services', icon: 'services' },
    { id: '6', title: 'NAT & Stateful Conntrack', category: 'Fabric', route: '/nat', icon: 'nat' },
    { id: '7', title: 'Segments & Multi-Tenant IPAM', category: 'Fabric', route: '/network', icon: 'network' },
    { id: '8', title: 'Transit Gateways & Meshing', category: 'Fabric', route: '/transit', icon: 'transit' },
    { id: '9', title: 'WireGuard & IPsec Tunnels', category: 'Fabric', route: '/tunnels', icon: 'tunnels' },
    { id: '10', title: 'BGP Routing & Peering', category: 'Fabric', route: '/bgp', icon: 'bgp' },
    { id: '11', title: 'eBPF Live Flow Observer', category: 'Telemetry', route: '/flows', icon: 'flows' },
    { id: '12', title: 'Kubernetes Nodes & eBPF State', category: 'Telemetry', route: '/nodes', icon: 'nodes' },
    { id: '13', title: 'Observability & P95/P99 Latency', category: 'Telemetry', route: '/observability', icon: 'observability' },
    { id: '14', title: 'Audit & Operational Events', category: 'Telemetry', route: '/events', icon: 'events' },
    { id: '15', title: 'Namespaces & Multi-Tenancy', category: 'Control', route: '/namespaces', icon: 'namespaces' },
    { id: '16', title: 'Workloads & Pods', category: 'Control', route: '/workloads', icon: 'workloads' },
    { id: '17', title: 'System & Controller Settings', category: 'Settings', route: '/settings', icon: 'settings' }
  ];

  readonly filteredItems = computed(() => {
    const q = this.searchQuery().toLowerCase().trim();
    if (!q) return this.allItems;
    return this.allItems.filter(
      item => item.title.toLowerCase().includes(q) || item.category.toLowerCase().includes(q) || item.route.toLowerCase().includes(q)
    );
  });

  @HostListener('window:keydown', ['$event'])
  handleGlobalKey(event: KeyboardEvent): void {
    if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
      event.preventDefault();
      this.appStore.setCommandPalette(!this.appStore.isCommandPaletteOpen());
    } else if (event.key === 'Escape' && this.appStore.isCommandPaletteOpen()) {
      this.close();
    }
  }

  onInput(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.searchQuery.set(input.value);
  }

  navigate(route: string): void {
    this.router.navigateByUrl(route);
    this.close();
  }

  close(): void {
    this.appStore.setCommandPalette(false);
    this.searchQuery.set('');
  }
}

import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ConfigService } from '../../core/config/config.service';
import { ApplicationStore } from '../../core/state/application.store';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterModule],
  template: `
    <aside class="w-64 flex-shrink-0 flex flex-col border-r border-slate-800/80 bg-slate-900/60 backdrop-blur-xl h-full select-none">
      <!-- Logo Header -->
      <div class="h-16 flex items-center px-6 border-b border-slate-800/60 gap-3">
        <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-indigo-500 via-sky-500 to-emerald-400 flex items-center justify-center shadow-lg shadow-indigo-500/20 ring-1 ring-white/20">
          <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </div>
        <div class="flex flex-col">
          <span class="font-bold text-sm tracking-wide bg-clip-text text-transparent bg-gradient-to-r from-indigo-200 via-white to-sky-200">straitKube</span>
          <span class="text-[10px] text-indigo-400 font-mono tracking-wider -mt-0.5">GATEWAY v1.0</span>
        </div>
      </div>

      <!-- Navigation Links -->
      <nav class="flex-1 overflow-y-auto px-3 py-4 space-y-1 custom-scrollbar">
        <div class="px-3 py-1.5 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Control & Operations</div>
        <a routerLink="/dashboard" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" /></svg>
          Dashboard
        </a>
        <a routerLink="/gateways" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>
          Gateways & Listeners
        </a>
        <a routerLink="/routes" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7" /></svg>
          Routes (L7/L4)
        </a>
        <a routerLink="/policies" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
          Policies
        </a>
        <a routerLink="/namespaces" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2" /></svg>
          Namespaces
        </a>
        <a routerLink="/workloads" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" /></svg>
          Workloads & Pods
        </a>

        <div class="pt-3 px-3 py-1.5 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Networking & Fabrics</div>
        <a routerLink="/services" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" /></svg>
          Services & Maglev
        </a>
        <a routerLink="/nat" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" /></svg>
          NAT & Conntrack
        </a>
        <a routerLink="/network" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 100-6 3 3 0 000 6z" /></svg>
          Segments & Topology
        </a>
        <a routerLink="/transit" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" /></svg>
          Transit Gateways
        </a>
        <a routerLink="/tunnels" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
          Tunnels (WG / IPsec)
        </a>
        <a routerLink="/bgp" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
          BGP Routing
        </a>

        <div class="pt-3 px-3 py-1.5 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Telemetry & Operations</div>
        <a routerLink="/flows" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>
          Flow Observer
        </a>
        <a routerLink="/nodes" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7" /></svg>
          Nodes & eBPF
        </a>
        <a routerLink="/observability" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
          Observability
        </a>
        <a routerLink="/events" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
          Audit Events
        </a>
        <a routerLink="/settings" routerLinkActive="bg-indigo-600/20 text-indigo-300 border-indigo-500/50" class="nav-item">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
          Settings
        </a>
      </nav>

      <!-- Cluster Footer -->
      <div class="p-3 border-t border-slate-800/80 bg-slate-950/40">
        <div class="flex items-center gap-2 px-3 py-2 rounded-lg bg-slate-800/40 text-xs">
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          <div class="flex flex-col overflow-hidden">
            <span class="font-medium text-slate-200 truncate">{{ configService.config().cluster.name }}</span>
            <span class="text-[10px] text-slate-400">{{ configService.config().cluster.region }} / {{ configService.config().cluster.zone }}</span>
          </div>
        </div>
      </div>
    </aside>
  `,
  styles: [`
    .nav-item {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      padding: 0.5rem 0.75rem;
      border-radius: 0.5rem;
      font-size: 0.8125rem;
      font-weight: 500;
      color: #94a3b8;
      border: 1px solid transparent;
      transition: all 0.15s ease-in-out;
    }
    .nav-item:hover {
      background-color: rgba(30, 41, 59, 0.6);
      color: #f8fafc;
    }
  `]
})
export class SidebarComponent {
  readonly configService = inject(ConfigService);
  readonly appStore = inject(ApplicationStore);
}

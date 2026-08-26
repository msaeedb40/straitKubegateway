import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector: 'sg-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive],
  template: `
    <aside 
      class="w-64 bg-slate-950/95 border-r border-slate-800/80 flex flex-col justify-between shrink-0 select-none z-30 transition-all duration-300">
      
      <!-- Top Brand Header -->
      <div>
        <div class="h-16 flex items-center px-5 border-b border-slate-800/80 gap-3">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-sky-400 via-indigo-500 to-purple-600 flex items-center justify-center font-bold text-white shadow-lg shadow-sky-500/20 text-xs">
            SKG
          </div>
          <div>
            <div class="text-sm font-bold text-slate-100 tracking-wide flex items-center gap-1.5">
              <span>straitKube</span>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 font-mono">v1.0</span>
            </div>
            <div class="text-[10px] text-slate-400 font-medium tracking-wider uppercase">Transit Gateway</div>
          </div>
        </div>

        <!-- Navigation Categories -->
        <nav class="p-3 space-y-4 overflow-y-auto max-h-[calc(100vh-140px)]">
          <!-- Overview Section -->
          <div>
            <div class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider px-3 mb-1">Overview</div>
            <div class="space-y-0.5">
              <a routerLink="/dashboard" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" [routerLinkActiveOptions]="{ exact: true }" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>📊</span>
                <span>Dashboard</span>
              </a>
              <a routerLink="/gateways" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>🌐</span>
                <span>Gateways</span>
              </a>
              <a routerLink="/nodes" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>🖥️</span>
                <span>Nodes & CNI</span>
              </a>
              <a routerLink="/tunnels" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>🚇</span>
                <span>Transit Tunnels</span>
              </a>
              <a routerLink="/flows" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>⚡</span>
                <span>eBPF Flows</span>
              </a>
              <a routerLink="/topology" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>🕸️</span>
                <span>Topology Map</span>
              </a>
            </div>
          </div>

          <!-- Infrastructure Section -->
          <div>
            <div class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider px-3 mb-1">Infrastructure</div>
            <div class="space-y-0.5">
              <a routerLink="/services" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>⚖️</span>
                <span>Service LB</span>
              </a>
              <a routerLink="/endpoints" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>📍</span>
                <span>Endpoints</span>
              </a>
            </div>
          </div>

          <!-- Network & Kernel Section -->
          <div>
            <div class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider px-3 mb-1">Dataplane</div>
            <div class="space-y-0.5">
              <a routerLink="/ebpf" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>🧬</span>
                <span>eBPF Maps</span>
              </a>
              <a routerLink="/cni" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>🏷️</span>
                <span>CNI & IPAM</span>
              </a>
            </div>
          </div>

          <!-- Telemetry Section -->
          <div>
            <div class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider px-3 mb-1">Telemetry</div>
            <div class="space-y-0.5">
              <a routerLink="/events" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>🔔</span>
                <span>Events</span>
              </a>
              <a routerLink="/logs" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>📜</span>
                <span>Logs</span>
              </a>
              <a routerLink="/metrics" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>📈</span>
                <span>Metrics</span>
              </a>
            </div>
          </div>

          <!-- System Section -->
          <div>
            <div class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider px-3 mb-1">System</div>
            <div class="space-y-0.5">
              <a routerLink="/settings" routerLinkActive="bg-sky-500/10 text-sky-400 font-medium border-l-2 border-sky-400" class="flex items-center gap-2.5 px-3 py-1.5 rounded-r-lg text-xs text-slate-300 hover:bg-slate-900 hover:text-slate-100 transition">
                <span>⚙️</span>
                <span>Settings</span>
              </a>
            </div>
          </div>
        </nav>
      </div>

      <!-- Footer Cluster Info -->
      <div class="p-3.5 border-t border-slate-800/80 bg-slate-950/60">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
            <span class="text-xs font-semibold text-slate-200 font-mono">straitd</span>
          </div>
          <span class="text-[10px] text-emerald-400 font-mono font-medium">100% OK</span>
        </div>
        <div class="text-[10px] text-slate-500 mt-1 font-mono">Kernel 6.12 | NetKit</div>
      </div>
    </aside>
  `
})
export class SidebarComponent {}

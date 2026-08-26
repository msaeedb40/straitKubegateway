import { Component, HostListener, signal, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';

interface CommandItem {
  icon: string;
  title: string;
  category: string;
  route: string;
}

@Component({
  selector: 'sg-command-bar',
  standalone: true,
  imports: [CommonModule],
  template: `
    @if (isOpen()) {
      <div 
        (click)="close()"
        class="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-start justify-center pt-24 p-4 animate-fade-in">
        
        <div 
          (click)="$event.stopPropagation()"
          class="w-full max-w-xl bg-slate-900 border border-slate-700/80 rounded-2xl shadow-2xl overflow-hidden animate-scale-in">
          
          <!-- Input search box -->
          <div class="flex items-center px-4 py-3.5 border-b border-slate-800 gap-3">
            <svg class="w-4 h-4 text-sky-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
            </svg>
            <input
              type="text"
              #cmdInput
              [value]="query()"
              (input)="query.set(cmdInput.value)"
              placeholder="Type a command or jump to page (e.g., Gateways, eBPF, Flows)..."
              class="w-full bg-transparent text-sm text-slate-100 placeholder-slate-500 outline-none"
              autofocus />
            <kbd class="text-[10px] font-mono text-slate-400 bg-slate-800 px-2 py-0.5 rounded border border-slate-700">ESC</kbd>
          </div>

          <!-- Command results list -->
          <div class="max-h-72 overflow-y-auto p-2 space-y-1">
            @for (item of filteredCommands(); track item.route) {
              <div 
                (click)="navigate(item.route)"
                class="flex items-center justify-between px-3 py-2 rounded-xl text-xs text-slate-200 hover:bg-sky-500/15 hover:text-sky-300 transition cursor-pointer group">
                <div class="flex items-center gap-3">
                  <span>{{ item.icon }}</span>
                  <span class="font-medium">{{ item.title }}</span>
                </div>
                <span class="text-[10px] font-mono text-slate-500 uppercase">{{ item.category }}</span>
              </div>
            } @empty {
              <div class="py-6 text-center text-xs text-slate-500 italic">No matching pages found</div>
            }
          </div>
        </div>
      </div>
    }
  `
})
export class CommandBarComponent {
  private readonly router = inject(Router);
  readonly isOpen = signal<boolean>(false);
  readonly query = signal<string>('');

  readonly commands: CommandItem[] = [
    { icon: '📊', title: 'Dashboard & KPIs', category: 'Overview', route: '/dashboard' },
    { icon: '🌐', title: 'Gateway API Resources', category: 'Overview', route: '/gateways' },
    { icon: '🖥️', title: 'Nodes & CNI Status', category: 'Overview', route: '/nodes' },
    { icon: '🚇', title: 'Transit Tunnels & WireGuard', category: 'Overview', route: '/tunnels' },
    { icon: '⚡', title: 'eBPF Flow Monitor', category: 'Overview', route: '/flows' },
    { icon: '🕸️', title: 'Interactive Topology Map', category: 'Overview', route: '/topology' },
    { icon: '⚖️', title: 'Service Load Balancing', category: 'Infrastructure', route: '/services' },
    { icon: '📍', title: 'EndpointSlices', category: 'Infrastructure', route: '/endpoints' },
    { icon: '🧬', title: 'eBPF Map Inspector', category: 'Dataplane', route: '/ebpf' },
    { icon: '🏷️', title: 'CNI IPAM Allocations', category: 'Dataplane', route: '/cni' },
    { icon: '🔔', title: 'Network Events', category: 'Telemetry', route: '/events' },
    { icon: '📜', title: 'Structured Logs', category: 'Telemetry', route: '/logs' },
    { icon: '📈', title: 'Prometheus Metrics', category: 'Telemetry', route: '/metrics' },
    { icon: '⚙️', title: 'Settings & Control Plane Config', category: 'System', route: '/settings' }
  ];

  @HostListener('window:keydown', ['$event'])
  handleKeyDown(event: KeyboardEvent): void {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault();
      this.isOpen.update(v => !v);
    } else if (event.key === 'Escape' && this.isOpen()) {
      this.close();
    }
  }

  filteredCommands(): CommandItem[] {
    const q = this.query().toLowerCase().trim();
    if (!q) return this.commands;
    return this.commands.filter(c => 
      c.title.toLowerCase().includes(q) || c.category.toLowerCase().includes(q)
    );
  }

  navigate(route: string): void {
    this.router.navigateByUrl(route);
    this.close();
  }

  close(): void {
    this.isOpen.set(false);
    this.query.set('');
  }
}

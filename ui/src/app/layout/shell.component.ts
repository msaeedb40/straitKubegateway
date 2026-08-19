import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { SidebarComponent } from './sidebar/sidebar.component';
import { TopbarComponent } from './topbar/topbar.component';
import { BreadcrumbsComponent } from './breadcrumbs/breadcrumbs.component';
import { FooterComponent } from './footer/footer.component';
import { CommandPaletteComponent } from '../shared/components/command-palette/command-palette.component';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    SidebarComponent,
    TopbarComponent,
    BreadcrumbsComponent,
    FooterComponent,
    CommandPaletteComponent
  ],
  template: `
    <div class="flex h-screen w-screen overflow-hidden bg-slate-950 text-slate-100 font-sans antialiased selection:bg-indigo-500 selection:text-white">
      <!-- Modular Sidebar -->
      <app-sidebar></app-sidebar>

      <!-- Main Shell Area -->
      <main class="flex-1 flex flex-col min-w-0 overflow-hidden bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950">
        <!-- Modular Topbar -->
        <app-topbar></app-topbar>

        <!-- Sub-header Breadcrumb Bar -->
        <div class="px-8 py-2.5 border-b border-slate-800/40 bg-slate-950/20 flex items-center justify-between">
          <app-breadcrumbs></app-breadcrumbs>
        </div>

        <!-- Dynamic Router Outlet Viewport -->
        <div class="flex-1 overflow-y-auto p-8 custom-scrollbar">
          <router-outlet></router-outlet>
        </div>

        <!-- Modular Trace Footer -->
        <app-footer></app-footer>
      </main>

      <!-- Global Command Palette Modal (Cmd+K) -->
      <app-command-palette></app-command-palette>
    </div>
  `
})
export class ShellComponent {}

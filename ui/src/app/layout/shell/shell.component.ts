import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { HeaderComponent } from '../header/header.component';
import { CommandBarComponent } from '../command-bar/command-bar.component';
import { BreadcrumbsComponent } from '../breadcrumbs/breadcrumbs.component';
import { ToastOverlayComponent } from '../../shared/overlays/toast-overlay.component';

@Component({
  selector: 'sg-shell',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    SidebarComponent,
    HeaderComponent,
    CommandBarComponent,
    BreadcrumbsComponent,
    ToastOverlayComponent
  ],
  template: `
    <div class="flex h-screen bg-slate-950 text-slate-100 font-sans antialiased overflow-hidden">
      <!-- Sidebar -->
      <sg-sidebar />

      <!-- Main viewport area -->
      <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
        <sg-header />

        <!-- Scrollable view container -->
        <main class="flex-1 overflow-y-auto p-6 md:p-8">
          <sg-breadcrumbs />
          <router-outlet />
        </main>
      </div>

      <!-- Quick Command Bar Palette -->
      <sg-command-bar />

      <!-- Toast Overlay Container -->
      <sg-toast-overlay />
    </div>
  `
})
export class ShellComponent {}

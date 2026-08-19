import { Component, inject, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, NavigationEnd } from '@angular/router';
import { toSignal } from '@angular/core/rxjs-interop';
import { filter, map } from 'rxjs';

@Component({
  selector: 'app-breadcrumbs',
  standalone: true,
  imports: [CommonModule],
  template: `
    <nav class="flex items-center gap-2 text-xs text-slate-400 select-none" aria-label="Breadcrumb">
      <span>straitKube</span>
      <span class="text-slate-600">/</span>
      <span class="text-indigo-400 capitalize font-medium">{{ currentPath() }}</span>
    </nav>
  `
})
export class BreadcrumbsComponent {
  private readonly router = inject(Router);

  private readonly navEnd = toSignal(
    this.router.events.pipe(
      filter(e => e instanceof NavigationEnd),
      map(e => (e as NavigationEnd).urlAfterRedirects)
    ),
    { initialValue: this.router.url }
  );

  readonly currentPath = computed(() => {
    const raw = this.navEnd() || '/dashboard';
    return raw.replace('/', '').split('?')[0] || 'dashboard';
  });
}

import { Component, inject, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, NavigationEnd } from '@angular/router';
import { toSignal } from '@angular/core/rxjs-interop';
import { filter, map } from 'rxjs/operators';

@Component({
  selector: 'sg-breadcrumbs',
  standalone: true,
  imports: [CommonModule],
  template: `
    <nav class="flex items-center gap-2 text-xs text-slate-400 mb-4 select-none">
      <span class="text-slate-500">StraitKube</span>
      <span class="text-slate-600">/</span>
      <span class="text-slate-200 capitalize font-medium">{{ currentSection() }}</span>
    </nav>
  `
})
export class BreadcrumbsComponent {
  private readonly router = inject(Router);

  private readonly routeUrl = toSignal(
    this.router.events.pipe(
      filter((e): e is NavigationEnd => e instanceof NavigationEnd),
      map(e => e.urlAfterRedirects)
    ),
    { initialValue: this.router.url }
  );

  readonly currentSection = computed(() => {
    const url = this.routeUrl() || '/dashboard';
    const clean = url.split('?')[0].replace(/^\//, '');
    return clean || 'dashboard';
  });
}

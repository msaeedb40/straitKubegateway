import { Component, Input, ContentChild, TemplateRef, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface TableColumn {
  key: string;
  header: string;
  sortable?: boolean;
  align?: 'left' | 'center' | 'right';
  width?: string;
  mono?: boolean;
}

@Component({
  selector: 'app-data-table',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl overflow-hidden">
      <!-- Table Header Bar -->
      @if (title || searchEnabled) {
        <div class="p-4 border-b border-slate-800 flex items-center justify-between gap-4">
          @if (title) {
            <div>
              <h3 class="text-sm font-semibold text-white">{{ title }}</h3>
              @if (subtitle) {
                <p class="text-[11px] text-slate-400 mt-0.5">{{ subtitle }}</p>
              }
            </div>
          }
          @if (searchEnabled) {
            <div class="relative max-w-xs w-full">
              <input
                type="text"
                placeholder="Filter table..."
                [value]="filterQuery()"
                (input)="onFilterInput($event)"
                class="w-full bg-slate-950/60 text-xs text-slate-200 placeholder-slate-500 rounded-lg pl-8 pr-3 py-1.5 border border-slate-800 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
              <svg class="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          }
        </div>
      }

      <!-- Table View -->
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs" role="table">
          <thead class="text-slate-400 border-b border-slate-800/80 bg-slate-950/30">
            <tr>
              @for (col of columns; track col.key) {
                <th
                  class="p-3 font-medium cursor-pointer select-none"
                  [ngClass]="{
                    'text-right': col.align === 'right',
                    'text-center': col.align === 'center',
                    'hover:text-slate-200': col.sortable
                  }"
                  [style.width]="col.width"
                  (click)="col.sortable && sort(col.key)"
                  [attr.aria-sort]="getAriaSort(col.key)"
                >
                  <div class="flex items-center gap-1.5" [ngClass]="{ 'justify-end': col.align === 'right', 'justify-center': col.align === 'center' }">
                    <span>{{ col.header }}</span>
                    @if (col.sortable && sortKey() === col.key) {
                      <span class="text-indigo-400 font-mono">{{ sortOrder() === 'asc' ? '↑' : '↓' }}</span>
                    }
                  </div>
                </th>
              }
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/50 text-slate-300">
            @for (row of processedData(); track trackByFn(row)) {
              <tr class="hover:bg-slate-800/30 transition-colors">
                @for (col of columns; track col.key) {
                  <td
                    class="p-3"
                    [ngClass]="{
                      'font-mono': col.mono,
                      'text-right': col.align === 'right',
                      'text-center': col.align === 'center'
                    }"
                  >
                    @if (cellTemplate) {
                      <ng-container *ngTemplateOutlet="cellTemplate; context: { $implicit: row, column: col.key, value: row[col.key] }"></ng-container>
                    } @else {
                      {{ row[col.key] }}
                    }
                  </td>
                }
              </tr>
            }
            @if (processedData().length === 0) {
              <tr>
                <td [attr.colspan]="columns.length" class="p-8 text-center text-slate-500">
                  No records to display.
                </td>
              </tr>
            }
          </tbody>
        </table>
      </div>
    </div>
  `
})
export class DataTableComponent<T extends Record<string, any>> {
  @Input() columns: TableColumn[] = [];
  @Input() data: T[] = [];
  @Input() title?: string;
  @Input() subtitle?: string;
  @Input() searchEnabled: boolean = true;
  @Input() keyField: string = 'id';

  @ContentChild('cellTemplate') cellTemplate?: TemplateRef<any>;

  readonly filterQuery = signal<string>('');
  readonly sortKey = signal<string>('');
  readonly sortOrder = signal<'asc' | 'desc'>('asc');

  readonly processedData = computed(() => {
    let result = [...this.data];
    const q = this.filterQuery().toLowerCase().trim();

    if (q) {
      result = result.filter(item =>
        Object.values(item).some(val =>
          val !== null && val !== undefined && String(val).toLowerCase().includes(q)
        )
      );
    }

    const key = this.sortKey();
    if (key) {
      const order = this.sortOrder() === 'asc' ? 1 : -1;
      result.sort((a, b) => {
        const valA = a[key];
        const valB = b[key];
        if (typeof valA === 'number' && typeof valB === 'number') {
          return (valA - valB) * order;
        }
        return String(valA).localeCompare(String(valB)) * order;
      });
    }

    return result;
  });

  onFilterInput(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.filterQuery.set(input.value);
  }

  sort(key: string): void {
    if (this.sortKey() === key) {
      this.sortOrder.update(o => (o === 'asc' ? 'desc' : 'asc'));
    } else {
      this.sortKey.set(key);
      this.sortOrder.set('asc');
    }
  }

  getAriaSort(key: string): string {
    if (this.sortKey() !== key) return 'none';
    return this.sortOrder() === 'asc' ? 'ascending' : 'descending';
  }

  trackByFn(item: T): any {
    return item[this.keyField] || item;
  }
}

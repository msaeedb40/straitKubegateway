import { Component, input, signal, computed, output, TemplateRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SearchInputComponent } from '../components/search-input/search-input.component';

export interface TableColumn<T = any> {
  key: string;
  header: string;
  sortable?: boolean;
  width?: string;
  render?: (item: T) => string;
}

@Component({
  selector: 'sg-data-table',
  standalone: true,
  imports: [CommonModule, SearchInputComponent],
  template: `
    <div class="flex flex-col gap-3 w-full">
      <!-- Table Header controls (search & actions) -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        @if (searchable()) {
          <div class="max-w-xs w-full">
            <sg-search-input [placeholder]="searchPlaceholder()" (searchChange)="searchTerm.set($event)" />
          </div>
        }
        <div class="flex items-center gap-2">
          <ng-content select="[table-actions]"></ng-content>
        </div>
      </div>

      <!-- Table Container -->
      <div class="overflow-x-auto rounded-xl border border-slate-800/80 bg-slate-900/60 backdrop-blur-md">
        <table class="w-full text-left text-xs text-slate-300">
          <thead class="bg-slate-950/80 text-[11px] uppercase tracking-wider text-slate-400 border-b border-slate-800/80 select-none">
            <tr>
              @for (col of columns(); track col.key) {
                <th 
                  class="px-4 py-3 font-semibold"
                  [style.width]="col.width || 'auto'"
                  [class.cursor-pointer]="col.sortable"
                  [class.hover:text-slate-200]="col.sortable"
                  (click)="toggleSort(col)">
                  <div class="flex items-center gap-1.5">
                    <span>{{ col.header }}</span>
                    @if (col.sortable && sortKey() === col.key) {
                      <span class="text-sky-400 font-mono text-[10px]">
                        {{ sortAsc() ? '▲' : '▼' }}
                      </span>
                    }
                  </div>
                </th>
              }
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/40">
            @for (row of paginatedData(); track $index) {
              <tr 
                class="hover:bg-slate-800/40 transition group cursor-pointer"
                (click)="rowClicked.emit(row)">
                @for (col of columns(); track col.key) {
                  <td class="px-4 py-3 align-middle font-mono">
                    <ng-container *ngTemplateOutlet="cellTemplate() || defaultCell; context: { $implicit: row, column: col }">
                    </ng-container>
                  </td>
                }
              </tr>
            } @empty {
              <tr>
                <td [attr.colspan]="columns().length" class="px-4 py-8 text-center text-slate-500 italic">
                  No records found
                </td>
              </tr>
            }
          </tbody>
        </table>
      </div>

      <!-- Pagination footer -->
      @if (totalPages() > 1) {
        <div class="flex items-center justify-between text-xs text-slate-400 px-1">
          <div>
            Showing {{ (page() - 1) * pageSize() + 1 }} to {{ Math.min(page() * pageSize(), filteredData().length) }} of {{ filteredData().length }}
          </div>
          <div class="flex items-center gap-1">
            <button 
              [disabled]="page() <= 1"
              (click)="page.set(page() - 1)"
              class="px-2.5 py-1 rounded bg-slate-800 text-slate-300 disabled:opacity-40 hover:bg-slate-700 transition cursor-pointer">
              Previous
            </button>
            <span class="px-2 font-mono text-slate-200">{{ page() }} / {{ totalPages() }}</span>
            <button 
              [disabled]="page() >= totalPages()"
              (click)="page.set(page() + 1)"
              class="px-2.5 py-1 rounded bg-slate-800 text-slate-300 disabled:opacity-40 hover:bg-slate-700 transition cursor-pointer">
              Next
            </button>
          </div>
        </div>
      }
    </div>

    <!-- Default cell template -->
    <ng-template #defaultCell let-row let-column="column">
      <span class="text-slate-200">
        {{ column.render ? column.render(row) : (row[column.key] ?? '-') }}
      </span>
    </ng-template>
  `
})
export class DataTableComponent<T = any> {
  readonly columns = input.required<TableColumn<T>[]>();
  readonly data = input.required<T[]>();
  readonly searchable = input<boolean>(true);
  readonly searchPlaceholder = input<string>('Filter table...');
  readonly pageSize = input<number>(10);
  readonly cellTemplate = input<TemplateRef<any> | null>(null);

  readonly rowClicked = output<T>();

  readonly searchTerm = signal<string>('');
  readonly sortKey = signal<string>('');
  readonly sortAsc = signal<boolean>(true);
  readonly page = signal<number>(1);
  readonly Math = Math;

  readonly filteredData = computed(() => {
    let result = [...this.data()];
    const term = this.searchTerm().toLowerCase().trim();

    if (term) {
      result = result.filter(item => {
        return Object.values(item as any).some(val => 
          String(val).toLowerCase().includes(term)
        );
      });
    }

    const key = this.sortKey();
    if (key) {
      const asc = this.sortAsc();
      result.sort((a: any, b: any) => {
        const valA = a[key];
        const valB = b[key];
        if (valA === valB) return 0;
        if (valA == null) return 1;
        if (valB == null) return -1;
        return (valA > valB ? 1 : -1) * (asc ? 1 : -1);
      });
    }

    return result;
  });

  readonly totalPages = computed(() => {
    return Math.ceil(this.filteredData().length / this.pageSize()) || 1;
  });

  readonly paginatedData = computed(() => {
    const start = (this.page() - 1) * this.pageSize();
    return this.filteredData().slice(start, start + this.pageSize());
  });

  toggleSort(col: TableColumn<T>): void {
    if (!col.sortable) return;
    if (this.sortKey() === col.key) {
      this.sortAsc.update(a => !a);
    } else {
      this.sortKey.set(col.key);
      this.sortAsc.set(true);
    }
  }
}

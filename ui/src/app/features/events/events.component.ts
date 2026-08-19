import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { DataTableComponent, TableColumn } from '../../shared/components/data-table/data-table.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';

@Component({
  selector: 'app-events',
  standalone: true,
  imports: [CommonModule, DataTableComponent, StatusBadgeComponent],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Cluster Audit Events & Alerts</h1>
          <p class="text-xs text-slate-400 mt-1">Real-time gateway lifecycle events, policy denials, and tunnel health logs.</p>
        </div>
      </div>

      <app-data-table
        title="Real-Time Event Stream"
        subtitle="Search across components, messages, and resource references"
        [columns]="eventColumns"
        [data]="state.events()"
      >
        <ng-template #cellTemplate let-row let-col="column" let-val="value">
          @if (col === 'type') {
            <app-status-badge [status]="val">{{ val }}</app-status-badge>
          } @else if (col === 'timestamp') {
            <span class="text-slate-400">{{ val | date:'yyyy-MM-dd HH:mm:ss' }}</span>
          } @else if (col === 'resourceRef') {
            <span class="text-indigo-300 font-mono">{{ val }}</span>
          } @else if (col === 'component') {
            <span class="text-slate-400 font-mono">{{ val }}</span>
          } @else {
            <span class="text-slate-200">{{ val }}</span>
          }
        </ng-template>
      </app-data-table>
    </div>
  `
})
export class EventsComponent {
  readonly state = inject(StateService);

  readonly eventColumns: TableColumn[] = [
    { key: 'type', header: 'Severity', sortable: true, width: '120px' },
    { key: 'component', header: 'Component', sortable: true, mono: true, width: '160px' },
    { key: 'message', header: 'Message', sortable: true },
    { key: 'resourceRef', header: 'Resource Ref', sortable: true, mono: true },
    { key: 'timestamp', header: 'Timestamp', sortable: true, mono: true, width: '180px' }
  ];
}

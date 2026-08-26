import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClientService } from '../../core/api/api-client';
import { EventInfo } from '../../core/models/event.model';
import { CardComponent } from '../../shared/components/card/card.component';
import { DataTableComponent, TableColumn } from '../../shared/tables/data-table.component';

@Component({
  selector: 'app-events',
  standalone: true,
  imports: [
    CommonModule,
    CardComponent,
    DataTableComponent
  ],
  template: `
    <div class="space-y-6">
      <!-- Header Banner -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-bold text-slate-100">Telemetry & Network Events</h1>
          <p class="text-xs text-slate-400 mt-1">
            Real-time audit log with unified 11-field metadata tuple tracking across control plane and kernel
          </p>
        </div>
      </div>

      <!-- Events Table Card -->
      <sg-card title="System & Network Event Log" subtitle="Comprehensive audit and state transition feed">
        <sg-data-table 
          [columns]="eventColumns" 
          [data]="api.events()"
          searchPlaceholder="Filter events by component, message, IDs..." />
      </sg-card>
    </div>
  `
})
export class EventsComponent {
  readonly api = inject(ApiClientService);

  readonly eventColumns: TableColumn<EventInfo>[] = [
    { key: 'id', header: 'Event ID', sortable: true },
    { key: 'component', header: 'Subsystem', sortable: true },
    { key: 'type', header: 'Severity', sortable: true },
    { key: 'message', header: 'Message', sortable: true },
    { key: 'clusterID', header: 'Cluster ID', sortable: true },
    { key: 'timestamp', header: 'Time', sortable: true }
  ];
}

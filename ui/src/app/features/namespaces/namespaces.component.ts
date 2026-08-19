import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { DataTableComponent, TableColumn } from '../../shared/components/data-table/data-table.component';

@Component({
  selector: 'app-namespaces',
  standalone: true,
  imports: [CommonModule, DataTableComponent],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Namespaces & Multi-Tenancy</h1>
          <p class="text-xs text-slate-400 mt-1">Tenant boundaries, segment assignments, and gateway bindings per namespace.</p>
        </div>
      </div>

      <app-data-table
        title="Active Kubernetes Namespaces"
        subtitle="Network segmentation and policy enforcement scope"
        [columns]="namespaceColumns"
        [data]="namespacesData"
      ></app-data-table>
    </div>
  `
})
export class NamespacesComponent {
  readonly state = inject(StateService);

  readonly namespaceColumns: TableColumn[] = [
    { key: 'name', header: 'Namespace Name', sortable: true },
    { key: 'segment', header: 'Assigned Segment', sortable: true, mono: true },
    { key: 'gateways', header: 'Bound Gateways', sortable: true, align: 'center', mono: true },
    { key: 'policies', header: 'Active Policies', sortable: true, align: 'center', mono: true },
    { key: 'status', header: 'Isolation Status', sortable: true }
  ];

  readonly namespacesData = [
    { id: '1', name: 'default', segment: 'Segment 0 (Global)', gateways: 1, policies: 4, status: 'Active' },
    { id: '2', name: 'strait-system', segment: 'Segment 0 (System)', gateways: 2, policies: 8, status: 'Active' },
    { id: '3', name: 'production', segment: 'Segment 100 (Isolated)', gateways: 1, policies: 12, status: 'Strict Isolation' },
    { id: '4', name: 'staging', segment: 'Segment 200 (Staging)', gateways: 0, policies: 6, status: 'Strict Isolation' },
    { id: '5', name: 'payments', segment: 'Segment 100 (PCI-DSS)', gateways: 1, policies: 14, status: 'Strict Isolation' }
  ];
}

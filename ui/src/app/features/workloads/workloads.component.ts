import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../../core/state/state.service';
import { DataTableComponent, TableColumn } from '../../shared/components/data-table/data-table.component';

@Component({
  selector: 'app-workloads',
  standalone: true,
  imports: [CommonModule, DataTableComponent],
  template: `
    <div class="space-y-6 max-w-7xl mx-auto">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">Workloads & Pod Endpoints</h1>
          <p class="text-xs text-slate-400 mt-1">Pod IPAM assignments, NetKit interface bindings, and eBPF cgroup redirection.</p>
        </div>
      </div>

      <app-data-table
        title="Discovered Pod Endpoints"
        subtitle="Live cgroup and NetKit host veth attachments"
        [columns]="workloadColumns"
        [data]="workloadsData"
      ></app-data-table>
    </div>
  `
})
export class WorkloadsComponent {
  readonly state = inject(StateService);

  readonly workloadColumns: TableColumn[] = [
    { key: 'podName', header: 'Pod Name', sortable: true },
    { key: 'namespace', header: 'Namespace', sortable: true },
    { key: 'node', header: 'Host Node', sortable: true },
    { key: 'podIP', header: 'Pod IP', sortable: true, mono: true },
    { key: 'netkitIf', header: 'eBPF NetKit Interface', sortable: true, mono: true },
    { key: 'status', header: 'Status', sortable: true }
  ];

  readonly workloadsData = [
    { id: '1', podName: 'frontend-web-74d9f68b7-8klp2', namespace: 'default', node: 'strait-k8s-worker-01', podIP: '10.244.1.12', netkitIf: 'netkit-host-0', status: 'Running' },
    { id: '2', podName: 'payment-processor-58c894b68-x9q41', namespace: 'payments', node: 'strait-k8s-worker-02', podIP: '10.244.2.35', netkitIf: 'netkit-host-1', status: 'Running' },
    { id: '3', podName: 'user-auth-service-678fdb944-v4k99', namespace: 'default', node: 'strait-k8s-worker-01', podIP: '10.244.1.44', netkitIf: 'netkit-host-2', status: 'Running' },
    { id: '4', podName: 'redis-cache-cluster-0', namespace: 'production', node: 'strait-k8s-worker-02', podIP: '10.244.2.18', netkitIf: 'netkit-host-3', status: 'Running' }
  ];
}

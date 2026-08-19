import { Component, Input, ContentChild, TemplateRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ScrollingModule } from '@angular/cdk/scrolling';

@Component({
  selector: 'app-virtual-list',
  standalone: true,
  imports: [CommonModule, ScrollingModule],
  template: `
    <div class="rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl overflow-hidden">
      @if (headerTitle) {
        <div class="p-4 border-b border-slate-800 flex items-center justify-between">
          <h3 class="text-sm font-semibold text-white">{{ headerTitle }}</h3>
          <span class="text-xs font-mono text-slate-400">{{ items.length }} items</span>
        </div>
      }

      <cdk-virtual-scroll-viewport [itemSize]="itemSize" [style.height]="viewportHeight" class="custom-scrollbar w-full">
        <div *cdkVirtualFor="let item of items; trackBy: trackByFn" [style.height.px]="itemSize" class="border-b border-slate-800/40 hover:bg-slate-800/30 transition-colors flex items-center px-4">
          @if (itemTemplate) {
            <ng-container *ngTemplateOutlet="itemTemplate; context: { $implicit: item }"></ng-container>
          } @else {
            <div class="text-xs text-slate-300 font-mono truncate">{{ item | json }}</div>
          }
        </div>
      </cdk-virtual-scroll-viewport>
    </div>
  `
})
export class VirtualListComponent<T> {
  @Input() items: T[] = [];
  @Input() itemSize: number = 44;
  @Input() viewportHeight: string = '400px';
  @Input() headerTitle?: string;
  @Input() keyField: string = 'id';

  @ContentChild('itemTemplate') itemTemplate?: TemplateRef<any>;

  trackByFn(_: number, item: any): any {
    return item?.[this.keyField] ?? item;
  }
}

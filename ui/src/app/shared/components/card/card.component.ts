import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-card',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="bg-slate-900/70 border border-slate-800/80 rounded-xl p-5 backdrop-blur-md shadow-sm transition hover:border-slate-700/80 flex flex-col justify-between"
         [ngClass]="customClass()">
      <div>
        @if (title() || subtitle()) {
          <div class="flex items-center justify-between mb-3">
            <div>
              @if (title()) {
                <h3 class="text-sm font-semibold text-slate-200 tracking-tight">{{ title() }}</h3>
              }
              @if (subtitle()) {
                <p class="text-xs text-slate-400 mt-0.5">{{ subtitle() }}</p>
              }
            </div>
            <ng-content select="[card-action]"></ng-content>
          </div>
        }
        <ng-content></ng-content>
      </div>
      @if (hasFooter()) {
        <div class="mt-4 pt-3 border-t border-slate-800/60 text-xs text-slate-400">
          <ng-content select="[card-footer]"></ng-content>
        </div>
      }
    </div>
  `
})
export class CardComponent {
  readonly title = input<string>('');
  readonly subtitle = input<string>('');
  readonly customClass = input<string>('');
  readonly hasFooter = input<boolean>(false);
}

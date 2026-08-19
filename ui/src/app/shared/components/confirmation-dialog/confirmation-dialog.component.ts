import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FocusTrapDirective } from '../../directives/focus-trap.directive';

@Component({
  selector: 'app-confirmation-dialog',
  standalone: true,
  imports: [CommonModule, FocusTrapDirective],
  template: `
    @if (isOpen) {
      <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm" role="dialog" aria-modal="true">
        <div appFocusTrap class="w-full max-w-md p-6 rounded-2xl bg-slate-900 border border-slate-800 shadow-2xl space-y-4">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-xl" [ngClass]="isDestructive ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20' : 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20'">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <div>
              <h3 class="text-base font-semibold text-white">{{ title }}</h3>
              <p class="text-xs text-slate-400 mt-0.5">{{ message }}</p>
            </div>
          </div>

          <div class="flex items-center justify-end gap-3 pt-2">
            <button
              (click)="cancel.emit()"
              class="px-3.5 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-xs font-medium text-slate-300 border border-slate-700 transition-all"
            >
              Cancel
            </button>
            <button
              (click)="confirm.emit()"
              class="px-3.5 py-1.5 rounded-xl text-xs font-semibold text-white shadow-lg transition-all"
              [ngClass]="isDestructive ? 'bg-rose-600 hover:bg-rose-500 shadow-rose-600/20' : 'bg-indigo-600 hover:bg-indigo-500 shadow-indigo-600/20'"
            >
              {{ confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    }
  `
})
export class ConfirmationDialogComponent {
  @Input() isOpen: boolean = false;
  @Input() title: string = 'Confirm Action';
  @Input() message: string = 'Are you sure you want to proceed?';
  @Input() confirmLabel: string = 'Confirm';
  @Input() isDestructive: boolean = false;

  @Output() confirm = new EventEmitter<void>();
  @Output() cancel = new EventEmitter<void>();
}

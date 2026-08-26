import { Component, input, output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-search-input',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="relative w-full">
      <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
        </svg>
      </div>
      <input
        type="text"
        [value]="value()"
        [placeholder]="placeholder()"
        (input)="onInput($event)"
        class="w-full bg-slate-900/90 border border-slate-800 focus:border-sky-500/60 focus:ring-1 focus:ring-sky-500/50 rounded-lg pl-9 pr-8 py-1.5 text-xs text-slate-200 placeholder-slate-500 transition outline-none" />
      @if (value()) {
        <button 
          (click)="clear()"
          class="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-500 hover:text-slate-300">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      }
    </div>
  `
})
export class SearchInputComponent {
  readonly placeholder = input<string>('Search resources...');
  readonly value = signal<string>('');
  readonly searchChange = output<string>();

  onInput(event: Event): void {
    const val = (event.target as HTMLInputElement).value;
    this.value.set(val);
    this.searchChange.emit(val);
  }

  clear(): void {
    this.value.set('');
    this.searchChange.emit('');
  }
}

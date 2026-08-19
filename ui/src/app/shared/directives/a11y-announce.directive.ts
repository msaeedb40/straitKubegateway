import { Directive, Input, inject, ElementRef, effect } from '@angular/core';

@Directive({
  selector: '[appA11yAnnounce]',
  standalone: true
})
export class A11yAnnounceDirective {
  private readonly el = inject(ElementRef);
  @Input() appA11yAnnounce: string = '';

  constructor() {
    this.el.nativeElement.setAttribute('aria-live', 'polite');
    this.el.nativeElement.setAttribute('aria-atomic', 'true');
    this.el.nativeElement.classList.add('sr-only');
  }
}

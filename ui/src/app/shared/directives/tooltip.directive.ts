import { Directive, ElementRef, HostListener, input, Renderer2, inject } from '@angular/core';

@Directive({
  selector: '[sgTooltip]',
  standalone: true
})
export class TooltipDirective {
  readonly sgTooltip = input.required<string>();
  private readonly el = inject(ElementRef);
  private readonly renderer = inject(Renderer2);
  private tooltipEl?: HTMLElement;

  @HostListener('mouseenter')
  onMouseEnter(): void {
    if (!this.sgTooltip()) return;
    this.createTooltip();
  }

  @HostListener('mouseleave')
  onMouseLeave(): void {
    this.removeTooltip();
  }

  private createTooltip(): void {
    this.tooltipEl = this.renderer.createElement('div');
    this.renderer.appendChild(
      this.tooltipEl,
      this.renderer.createText(this.sgTooltip())
    );

    const classes = [
      'fixed', 'z-50', 'px-2', 'py-1', 'text-[11px]', 'font-medium',
      'text-slate-100', 'bg-slate-900', 'border', 'border-slate-700',
      'rounded-md', 'shadow-xl', 'pointer-events-none', 'transition-opacity'
    ];
    classes.forEach(c => this.renderer.addClass(this.tooltipEl, c));

    this.renderer.appendChild(document.body, this.tooltipEl);

    const hostRect = this.el.nativeElement.getBoundingClientRect();
    const tooltipRect = this.tooltipEl!.getBoundingClientRect();

    const top = hostRect.top - tooltipRect.height - 6;
    const left = hostRect.left + (hostRect.width / 2) - (tooltipRect.width / 2);

    this.renderer.setStyle(this.tooltipEl, 'top', `${Math.max(4, top)}px`);
    this.renderer.setStyle(this.tooltipEl, 'left', `${Math.max(4, left)}px`);
  }

  private removeTooltip(): void {
    if (this.tooltipEl) {
      this.renderer.removeChild(document.body, this.tooltipEl);
      this.tooltipEl = undefined;
    }
  }
}

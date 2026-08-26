import { Directive, HostListener, input, inject } from '@angular/core';
import { NotificationService } from '../../core/services/notification.service';

@Directive({
  selector: '[sgCopyClipboard]',
  standalone: true
})
export class CopyClipboardDirective {
  readonly sgCopyClipboard = input.required<string>();
  private readonly notif = inject(NotificationService);

  @HostListener('click', ['$event'])
  async onClick(event: MouseEvent): Promise<void> {
    event.stopPropagation();
    try {
      await navigator.clipboard.writeText(this.sgCopyClipboard());
      this.notif.show({
        type: 'success',
        title: 'Copied to clipboard',
        message: this.sgCopyClipboard(),
        autoCloseMs: 2500
      });
    } catch {
      // fallback
    }
  }
}

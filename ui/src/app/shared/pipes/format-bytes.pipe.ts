import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'formatBytes',
  standalone: true
})
export class FormatBytesPipe implements PipeTransform {
  transform(bytes: number | null | undefined, decimals: number = 2): string {
    if (bytes === 0 || bytes === null || bytes === undefined) return '0 B';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    const normalizedIndex = Math.min(i, sizes.length - 1);
    return `${parseFloat((bytes / Math.pow(k, normalizedIndex)).toFixed(dm))} ${sizes[normalizedIndex]}`;
  }
}

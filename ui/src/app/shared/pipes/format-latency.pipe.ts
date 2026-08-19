import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'formatLatency',
  standalone: true
})
export class FormatLatencyPipe implements PipeTransform {
  transform(latencyValue: number | null | undefined, unit: 'us' | 'ms' = 'ms'): string {
    if (latencyValue === null || latencyValue === undefined) return '-';

    if (unit === 'us') {
      if (latencyValue < 1000) {
        return `${latencyValue.toFixed(0)} µs`;
      }
      return `${(latencyValue / 1000).toFixed(2)} ms`;
    }

    if (latencyValue < 1) {
      return `${(latencyValue * 1000).toFixed(0)} µs`;
    }
    return `${latencyValue.toFixed(2)} ms`;
  }
}

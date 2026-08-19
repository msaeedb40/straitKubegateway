import { LatencyPercentileMetrics } from '../../core/models/models';

export function calculatePercentiles(values: number[]): LatencyPercentileMetrics {
  if (!values || values.length === 0) {
    return { p50: 0, p90: 0, p95: 0, p99: 0, p99_9: 0, max: 0 };
  }

  const sorted = [...values].sort((a, b) => a - b);
  const getP = (p: number) => {
    const idx = Math.ceil((p / 100) * sorted.length) - 1;
    return sorted[Math.max(0, Math.min(idx, sorted.length - 1))];
  };

  return {
    p50: getP(50),
    p90: getP(90),
    p95: getP(95),
    p99: getP(99),
    p99_9: getP(99.9),
    max: sorted[sorted.length - 1]
  };
}

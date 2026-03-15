import { useQuery } from '@tanstack/react-query'
import { getBenchmark } from '../api/benchmark'
import type { BenchmarkResult } from '../api/benchmark'

/**
 * usePriceBenchmark fetches the AI market benchmark for a shopping item.
 *
 * @param itemId - UUID of the shopping item (null/undefined to disable).
 * @param enabled - explicit enable flag; allows lazy loading (e.g. on user click).
 *
 * Stale time is 2h to match the backend cache TTL.
 */
export function usePriceBenchmark(itemId: string | null | undefined, enabled: boolean) {
  return useQuery<BenchmarkResult>({
    queryKey: ['price-benchmark', itemId],
    queryFn: () => getBenchmark(itemId!),
    enabled: !!itemId && enabled,
    staleTime: 2 * 60 * 60 * 1000, // 2h — matches backend cache TTL
  })
}

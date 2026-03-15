import { useQuery } from '@tanstack/react-query'
import { getPriceStats } from '../api/priceanalytics'
import type { PriceStats } from '../api/priceanalytics'

export function usePriceStats(missionId: string, itemId: string): { stats: PriceStats | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['price-stats', missionId, itemId],
    queryFn: () => getPriceStats(missionId, itemId),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: Boolean(missionId && itemId),
  })

  return { stats: data ?? null, isLoading }
}

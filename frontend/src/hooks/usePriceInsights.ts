import { useQuery } from '@tanstack/react-query'
import { getPriceInsights } from '../api/priceinsights'
import type { PriceInsights } from '../api/priceinsights'

export function usePriceInsights(missionId: string, itemId: string): { insights: PriceInsights | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['price-insights', missionId, itemId],
    queryFn: () => getPriceInsights(missionId, itemId),
    staleTime: 2 * 60 * 60 * 1000, // 2 hours
    enabled: Boolean(missionId && itemId),
  })

  return { insights: data ?? null, isLoading }
}

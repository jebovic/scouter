import { useQuery } from '@tanstack/react-query'
import { getCompetitorPrices } from '../api/competitorprice'

export function useCompetitorPrices(missionId: string, itemId: string, enabled = true) {
  return useQuery({
    queryKey: ['competitor-prices', missionId, itemId],
    queryFn: () => getCompetitorPrices(missionId, itemId),
    enabled: enabled && !!missionId && !!itemId,
    staleTime: 30 * 60 * 1000, // 30 minutes — mirrors backend cache TTL
    retry: false,
  })
}

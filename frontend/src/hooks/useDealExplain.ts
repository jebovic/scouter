import { useQuery } from '@tanstack/react-query'
import { fetchDealExplanation } from '../api/dealExplain'
import type { DealExplanation } from '../api/dealExplain'

export function useDealExplain(
  itemId: string | null,
  score?: number,
  trend?: string,
) {
  return useQuery<DealExplanation>({
    queryKey: ['deal-explain', itemId, score, trend],
    queryFn: () => fetchDealExplanation(itemId!, score, trend),
    enabled: !!itemId,
    staleTime: 60 * 60 * 1000, // 1h — matches backend cache TTL
  })
}

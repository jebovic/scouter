import { useQuery } from '@tanstack/react-query'
import { getDeals } from '../api/dealaggregator'

export function useDealAggregator(missionId: string, itemId: string) {
  return useQuery({
    queryKey: ['deals', missionId, itemId],
    queryFn: () => getDeals(missionId, itemId),
    staleTime: 30 * 60 * 1000,
    enabled: !!missionId && !!itemId,
  })
}

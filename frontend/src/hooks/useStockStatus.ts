import { useQuery } from '@tanstack/react-query'
import { getStockStatus } from '../api/stockalert'

export function useStockStatus(missionId: string, itemId: string) {
  return useQuery({
    queryKey: ['stock-status', missionId, itemId],
    queryFn: () => getStockStatus(missionId, itemId),
    enabled: !!missionId && !!itemId,
    staleTime: 30 * 60 * 1000, // 30 minutes — mirrors backend cache TTL
    retry: false,
  })
}

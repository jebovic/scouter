import { useQuery } from '@tanstack/react-query'
import { getConditionPricing } from '../api/conditionpricing'

export function useConditionPricing(missionId: string, itemId: string, enabled = true) {
  return useQuery({
    queryKey: ['condition-pricing', missionId, itemId],
    queryFn: () => getConditionPricing(missionId, itemId),
    enabled: enabled && !!missionId && !!itemId,
    staleTime: 2 * 60 * 60 * 1000, // 2 hours — mirrors backend cache TTL
    retry: false,
  })
}

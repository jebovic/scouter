import { useQuery } from '@tanstack/react-query'
import { getCoupons } from '../api/couponfinder'

export function useCoupons(missionId: string, itemId: string, enabled = true) {
  return useQuery({
    queryKey: ['coupons', missionId, itemId],
    queryFn: () => getCoupons(missionId, itemId),
    enabled: enabled && !!missionId && !!itemId,
    staleTime: 60 * 60 * 1000, // 1 hour — mirrors backend cache TTL
    retry: false,
  })
}

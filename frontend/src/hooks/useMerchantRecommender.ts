import { useQuery } from '@tanstack/react-query'
import { getMerchantRecommendations } from '../api/merchantrecommender'
import type { MerchantRecommenderResponse } from '../api/merchantrecommender'

/**
 * useMerchantRecommender fetches merchant recommendations for a shopping item.
 *
 * @param missionId - UUID of the mission (null/undefined to disable).
 * @param itemId - UUID of the shopping item (null/undefined to disable).
 * @param enabled - explicit enable flag; allows lazy loading (e.g. on user click).
 *
 * Stale time is 30 min to match the backend cache TTL.
 */
export function useMerchantRecommender(
  missionId: string | null | undefined,
  itemId: string | null | undefined,
  enabled: boolean = true,
) {
  return useQuery<MerchantRecommenderResponse>({
    queryKey: ['merchant-recommender', missionId, itemId],
    queryFn: () => getMerchantRecommendations(missionId!, itemId!),
    enabled: !!missionId && !!itemId && enabled,
    staleTime: 30 * 60 * 1000, // 30 min — matches backend cache TTL
  })
}

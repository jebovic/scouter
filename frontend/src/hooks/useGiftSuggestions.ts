import { useQuery } from '@tanstack/react-query'
import { getGiftSuggestions, type GiftSuggestionsParams } from '../api/giftfinder'

export function useGiftSuggestions(
  missionId: string,
  params: GiftSuggestionsParams = {},
) {
  return useQuery({
    queryKey: ['gift-suggestions', missionId, params],
    queryFn: () => getGiftSuggestions(missionId, params),
    enabled: !!missionId,
    staleTime: 30 * 60 * 1000, // 30 minutes — matches backend cache TTL
    retry: false,
  })
}

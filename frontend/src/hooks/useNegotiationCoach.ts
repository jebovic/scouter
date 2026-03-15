import { useQuery } from '@tanstack/react-query'
import { getNegotiationAdvice } from '../api/negotiation_coach'
import type { NegotiationAdvice } from '../api/negotiation_coach'

/**
 * useNegotiationAdvice fetches AI negotiation tips for a shopping item.
 *
 * @param itemId - UUID of the shopping item (null/undefined to disable).
 * @param enabled - explicit enable flag; allows lazy loading on user click.
 *
 * Stale time is 24h to match the backend cache TTL.
 */
export function useNegotiationAdvice(itemId: string | null | undefined, enabled: boolean) {
  return useQuery<NegotiationAdvice>({
    queryKey: ['negotiation-advice', itemId],
    queryFn: () => getNegotiationAdvice(itemId!),
    enabled: !!itemId && enabled,
    staleTime: 24 * 60 * 60 * 1000, // 24h — matches backend cache TTL
  })
}

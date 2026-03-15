import { useQuery } from '@tanstack/react-query'
import { getNegotiationScriptContextual } from '../api/negotiationscript'
import type { NegotiationScriptResponse } from '../api/negotiationscript'

/**
 * useNegotiationScriptContextual fetches a context-aware French negotiation
 * script for a shopping item. Script sections are derived from price history
 * and merchant context rather than FNV-seeded templates.
 *
 * Stale time is 1h to match the backend cache TTL.
 */
export function useNegotiationScriptContextual(
  missionId: string | null | undefined,
  itemId: string | null | undefined,
): { data: NegotiationScriptResponse | undefined; isLoading: boolean } {
  const { data, isLoading } = useQuery<NegotiationScriptResponse>({
    queryKey: ['negotiation-script-contextual', missionId, itemId],
    queryFn: () => getNegotiationScriptContextual(missionId!, itemId!),
    enabled: Boolean(missionId && itemId),
    staleTime: 60 * 60 * 1000, // 1h — matches backend cache TTL
  })
  return { data, isLoading }
}

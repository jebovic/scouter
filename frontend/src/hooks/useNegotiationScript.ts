import { useQuery } from '@tanstack/react-query'
import { getNegotiationScript } from '../api/negotiationsim'
import type { NegotiationScript } from '../api/negotiationsim'

/**
 * useNegotiationScript fetches a deterministic negotiation script for a shopping item.
 *
 * @param missionId - UUID of the parent mission.
 * @param itemId    - UUID of the shopping item (null/undefined to disable).
 * @param enabled   - explicit enable flag; allows lazy loading on user click.
 *
 * Stale time is 1h to match the backend cache TTL.
 */
export function useNegotiationScript(
  missionId: string | null | undefined,
  itemId: string | null | undefined,
  enabled: boolean,
) {
  return useQuery<NegotiationScript>({
    queryKey: ['negotiation-script', missionId, itemId],
    queryFn: () => getNegotiationScript(missionId!, itemId!),
    enabled: !!missionId && !!itemId && enabled,
    staleTime: 60 * 60 * 1000, // 1h — matches backend cache TTL
  })
}

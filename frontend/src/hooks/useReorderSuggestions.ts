import { useQuery } from '@tanstack/react-query'
import { fetchReorderSuggestions } from '../api/reordersuggestion'
import type { ReorderSuggestionsResponse } from '../api/reordersuggestion'

/** Fetches smart reorder suggestions for a mission. Refetches every 10 minutes to match backend cache TTL. */
export function useReorderSuggestions(missionId: string | undefined) {
  return useQuery<ReorderSuggestionsResponse>({
    queryKey: ['reorderSuggestions', missionId],
    queryFn: () => fetchReorderSuggestions(missionId!),
    enabled: Boolean(missionId),
    staleTime: 10 * 60 * 1000,
    refetchInterval: 10 * 60 * 1000,
  })
}

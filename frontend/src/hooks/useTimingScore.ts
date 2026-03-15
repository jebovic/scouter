import { useQuery } from '@tanstack/react-query'
import { getTimingScore } from '../api/timingrecommender'

export function useTimingScore(missionId: string, itemId: string) {
  return useQuery({
    queryKey: ['timing-score', missionId, itemId],
    queryFn: () => getTimingScore(missionId, itemId),
    staleTime: 30 * 60 * 1000, // 30min — matches server cache TTL
    enabled: Boolean(missionId && itemId),
  })
}

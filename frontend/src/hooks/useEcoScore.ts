import { useQuery } from '@tanstack/react-query'
import { getEcoScore } from '../api/ecoscore'

export function useEcoScore(missionId: string) {
  return useQuery({
    queryKey: ['eco-score', missionId],
    queryFn: () => getEcoScore(missionId),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!missionId,
  })
}

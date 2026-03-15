import { useQuery } from '@tanstack/react-query'
import { getVoteSummary } from '../api/wishlistvote'

export function useVoteSummary(missionId: string) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['vote-summary', missionId],
    queryFn: () => getVoteSummary(missionId),
    enabled: !!missionId,
    staleTime: 10 * 60 * 1000, // 10 min — matches backend cache TTL
  })

  return { data, isLoading, error }
}

import { useQuery } from '@tanstack/react-query'
import { fetchComparisonScore, type ComparisonReport } from '../api/comparisonscore'

export function useComparisonScore(missionId: string) {
  return useQuery<ComparisonReport>({
    queryKey: ['comparisonScore', missionId],
    queryFn: () => fetchComparisonScore(missionId),
    staleTime: 20 * 60 * 1000, // 20 minutes
    enabled: !!missionId,
  })
}

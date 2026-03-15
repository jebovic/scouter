import { useQuery } from '@tanstack/react-query'
import { getRegretAnalysis } from '../api/regretanalyzer'
import type { RegretAnalyzerResponse } from '../api/regretanalyzer'

export function useRegretAnalyzer(missionId: string | undefined) {
  const { data = null, isLoading, error } = useQuery({
    queryKey: ['regretAnalysis', missionId],
    queryFn: () => (missionId ? getRegretAnalysis(missionId) : Promise.reject(new Error('No mission ID'))),
    enabled: Boolean(missionId),
    staleTime: 60 * 60 * 1000, // 1 hour
  })

  return { data: data as RegretAnalyzerResponse | null, isLoading, error }
}

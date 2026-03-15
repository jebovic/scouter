import { useQuery } from '@tanstack/react-query'
import { getBudgetAnalysis } from '../api/budgetrec'

export function useBudgetAnalysis(missionId: string) {
  return useQuery({
    queryKey: ['budget-analysis', missionId],
    queryFn: () => getBudgetAnalysis(missionId),
    staleTime: 5 * 60 * 1000,
    enabled: !!missionId,
  })
}

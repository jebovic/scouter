import { useQuery } from '@tanstack/react-query'
import type { BudgetAdvice } from '../api/budgetadvisor'
import { getBudgetAdvice } from '../api/budgetadvisor'

export function useBudgetAdvisor(missionId: string) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['budgetadvisor', missionId],
    queryFn: () => getBudgetAdvice(missionId),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!missionId,
  })

  return { data, isLoading, error }
}

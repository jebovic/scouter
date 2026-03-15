import { useQuery } from '@tanstack/react-query'
import { getBudgetPlan } from '../api/budgetplanner'

export function useBudgetPlan(missionId: string) {
  return useQuery({
    queryKey: ['budget-plan', missionId],
    queryFn: () => getBudgetPlan(missionId),
    staleTime: 10 * 60 * 1000, // 10 minutes — matches server-side cache TTL
    enabled: !!missionId,
  })
}

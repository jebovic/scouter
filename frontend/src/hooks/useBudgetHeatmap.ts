import { useQuery } from '@tanstack/react-query'
import { getBudgetHeatmap } from '../api/budgetheatmap'
import type { BudgetHeatmap } from '../api/budgetheatmap'

const STALE_TIME = 10 * 60 * 1000 // 10 minutes, matches backend cache TTL

export function useBudgetHeatmap() {
  const { data, isLoading, error } = useQuery<BudgetHeatmap>({
    queryKey: ['budgetHeatmap'],
    queryFn: getBudgetHeatmap,
    staleTime: STALE_TIME,
  })

  return { data, isLoading, error }
}

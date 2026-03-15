import { useQuery } from '@tanstack/react-query'
import { getExpenseSummary } from '../api/expensecategorizer'

const EXPENSE_SUMMARY_KEY = (missionId: string) => ['expense-summary', missionId]

export function useExpenseCategorizer(missionId: string) {
  return useQuery({
    queryKey: EXPENSE_SUMMARY_KEY(missionId),
    queryFn: () => getExpenseSummary(missionId),
    enabled: missionId.length > 0,
    staleTime: 15 * 60 * 1000, // 15 minutes
  })
}

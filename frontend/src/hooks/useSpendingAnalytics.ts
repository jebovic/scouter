import { useQuery } from '@tanstack/react-query'
import { getSpendingAnalytics } from '../api/spendinganalytics'
import type { SpendingAnalytics } from '../api/spendinganalytics'

const STALE_TIME = 10 * 60 * 1000 // 10 minutes, matches backend cache TTL

export function useSpendingAnalytics() {
  const { data, isLoading, error } = useQuery<SpendingAnalytics>({
    queryKey: ['spendingAnalytics'],
    queryFn: getSpendingAnalytics,
    staleTime: STALE_TIME,
  })

  return { data, isLoading, error }
}

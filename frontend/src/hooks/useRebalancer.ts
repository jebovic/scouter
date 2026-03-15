import { useQuery } from '@tanstack/react-query'
import { getRebalancePlan } from '../api/rebalancer'
import type { RebalancePlan } from '../api/rebalancer'

const REBALANCER_QUERY_KEY = ['budget', 'rebalance'] as const

const SIX_HOURS_MS = 6 * 60 * 60 * 1000

export function useBudgetRebalancer() {
  const query = useQuery<RebalancePlan, Error>({
    queryKey: REBALANCER_QUERY_KEY,
    queryFn: getRebalancePlan,
    staleTime: SIX_HOURS_MS,
    retry: 1,
  })

  return {
    plan: query.data ?? null,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,
  }
}

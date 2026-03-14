import { useQuery } from '@tanstack/react-query'
import { getUsageSummary, type UsagePeriod } from '../api/usage'

export function useUsageSummary(period: UsagePeriod = '7d') {
  const { data, isLoading, error } = useQuery({
    queryKey: ['usage', period],
    queryFn: () => getUsageSummary(period),
    staleTime: 60_000, // data considered stale after 1 minute
  })
  return { summary: data, isLoading, error }
}

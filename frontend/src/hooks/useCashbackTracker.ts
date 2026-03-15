import { useQuery } from '@tanstack/react-query'
import { getCashbackSummary } from '../api/cashbacktracker'

export function useCashbackTracker(missionId: string) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['cashback', missionId],
    queryFn: () => getCashbackSummary(missionId),
    staleTime: 60 * 60 * 1000, // 1 hour
    enabled: !!missionId,
  })
  return { data, isLoading, isError }
}

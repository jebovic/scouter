import { useQuery } from '@tanstack/react-query'
import { getLoyaltySummary } from '../api/loyaltytracker'
import type { LoyaltySummary } from '../api/loyaltytracker'

const FIFTEEN_MINUTES = 15 * 60 * 1000

export function useLoyaltySummary(missionId: string) {
  const { data, isLoading, error } = useQuery<LoyaltySummary>({
    queryKey: ['loyaltySummary', missionId],
    queryFn: () => getLoyaltySummary(missionId),
    staleTime: FIFTEEN_MINUTES,
    enabled: Boolean(missionId),
  })

  return { summary: data, isLoading, error }
}

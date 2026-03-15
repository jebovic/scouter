import { useQuery } from '@tanstack/react-query'
import { getPriceAlertDigest } from '../api/pricealertdigest'
import type { PriceAlertDigest } from '../api/pricealertdigest'

const FIVE_MINUTES = 5 * 60 * 1000

/**
 * Fetches the price alert digest for a mission.
 * Includes target_hit, below_avg, and trending_down alerts.
 * 5-minute stale time, refetch every 5 min.
 */
export function usePriceAlertDigest(missionId: string | undefined) {
  const { data, isLoading, error } = useQuery<PriceAlertDigest>({
    queryKey: ['priceAlertDigest', missionId],
    queryFn: () => getPriceAlertDigest(missionId!),
    enabled: !!missionId,
    staleTime: FIVE_MINUTES,
    refetchInterval: FIVE_MINUTES,
  })

  return { digest: data, isLoading, error }
}

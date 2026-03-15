import { useQuery } from '@tanstack/react-query'
import { getItemForecast } from '../api/priceforecast'
import type { PriceForecastResult } from '../api/priceforecast'

/**
 * usePriceForecast fetches the AI-generated price forecast for a shopping item.
 *
 * @param missionId - UUID of the parent mission.
 * @param itemId    - UUID of the shopping item (null/undefined to disable).
 *
 * Stale time is 2h to match the backend cache TTL.
 */
export function usePriceForecast(missionId: string | null | undefined, itemId: string | null | undefined) {
  return useQuery<PriceForecastResult>({
    queryKey: ['price-forecast', missionId, itemId],
    queryFn: () => getItemForecast(missionId!, itemId!),
    enabled: !!missionId && !!itemId,
    staleTime: 2 * 60 * 60 * 1000, // 2h — matches backend cache TTL
    retry: false,
  })
}

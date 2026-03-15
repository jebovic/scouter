import { useQuery } from '@tanstack/react-query'
import { fetchPricePrediction } from '../api/prediction'

/**
 * Fetches a 30-day price prediction for the given shopping item.
 * Disabled when itemId is falsy. Results are considered fresh for 1 hour.
 */
export function usePricePrediction(itemId: string | undefined) {
  return useQuery({
    queryKey: ['prediction', itemId],
    queryFn: () => fetchPricePrediction(itemId!),
    enabled: Boolean(itemId),
    staleTime: 60 * 60_000, // 1 hour
  })
}

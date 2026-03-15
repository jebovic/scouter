import { useQuery } from '@tanstack/react-query'
import { getPriceComparison } from '../api/pricecomparison'

export function usePriceComparison(missionId: string, itemId: string) {
  return useQuery({
    queryKey: ['price-comparison', missionId, itemId],
    queryFn: () => getPriceComparison(missionId, itemId),
    staleTime: 60 * 60 * 1000, // 1h
    enabled: Boolean(missionId && itemId),
  })
}

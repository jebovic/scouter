import { useQuery } from '@tanstack/react-query'
import { getPriceFloor } from '../api/pricefloor'
import type { PriceFloor } from '../api/pricefloor'

export function usePriceFloor(missionId: string, itemId: string): { floor: PriceFloor | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['price-floor', missionId, itemId],
    queryFn: () => getPriceFloor(missionId, itemId),
    staleTime: 2 * 60 * 60 * 1000, // 2 hours
    enabled: Boolean(missionId && itemId),
  })

  return { floor: data ?? null, isLoading }
}

import { useQuery } from '@tanstack/react-query'
import { getPriceStreak } from '../api/pricestreak'
import type { PriceStreak } from '../api/pricestreak'

export function usePriceStreak(missionId: string, itemId: string): { streak: PriceStreak | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['price-streak', missionId, itemId],
    queryFn: () => getPriceStreak(missionId, itemId),
    staleTime: 60 * 60 * 1000, // 1 hour
    enabled: Boolean(missionId && itemId),
  })

  return { streak: data ?? null, isLoading }
}

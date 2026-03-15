import { useQuery } from '@tanstack/react-query'
import { getPriceAnnotations } from '../api/priceannotation'
import type { PriceAnnotations } from '../api/priceannotation'

export function usePriceAnnotations(
  missionId: string,
  itemId: string,
): { annotations: PriceAnnotations | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['price-annotations', missionId, itemId],
    queryFn: () => getPriceAnnotations(missionId, itemId),
    staleTime: 2 * 60 * 60 * 1000, // 2 hours — matches backend cache TTL
    enabled: Boolean(missionId && itemId),
  })

  return { annotations: data ?? null, isLoading }
}

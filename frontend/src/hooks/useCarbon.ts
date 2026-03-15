import { useQuery } from '@tanstack/react-query'
import { getCarbonEstimate } from '../api/carbon'

export function useCarbon(itemName: string, category = '') {
  return useQuery({
    queryKey: ['carbon', itemName, category],
    queryFn: () => getCarbonEstimate(itemName, category),
    enabled: itemName.length > 2,
    staleTime: 24 * 60 * 60 * 1000, // 24h — estimates don't change
    retry: false,
  })
}

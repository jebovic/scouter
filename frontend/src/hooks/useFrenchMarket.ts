import { useQuery } from '@tanstack/react-query'
import { getMarketInsight } from '../api/frenchmarket'

export function useFrenchMarket(itemName: string, category = '') {
  return useQuery({
    queryKey: ['frenchmarket', itemName, category],
    queryFn: () => getMarketInsight(itemName, category),
    enabled: itemName.length > 2,
    staleTime: 60 * 60 * 1000, // 1 hour
    retry: false,
  })
}

import { useQuery } from '@tanstack/react-query'
import { getSeasonalPrediction } from '../api/seasonal'

export function useSeasonal(itemName: string, category = '') {
  return useQuery({
    queryKey: ['seasonal', itemName, category],
    queryFn: () => getSeasonalPrediction(itemName, category),
    enabled: itemName.length > 2,
    staleTime: 24 * 60 * 60 * 1000,
    retry: false,
  })
}

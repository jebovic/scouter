import { useQuery } from '@tanstack/react-query'
import { getMarketplaceResults } from '../api/marketplace'

export function useMarketplace(query: string) {
  return useQuery({
    queryKey: ['marketplace', query],
    queryFn: () => getMarketplaceResults(query),
    enabled: query.length > 2,
    staleTime: 5 * 60 * 1000, // 5min
    retry: false,
  })
}

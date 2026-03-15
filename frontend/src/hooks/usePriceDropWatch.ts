import { useQuery } from '@tanstack/react-query'
import { getPriceDropWatchlist } from '../api/pricedropwatch'

export function usePriceDropWatch(missionId: string) {
  return useQuery({
    queryKey: ['pricedropwatch', missionId],
    queryFn: () => getPriceDropWatchlist(missionId),
    staleTime: 20 * 60 * 1000, // 20 minutes
  })
}

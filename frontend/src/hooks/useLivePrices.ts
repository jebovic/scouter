import { useQuery } from '@tanstack/react-query'
import { fetchLivePrices } from '../api/liveprices'

export function useLivePrices(missionId: string, optionId: string, enabled: boolean) {
  const query = useQuery({
    queryKey: ['live-prices', missionId, optionId],
    queryFn: () => fetchLivePrices(missionId, optionId),
    enabled: !!missionId && !!optionId && enabled,
    staleTime: 60_000,
    gcTime: 5 * 60_000,
    retry: 1,
    refetchOnWindowFocus: false,
  })

  return {
    listings: query.data?.listings ?? [],
    providers: query.data?.providers ?? [],
    isCached: query.data?.cached ?? false,
    isLoading: query.isLoading,
    isError: query.isError,
  }
}

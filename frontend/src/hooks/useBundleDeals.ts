import { useQuery } from '@tanstack/react-query'
import { getBundleDeals } from '../api/bundledetector'

export function useBundleDeals(missionId: string) {
  const { data, isLoading } = useQuery({
    queryKey: ['bundle-deals', missionId],
    queryFn: () => getBundleDeals(missionId),
    staleTime: 1000 * 60 * 60,
    enabled: !!missionId,
  })
  return { bundleResponse: data, isLoading }
}

import { useQuery } from '@tanstack/react-query'
import { getBurnRate } from '../api/burnrate'

export function useBurnRate(missionId: string) {
  return useQuery({
    queryKey: ['burnRate', missionId],
    queryFn: () => getBurnRate(missionId),
    staleTime: 5 * 60 * 1000,
    enabled: !!missionId,
  })
}

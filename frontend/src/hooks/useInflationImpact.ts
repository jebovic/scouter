import { useQuery } from '@tanstack/react-query'
import { fetchInflationImpact } from '../api/inflationtracker'
import type { InflationImpact } from '../api/inflationtracker'

/** Fetches inflation impact for a mission. 30-minute staleTime matches backend cache TTL. */
export function useInflationImpact(missionId: string | undefined) {
  return useQuery<InflationImpact>({
    queryKey: ['inflationImpact', missionId],
    queryFn: () => fetchInflationImpact(missionId!),
    enabled: Boolean(missionId),
    staleTime: 30 * 60 * 1000, // 30min — matches backend cache TTL
  })
}

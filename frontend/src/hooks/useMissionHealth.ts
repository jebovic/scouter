import { useQuery } from '@tanstack/react-query'
import { fetchHealthReport } from '../api/health'
import type { HealthReport } from '../api/health'

/** Fetches the mission health score report. 30-minute stale time matches the backend cache TTL. */
export function useMissionHealth(missionSlug: string | undefined) {
  return useQuery<HealthReport>({
    queryKey: ['missionHealth', missionSlug],
    queryFn: () => fetchHealthReport(missionSlug!),
    enabled: Boolean(missionSlug),
    staleTime: 30 * 60 * 1000, // 30min — matches backend cache TTL
  })
}

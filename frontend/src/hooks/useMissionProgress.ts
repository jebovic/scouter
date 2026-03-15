import { useQuery } from '@tanstack/react-query'
import { fetchMissionProgress } from '../api/missionprogress'
import type { MissionProgress } from '../api/missionprogress'

/** Fetches mission progress metrics. 5-minute staleTime matches backend cache TTL. */
export function useMissionProgress(missionId: string | undefined) {
  return useQuery<MissionProgress>({
    queryKey: ['missionProgress', missionId],
    queryFn: () => fetchMissionProgress(missionId!),
    enabled: Boolean(missionId),
    staleTime: 5 * 60 * 1000, // 5min — matches backend cache TTL
  })
}

import { useQuery } from '@tanstack/react-query'
import { fetchMissionROI } from '../api/roicalculator'
import type { MissionROI } from '../api/roicalculator'

/** Fetches mission ROI metrics. 30-minute staleTime matches backend cache TTL. */
export function useMissionROI(missionId: string | undefined) {
  return useQuery<MissionROI>({
    queryKey: ['missionROI', missionId],
    queryFn: () => fetchMissionROI(missionId!),
    enabled: Boolean(missionId),
    staleTime: 30 * 60 * 1000, // 30min — matches backend cache TTL
  })
}

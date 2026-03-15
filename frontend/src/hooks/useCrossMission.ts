import { useQuery } from '@tanstack/react-query'
import { getCrossMission } from '../api/crossmission'
import type { CrossMissionResponse } from '../api/crossmission'

const STALE_TIME = 15 * 60 * 1000 // 15 minutes, matches backend cache TTL

export function useCrossMission() {
  const { data, isLoading, error } = useQuery<CrossMissionResponse>({
    queryKey: ['crossMission'],
    queryFn: getCrossMission,
    staleTime: STALE_TIME,
  })

  return { data, isLoading, error }
}

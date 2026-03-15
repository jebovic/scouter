import { useQuery } from '@tanstack/react-query'
import { getDuplicates } from '../api/duplicatedetector'
import type { DuplicateReport } from '../api/duplicatedetector'

const STALE_TIME = 30 * 60_000 // 30 minutes

export function useDuplicates(missionId: string) {
  const { data, isLoading, error } = useQuery<DuplicateReport>({
    queryKey: ['duplicates', missionId],
    queryFn: () => getDuplicates(missionId),
    enabled: Boolean(missionId),
    staleTime: STALE_TIME,
  })
  return { report: data ?? null, isLoading, error }
}

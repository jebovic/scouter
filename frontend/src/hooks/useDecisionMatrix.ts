import { useQuery } from '@tanstack/react-query'
import { fetchDecisionMatrix } from '../api/decisionmatrix'
import type { MatrixResponse } from '../api/decisionmatrix'

/** Fetches the purchase decision matrix for a mission. 15-minute staleTime matches backend cache TTL. */
export function useDecisionMatrix(missionId: string | undefined) {
  return useQuery<MatrixResponse>({
    queryKey: ['decisionMatrix', missionId],
    queryFn: () => fetchDecisionMatrix(missionId!),
    enabled: Boolean(missionId),
    staleTime: 15 * 60 * 1000, // 15min — matches backend cache TTL
  })
}

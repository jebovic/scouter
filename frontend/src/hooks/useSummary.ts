import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { generateSummary, fetchSummary } from '../api/summary'
import type { ShoppingSummaryDTO } from '../api/summary'

const SUMMARY_KEY = (missionId: string) => ['summary', missionId] as const

export function useSummary(missionId: string) {
  const { data: summary, isLoading } = useQuery<ShoppingSummaryDTO | null>({
    queryKey: SUMMARY_KEY(missionId),
    queryFn: () => fetchSummary(missionId),
    staleTime: 24 * 60 * 60 * 1000, // 24 h
    enabled: !!missionId,
  })

  return { summary: summary ?? null, isLoading }
}

export function useGenerateSummary(missionId: string) {
  const qc = useQueryClient()

  const { mutate: generate, isPending, error } = useMutation<ShoppingSummaryDTO>({
    mutationFn: () => generateSummary(missionId),
    onSuccess: (data) => {
      qc.setQueryData(SUMMARY_KEY(missionId), data)
    },
  })

  return { generate, isPending, error }
}

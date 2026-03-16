import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { generateSummary, fetchSummary, getMissionSummary } from '../api/summary'
import type { ShoppingSummaryDTO, MissionSummaryDTO } from '../api/summary'

// ── Legacy Shopping Summary hooks (Phase 43) ──────────────────────────────────

const SUMMARY_KEY = (missionSlug: string) => ['summary', missionSlug] as const

export function useSummary(missionSlug: string) {
  const { data: summary, isLoading } = useQuery<ShoppingSummaryDTO | null>({
    queryKey: SUMMARY_KEY(missionSlug),
    queryFn: () => fetchSummary(missionSlug),
    staleTime: 24 * 60 * 60 * 1000, // 24 h
    enabled: !!missionSlug,
  })

  return { summary: summary ?? null, isLoading }
}

export function useGenerateSummary(missionSlug: string) {
  const qc = useQueryClient()

  const { mutate: generate, isPending, error } = useMutation<ShoppingSummaryDTO>({
    mutationFn: () => generateSummary(missionSlug),
    onSuccess: (data) => {
      qc.setQueryData(SUMMARY_KEY(missionSlug), data)
    },
  })

  return { generate, isPending, error }
}

// ── Mission Summary Card hook (Phase 72) ─────────────────────────────────────

const MISSION_BRIEF_KEY = (slug: string) => ['mission-brief', slug] as const

/**
 * useMissionSummary fetches the AI executive brief for a mission.
 *
 * @param slug   Mission slug used to call GET /api/missions/{slug}/summary
 * @param enabled Set to true to trigger the fetch (lazy by default).
 *               The backend generates and caches on first GET, so enabling
 *               this will call the LLM if not already cached server-side.
 */
export function useMissionSummary(slug: string, enabled: boolean = false) {
  const qc = useQueryClient()

  const { data, isLoading, isError, error, refetch } = useQuery<MissionSummaryDTO>({
    queryKey: MISSION_BRIEF_KEY(slug),
    queryFn: () => getMissionSummary(slug),
    staleTime: 60 * 60 * 1000, // 1h — mirrors server-side cache TTL
    enabled: !!slug && enabled,
    retry: 1,
  })

  function invalidate() {
    qc.invalidateQueries({ queryKey: MISSION_BRIEF_KEY(slug) })
  }

  return {
    brief: data ?? null,
    isLoading,
    isError,
    error,
    refetch,
    invalidate,
  }
}

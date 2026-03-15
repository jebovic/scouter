import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchCoachTips } from '../api/coach'
import { useToast } from '../components/scouter'

export function useCoachTips(missionSlug: string | undefined) {
  return useQuery({
    queryKey: ['coach', missionSlug],
    queryFn: () => (missionSlug ? fetchCoachTips(missionSlug) : Promise.reject('No mission slug')),
    enabled: !!missionSlug,
    staleTime: 15 * 60 * 1000, // 15 minutes
  })
}

export function useRefreshCoach(missionSlug: string | undefined) {
  const qc = useQueryClient()
  const { toast } = useToast()

  return useMutation({
    mutationFn: async () => {
      if (!missionSlug) throw new Error('No mission slug')
      return fetchCoachTips(missionSlug)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['coach', missionSlug] })
      toast('Coach tips refreshed', 'success')
    },
    onError: (err: unknown) => {
      toast(`Failed to refresh coach tips: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error')
    },
  })
}

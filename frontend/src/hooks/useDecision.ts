import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { runDecision, getDecision } from '../api/decision'
import { useToast } from '../components/scouter'

export function useDecision(missionId: string) {
  const { data: decision, isLoading, error } = useQuery({
    queryKey: ['decision', missionId],
    queryFn: () => getDecision(missionId),
    retry: false,
    enabled: !!missionId,
  })
  return { decision, isLoading, error }
}

export function useRunDecision(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending, error } = useMutation({
    mutationFn: () => runDecision(missionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['decision', missionId] })
      toast('Decision analysis complete', 'success')
    },
    onError: (err: unknown) => toast(`Decision analysis failed: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { runDecision: mutateAsync, isPending, error }
}

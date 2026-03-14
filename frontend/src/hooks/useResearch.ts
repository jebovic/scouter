import { useMutation, useQueryClient } from '@tanstack/react-query'
import { triggerResearch } from '../api'
import { useToast } from '../components/scouter'

export function useResearch(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending, error, isSuccess } = useMutation({
    mutationFn: (feedback?: string) => triggerResearch(missionId, feedback),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['options', missionId] })
      qc.invalidateQueries({ queryKey: ['agent-runs', missionId] })
      toast('Research complete', 'success')
    },
    onError: (err: unknown) => toast(`Research failed: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { triggerResearch: mutateAsync, isPending, error, isSuccess }
}

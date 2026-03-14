import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listOptions, updateOption, deleteOption } from '../api'
import { useToast } from '../components/scouter'
import type { OptionUpdateRequest } from '../types'

export function useOptions(missionId: string) {
  const {
    data: options = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['options', missionId],
    queryFn: () => listOptions(missionId),
    enabled: Boolean(missionId),
  })
  return { options, isLoading, error }
}

export function useUpdateOption(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: ({ optionId, req }: { optionId: string; req: OptionUpdateRequest }) =>
      updateOption(missionId, optionId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['options', missionId] })
      toast('Option updated', 'success')
    },
    onError: (err: unknown) => toast(`Failed to update option: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { updateOption: mutateAsync, isPending }
}

export function useDeleteOption(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (optionId: string) => deleteOption(missionId, optionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['options', missionId] })
      toast('Option removed', 'success')
    },
    onError: (err: unknown) => toast(`Failed to remove option: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { deleteOption: mutateAsync, isPending }
}

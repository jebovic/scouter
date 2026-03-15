import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listOptions, updateOption, deleteOption, pinOption, rejectOption, unrejectOption, deletePinnedOptions } from '../api'
import { useToast } from '../components/scouter'
import { queryKeys } from '../lib/queryKeys'
import type { OptionUpdateRequest } from '../types'

export function useOptions(missionId: string) {
  const {
    data: options = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: queryKeys.options.all(missionId),
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

export function usePinOption(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (optionId: string) => pinOption(missionId, optionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['options', missionId] })
      toast('Option pinned', 'success')
    },
    onError: (err: unknown) => toast(`Failed to pin option: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { pinOption: mutateAsync, isPending }
}

export function useRejectOption(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: ({ optionId, reason }: { optionId: string; reason: string }) =>
      rejectOption(missionId, optionId, reason),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['options', missionId] })
      toast('Option rejected', 'success')
    },
    onError: (err: unknown) => toast(`Failed to reject option: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { rejectOption: mutateAsync, isPending }
}

export function useUnrejectOption(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (optionId: string) => unrejectOption(missionId, optionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['options', missionId] })
      toast('Option un-rejected', 'success')
    },
    onError: (err: unknown) => toast(`Failed to un-reject option: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { unrejectOption: mutateAsync, isPending }
}

export function useDeletePinnedOptions(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: () => deletePinnedOptions(missionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['options', missionId] })
      toast('Pinned options cleared', 'success')
    },
    onError: (err: unknown) => toast(`Failed to clear pinned options: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { deletePinnedOptions: mutateAsync, isPending }
}

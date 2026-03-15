import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getVotes, recordVote } from '../api/vote'
import type { AggregatedVotes } from '../api/vote'

export function useVotes(optionId: string) {
  return useQuery({
    queryKey: ['votes', optionId],
    queryFn: () => getVotes(optionId),
    enabled: !!optionId,
    staleTime: 10_000,
  })
}

export function useRecordVote(optionId: string) {
  const qc = useQueryClient()
  return useMutation<void, Error, { collaboratorId: string; vote: number }>({
    mutationFn: ({ collaboratorId, vote }) => recordVote(optionId, collaboratorId, vote),
    onMutate: async ({ vote: _vote }) => {
      await qc.cancelQueries({ queryKey: ['votes', optionId] })
      const prev = qc.getQueryData<AggregatedVotes>(['votes', optionId])
      return { prev }
    },
    onError: (_err, _vars, context) => {
      const ctx = context as { prev?: AggregatedVotes } | undefined
      if (ctx?.prev) {
        qc.setQueryData(['votes', optionId], ctx.prev)
      }
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ['votes', optionId] }),
  })
}

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { getTimingAdvice } from '../api/timing'
import type { TimingAdvice } from '../api/timing'

export function useTimingAdvice(missionId: string) {
  const qc = useQueryClient()
  const { mutate: fetchAdvice, isPending, data: advice, error } = useMutation<TimingAdvice>({
    mutationFn: () => getTimingAdvice(missionId),
    onSuccess: (data) => {
      qc.setQueryData(['timing', missionId], data)
    },
  })

  const cached = qc.getQueryData<TimingAdvice>(['timing', missionId])

  return { fetchAdvice, isPending, advice: cached ?? advice, error }
}

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { triggerPricing } from '../api'
import { useToast } from '../components/scouter'

export function usePriceIntel(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending, error, isSuccess } = useMutation({
    mutationFn: () => triggerPricing(missionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['shopping', missionId] })
      toast('Price intel updated', 'success')
    },
    onError: (err: unknown) => toast(`Price intel failed: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { triggerPricing: mutateAsync, isPending, error, isSuccess }
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchPersona, runPersona, getShoppingPersona } from '../api/persona'
import { useToast } from '../components/scouter'

export function usePersona() {
  const { data: persona, isLoading, error } = useQuery({
    queryKey: ['persona'],
    queryFn: fetchPersona,
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
  return { persona, isLoading, error }
}

export function useShoppingPersona() {
  const { data: shoppingPersona, isLoading, error } = useQuery({
    queryKey: ['shoppingPersona'],
    queryFn: getShoppingPersona,
    staleTime: 4 * 60 * 60 * 1000, // 4h TTL matches backend cache
    retry: false,
  })
  return { shoppingPersona, isLoading, error }
}

export function useRunPersona() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending, error } = useMutation({
    mutationFn: runPersona,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['persona'] })
      toast('Persona analysis complete', 'success')
    },
    onError: (err: unknown) =>
      toast(
        `Analysis failed: ${err instanceof Error ? err.message : 'Unknown error'}`,
        'error',
      ),
  })
  return { runPersona: mutateAsync, isPending, error }
}

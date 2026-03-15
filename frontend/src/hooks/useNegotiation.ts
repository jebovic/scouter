import { useMutation } from '@tanstack/react-query'
import { fetchNegotiationScript } from '../api/negotiation'

export function useNegotiation(optionId: string) {
  const { mutate, data, isPending, error, reset } = useMutation({
    mutationFn: () => fetchNegotiationScript(optionId),
  })
  return { coach: mutate, script: data, isPending, error, reset }
}

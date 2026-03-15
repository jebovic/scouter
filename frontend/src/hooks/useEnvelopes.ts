import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listEnvelopes,
  createEnvelope,
  updateEnvelope,
  deleteEnvelope,
  getEnvelopeSummary,
  type EnvelopeDTO,
} from '../api/envelope'

const ENVELOPES_KEY = ['envelopes']

export function useEnvelopes() {
  const { data, isLoading, error } = useQuery({
    queryKey: ENVELOPES_KEY,
    queryFn: listEnvelopes,
    staleTime: 30_000,
  })
  return {
    envelopes: data ?? [],
    isLoading,
    error,
  }
}

export function useCreateEnvelope() {
  const qc = useQueryClient()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: createEnvelope,
    onSuccess: () => qc.invalidateQueries({ queryKey: ENVELOPES_KEY }),
  })
  return { create: mutateAsync, isPending }
}

export function useUpdateEnvelope() {
  const qc = useQueryClient()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Partial<EnvelopeDTO> }) =>
      updateEnvelope(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ENVELOPES_KEY }),
  })
  return { update: mutateAsync, isPending }
}

export function useDeleteEnvelope() {
  const qc = useQueryClient()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: deleteEnvelope,
    onSuccess: () => qc.invalidateQueries({ queryKey: ENVELOPES_KEY }),
  })
  return { remove: mutateAsync, isPending }
}

export function useEnvelopeSummary(envelopeId: string | null) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['envelope-summary', envelopeId],
    queryFn: () => getEnvelopeSummary(envelopeId!),
    staleTime: 60_000,
    enabled: !!envelopeId,
  })
  return { summary: data ?? null, isLoading, error }
}

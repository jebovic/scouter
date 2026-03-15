import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createCashbackEntry,
  deleteCashbackEntry,
  getCashbackSummary,
  listCashbackEntries,
  updateCashbackEntry,
} from '../api/cashback'
import type { CreateCashbackPayload, UpdateCashbackPayload } from '../api/cashback'

const ENTRIES_KEY = (userRef: string) => ['cashback-entries', userRef]
const SUMMARY_KEY = (userRef: string) => ['cashback-summary', userRef]

export function useCashbackEntries(userRef: string) {
  return useQuery({
    queryKey: ENTRIES_KEY(userRef),
    queryFn: () => listCashbackEntries(userRef),
    enabled: userRef.length > 0,
    staleTime: 60_000,
  })
}

export function useCashbackSummary(userRef: string) {
  return useQuery({
    queryKey: SUMMARY_KEY(userRef),
    queryFn: () => getCashbackSummary(userRef),
    enabled: userRef.length > 0,
    staleTime: 60_000,
  })
}

export function useCreateCashback(userRef: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateCashbackPayload) => createCashbackEntry(userRef, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ENTRIES_KEY(userRef) })
      qc.invalidateQueries({ queryKey: SUMMARY_KEY(userRef) })
    },
  })
}

export function useUpdateCashback(userRef: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateCashbackPayload }) =>
      updateCashbackEntry(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ENTRIES_KEY(userRef) })
      qc.invalidateQueries({ queryKey: SUMMARY_KEY(userRef) })
    },
  })
}

export function useDeleteCashback(userRef: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteCashbackEntry(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ENTRIES_KEY(userRef) })
      qc.invalidateQueries({ queryKey: SUMMARY_KEY(userRef) })
    },
  })
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getPurchaseRecord,
  createPurchaseRecord,
  updatePurchaseRecord,
  getStats,
  fetchMonthlyStats,
  type CreatePurchaseRequest,
  type UpdatePurchaseRequest,
} from '../api/purchase'

export function usePurchaseRecord(missionId: string | undefined) {
  return useQuery({
    queryKey: ['purchase', missionId],
    queryFn: () => getPurchaseRecord(missionId!),
    enabled: !!missionId,
  })
}

export function useRecordPurchase(missionId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: CreatePurchaseRequest) => createPurchaseRecord(missionId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['purchase', missionId] })
      qc.invalidateQueries({ queryKey: ['missions'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

export function useUpdatePurchase(missionId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: UpdatePurchaseRequest) => updatePurchaseRecord(missionId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['purchase', missionId] })
    },
  })
}

export function useStats() {
  return useQuery({
    queryKey: ['stats'],
    queryFn: getStats,
    staleTime: 5 * 60 * 1000,
  })
}

export function useMonthlyStats() {
  return useQuery({
    queryKey: ['stats', 'monthly'],
    queryFn: fetchMonthlyStats,
    staleTime: 5 * 60 * 1000,
  })
}

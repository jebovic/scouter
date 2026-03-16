import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../lib/queryKeys'
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
    queryKey: queryKeys.purchase(missionId ?? ''),
    queryFn: () => getPurchaseRecord(missionId!),
    enabled: !!missionId,
    retry: false,
  })
}

export function useRecordPurchase(missionId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: CreatePurchaseRequest) => createPurchaseRecord(missionId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.purchase(missionId) })
      qc.invalidateQueries({ queryKey: queryKeys.missions.all() })
      qc.invalidateQueries({ queryKey: queryKeys.stats() })
    },
  })
}

export function useUpdatePurchase(missionId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: UpdatePurchaseRequest) => updatePurchaseRecord(missionId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.purchase(missionId) })
    },
  })
}

export function useStats() {
  return useQuery({
    queryKey: queryKeys.stats(),
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

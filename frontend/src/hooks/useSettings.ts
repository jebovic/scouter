import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../lib/queryKeys'
import {
  getSettings,
  updateSettings as updateSettingsApi,
  deleteAllData,
} from '../api/settings'

export function useSettings() {
  return useQuery({
    queryKey: queryKeys.settings(),
    queryFn: getSettings,
    staleTime: 10 * 60 * 1000,
  })
}

export function useUpdateSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (patch: Record<string, unknown>) => updateSettingsApi(patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })
}

export function useDeleteAllData() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: deleteAllData,
    onSuccess: () => {
      qc.clear()
    },
  })
}

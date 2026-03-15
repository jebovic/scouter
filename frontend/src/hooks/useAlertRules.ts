import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getAlertRules, addAlertRule, deleteAlertRule } from '../api/pricealertrule'

export function useAlertRules(itemId: string) {
  const qc = useQueryClient()
  const key = ['alertRules', itemId]

  const { data: rules = [], isLoading } = useQuery({
    queryKey: key,
    queryFn: () => getAlertRules(itemId),
    enabled: !!itemId,
    staleTime: 60 * 1000,
  })

  const add = useMutation({
    mutationFn: (rule: { type: string; value: number; basePrice: number }) =>
      addAlertRule(itemId, rule),
    onSuccess: () => qc.invalidateQueries({ queryKey: key }),
  })

  const remove = useMutation({
    mutationFn: (ruleId: string) => deleteAlertRule(itemId, ruleId),
    onSuccess: () => qc.invalidateQueries({ queryKey: key }),
  })

  return { rules, isLoading, add, remove }
}

import { useQuery } from '@tanstack/react-query'
import { getBudgetAlerts } from '../api/budgetalert'

const FIVE_MINUTES = 5 * 60 * 1000

export function useBudgetAlertBanners() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['budget-alert-banners'],
    queryFn: getBudgetAlerts,
    staleTime: FIVE_MINUTES,
    refetchInterval: FIVE_MINUTES,
  })

  return {
    alerts: data?.alerts ?? [],
    missionsChecked: data?.missionsChecked ?? 0,
    isLoading,
    error,
  }
}

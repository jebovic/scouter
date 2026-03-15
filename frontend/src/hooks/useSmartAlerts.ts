import { useQuery } from '@tanstack/react-query'
import { fetchSmartAlerts } from '../api/notificationrules'
import type { AlertsResponse } from '../api/notificationrules'

/** Fetches smart alert rules evaluation for a mission. Refetches every 5 minutes to match backend cache TTL. */
export function useSmartAlerts(missionId: string | undefined) {
  return useQuery<AlertsResponse>({
    queryKey: ['smartAlerts', missionId],
    queryFn: () => fetchSmartAlerts(missionId!),
    enabled: Boolean(missionId),
    staleTime: 5 * 60 * 1000,
    refetchInterval: 5 * 60 * 1000,
  })
}

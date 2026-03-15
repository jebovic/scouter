import { useQuery } from '@tanstack/react-query'
import { getSortedItems } from '../api/listsorter'
import type { RankedItem } from '../api/listsorter'

export function useSortedItems(missionId: string, enabled: boolean) {
  const {
    data,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['sortedItems', missionId],
    queryFn: () => getSortedItems(missionId),
    enabled: Boolean(missionId) && enabled,
    staleTime: 15 * 60_000, // 15 min — matches backend cache TTL
  })

  const rankMap = new Map<string, RankedItem>()
  if (data) {
    for (const item of data.items) {
      rankMap.set(item.id, item)
    }
  }

  return { sortedItems: data?.items ?? [], rankMap, isLoading, error }
}

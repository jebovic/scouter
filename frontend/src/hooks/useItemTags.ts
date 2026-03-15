import { useQuery } from '@tanstack/react-query'
import { getItemTags } from '../api/itemtagger'

export function useItemTags(missionId: string, itemId: string) {
  return useQuery({
    queryKey: ['item-tags', missionId, itemId],
    queryFn: () => getItemTags(missionId, itemId),
    staleTime: 24 * 60 * 60 * 1000, // 24 hours
    enabled: Boolean(missionId) && Boolean(itemId),
  })
}

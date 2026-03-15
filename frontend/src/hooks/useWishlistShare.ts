import { useQuery } from '@tanstack/react-query'
import { getWishlistCard } from '../api/wishlistshare'
import type { WishlistShareCard } from '../api/wishlistshare'

const STALE_TIME = 5 * 60 * 1000 // 5 minutes

export function useWishlistShare(missionId: string) {
  return useQuery<WishlistShareCard>({
    queryKey: ['wishlist-card', missionId],
    queryFn: () => getWishlistCard(missionId),
    enabled: missionId.length > 0,
    staleTime: STALE_TIME,
  })
}

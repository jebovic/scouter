import { useQuery } from '@tanstack/react-query';
import { fetchWishlistPrioritized, type PrioritizedList } from '../api/wishlistPrioritizer';

export function useWishlistPrioritizer() {
  return useQuery<PrioritizedList, Error>({
    queryKey: ['wishlist-prioritized'],
    queryFn: fetchWishlistPrioritized,
    staleTime: 15 * 60 * 1000, // 15 minutes
  });
}

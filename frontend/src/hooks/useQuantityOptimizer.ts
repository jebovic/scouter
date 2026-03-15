import { useQuery } from '@tanstack/react-query';
import { fetchQuantityOptimizer, type OptimizerReport } from '../api/quantityOptimizer';

export function useQuantityOptimizer(
  missionId: string | undefined,
  itemId: string | undefined
) {
  return useQuery<OptimizerReport, Error>({
    queryKey: ['quantity-optimizer', missionId, itemId],
    queryFn: () => fetchQuantityOptimizer(missionId!, itemId!),
    enabled: !!missionId && !!itemId,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
}

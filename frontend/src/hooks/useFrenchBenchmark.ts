import { useQuery } from '@tanstack/react-query';
import { fetchFrenchBenchmark, type BenchmarkResult } from '../api/frenchBenchmark';

export function useFrenchBenchmark(missionId: string | undefined) {
  return useQuery<BenchmarkResult, Error>({
    queryKey: ['french-benchmark', missionId],
    queryFn: () => fetchFrenchBenchmark(missionId!),
    enabled: !!missionId,
    staleTime: 20 * 60 * 1000, // 20 minutes
  });
}

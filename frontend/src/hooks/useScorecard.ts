import { useQuery } from '@tanstack/react-query';
import { fetchScorecard, type Scorecard } from '../api/scorecard';

export function useScorecard(missionId: string | undefined) {
  return useQuery<Scorecard, Error>({
    queryKey: ['scorecard', missionId],
    queryFn: () => fetchScorecard(missionId!),
    enabled: !!missionId,
    staleTime: 30 * 60 * 1000, // 30 minutes
  });
}

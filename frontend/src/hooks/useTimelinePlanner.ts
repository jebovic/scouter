import { useQuery } from '@tanstack/react-query';
import { fetchTimeline, type Timeline } from '../api/timelinePlanner';

export function useTimelinePlanner(missionId: string | undefined) {
  return useQuery<Timeline, Error>({
    queryKey: ['purchase-timeline', missionId],
    queryFn: () => fetchTimeline(missionId!),
    enabled: !!missionId,
    staleTime: 20 * 60 * 1000, // 20 minutes
  });
}

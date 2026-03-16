// frontend/src/hooks/useOptionImages.ts
import { useQuery } from "@tanstack/react-query";
import { listOptionImages, type OptionImage } from "../api/images";

const POLL_INTERVAL_MS = 30_000;
const MAX_POLL_RETRIES = 10;
const STALE_TIME_MS = 50 * 60 * 1000; // 50 minutes — presigned URL TTL is 1h

export function useOptionImages(
  missionId: string,
  optionId: string
) {
  return useQuery<OptionImage[]>({
    queryKey: ["option-images", missionId, optionId],
    queryFn: () => listOptionImages(missionId, optionId),
    staleTime: STALE_TIME_MS,
    refetchOnWindowFocus: true,
    // Poll every 30s while no images found, up to 10 retries
    refetchInterval: (query) => {
      const data = query.state.data;
      const retries = query.state.dataUpdateCount;
      if (!data || (data.length === 0 && retries < MAX_POLL_RETRIES)) {
        return POLL_INTERVAL_MS;
      }
      return false;
    },
  });
}

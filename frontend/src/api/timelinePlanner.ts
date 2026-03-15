import { z } from 'zod';

export const WeekBucketSchema = z.object({
  week: z.number(),
  label: z.string(),
  items: z.array(z.string()),
  action: z.string(),
  promoHint: z.string(),
  itemCount: z.number(),
});

export const TimelineSchema = z.object({
  missionId: z.string(),
  missionName: z.string(),
  buckets: z.array(WeekBucketSchema),
  summary: z.string(),
  totalItems: z.number(),
});

export type WeekBucket = z.infer<typeof WeekBucketSchema>;
export type Timeline = z.infer<typeof TimelineSchema>;

export async function fetchTimeline(missionId: string): Promise<Timeline> {
  const res = await fetch(`/api/missions/${missionId}/purchase-timeline`);
  if (!res.ok) throw new Error(`fetchTimeline: ${res.status}`);
  const data = await res.json();
  return TimelineSchema.parse(data);
}

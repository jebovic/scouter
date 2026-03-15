import { z } from 'zod';

export const AchievementSchema = z.object({
  title: z.string(),
  description: z.string(),
  earned: z.boolean(),
});

export const DimensionsSchema = z.object({
  research: z.number(),
  pricing: z.number(),
  budgeting: z.number(),
  speed: z.number(),
});

export const ScorecardSchema = z.object({
  missionId: z.string(),
  missionName: z.string(),
  grade: z.enum(['A', 'B', 'C', 'D']),
  score: z.number(),
  summary: z.string(),
  achievements: z.array(AchievementSchema),
  dimensions: DimensionsSchema,
});

export type Achievement = z.infer<typeof AchievementSchema>;
export type Dimensions = z.infer<typeof DimensionsSchema>;
export type Scorecard = z.infer<typeof ScorecardSchema>;

export async function fetchScorecard(missionId: string): Promise<Scorecard> {
  const res = await fetch(`/api/missions/${missionId}/scorecard`);
  if (!res.ok) throw new Error(`fetchScorecard: ${res.status}`);
  const data = await res.json();
  return ScorecardSchema.parse(data);
}

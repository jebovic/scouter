import { z } from 'zod'
import { apiFetch } from './client'

const SignalSchema = z.object({
  name: z.string(),
  score: z.number(),
  weight: z.number(),
  label: z.string(),
})

const TimingScoreSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  timingScore: z.number(),
  recommendation: z.string(),
  verdict: z.string(),
  signals: z.array(SignalSchema),
  bestBuyWindow: z.string(),
  urgencyLevel: z.string(),
})

export type TimingScore = z.infer<typeof TimingScoreSchema>
export type Signal = z.infer<typeof SignalSchema>

export async function getTimingScore(missionId: string, itemId: string): Promise<TimingScore> {
  const data = await apiFetch(`/api/missions/${missionId}/items/${itemId}/timing-score`)
  return TimingScoreSchema.parse(data)
}

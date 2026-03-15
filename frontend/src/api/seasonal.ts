import { z } from 'zod'
import { apiFetch } from './client'

export const SeasonalPredictionSchema = z.object({
  itemName: z.string(),
  category: z.string(),
  bestMonths: z.array(z.string()),
  avgDiscountPct: z.number(),
  events: z.array(z.string()),
  nextEventDays: z.number(),
  rationale: z.string(),
  confidence: z.enum(['high', 'medium', 'low']),
})
export type SeasonalPrediction = z.infer<typeof SeasonalPredictionSchema>

export async function getSeasonalPrediction(itemName: string, category: string): Promise<SeasonalPrediction> {
  const params = new URLSearchParams({ itemName, category })
  const raw = await apiFetch<unknown>(`/api/seasonal?${params}`)
  return SeasonalPredictionSchema.parse(raw)
}

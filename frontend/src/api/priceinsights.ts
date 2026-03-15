import { z } from 'zod'
import { apiFetch } from './client'

export const DayStatsSchema = z.object({
  day: z.string(),
  avgPrice: z.number(),
  count: z.number().int(),
})

export const PriceInsightsSchema = z.object({
  itemId: z.string(),
  dataPoints: z.number().int(),
  bestDayOfWeek: z.string(),
  worstDayOfWeek: z.string(),
  bestTimeOfMonth: z.string(),
  priceByDay: z.array(DayStatsSchema),
  buyRecommendation: z.string(),
  confidence: z.enum(['low', 'medium', 'high']),
  generatedAt: z.string(),
})

export type PriceInsights = z.infer<typeof PriceInsightsSchema>
export type DayStats = z.infer<typeof DayStatsSchema>

export async function getPriceInsights(missionId: string, itemId: string): Promise<PriceInsights> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/items/${itemId}/price-insights`)
  return PriceInsightsSchema.parse(data)
}

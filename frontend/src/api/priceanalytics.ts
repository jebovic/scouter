import { z } from 'zod'
import { apiFetch } from './client'

export const PriceStatsSchema = z.object({
  itemId: z.string(),
  min: z.number(),
  max: z.number(),
  average: z.number(),
  volatility: z.number(),
  dataPoints: z.number().int(),
  trendSlope: z.number(),
  priceDrop7d: z.number(),
})

export type PriceStats = z.infer<typeof PriceStatsSchema>

export async function getPriceStats(missionId: string, itemId: string): Promise<PriceStats> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/items/${itemId}/price-stats`)
  return PriceStatsSchema.parse(data)
}

import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const PricePredictionSchema = z.object({
  itemId: z.string(),
  currentPrice: z.number(),
  predictedPrice: z.number(),
  confidenceLevel: z.enum(['high', 'medium', 'low']),
  trend: z.enum(['rising', 'falling', 'stable']),
  trendPct: z.number(),
  dataPoints: z.number(),
  forecastDays: z.number(),
})

export type PricePrediction = z.infer<typeof PricePredictionSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function fetchPricePrediction(itemId: string): Promise<PricePrediction> {
  const data = await apiFetch<unknown>(`/api/shopping-items/${itemId}/prediction`)
  return PricePredictionSchema.parse(data)
}

import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ──────────────────────────────────────────────────────────────

export const DataPointSchema = z.object({
  date: z.string(),
  price: z.number(),
})

export const ForecastPointSchema = z.object({
  date: z.string(),
  low: z.number(),
  mid: z.number(),
  high: z.number(),
})

export const PriceForecastResultSchema = z.object({
  item_id: z.string(),
  item_name: z.string(),
  currency: z.string(),
  historical: z.array(DataPointSchema),
  forecast: z.array(ForecastPointSchema),
  confidence: z.number().min(0).max(1),
  summary: z.string(),
  best_buy_window: z.string(),
  generated_at: z.string(),
})

export type DataPoint = z.infer<typeof DataPointSchema>
export type ForecastPoint = z.infer<typeof ForecastPointSchema>
export type PriceForecastResult = z.infer<typeof PriceForecastResultSchema>

// ── API functions ─────────────────────────────────────────────────────────────

export async function getItemForecast(missionId: string, itemId: string): Promise<PriceForecastResult> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/shopping/${itemId}/forecast`)
  return PriceForecastResultSchema.parse(data)
}

import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const ForecastResultSchema = z.object({
  id: z.string(),
  missionId: z.string(),
  estimatedTotal: z.number().nullable(),
  confidence: z.enum(['low', 'medium', 'high']),
  recommendations: z.array(z.string()),
  riskFactors: z.array(z.string()),
  monthsToSave: z.number().nullable(),
  optimalBuyTime: z.string().nullable(),
  createdAt: z.string(),
})

export type ForecastResult = z.infer<typeof ForecastResultSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function fetchForecast(missionId: string): Promise<ForecastResult> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/forecast`)
  return ForecastResultSchema.parse(data)
}

export async function runForecast(missionId: string): Promise<ForecastResult> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/forecast`, {
    method: 'POST',
  })
  return ForecastResultSchema.parse(data)
}

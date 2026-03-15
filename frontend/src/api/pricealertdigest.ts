import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const AlertEntrySchema = z.object({
  itemId: z.string(),
  name: z.string(),
  price: z.number(),
  alertType: z.enum(['target_hit', 'below_avg', 'trending_down']),
  message: z.string(),
  savings: z.number(),
})

export const PriceAlertDigestSchema = z.object({
  missionId: z.string(),
  alertCount: z.number(),
  alerts: z.array(AlertEntrySchema),
  summary: z.string(),
  generatedAt: z.string(),
})

export type AlertEntry = z.infer<typeof AlertEntrySchema>
export type PriceAlertDigest = z.infer<typeof PriceAlertDigestSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function getPriceAlertDigest(missionId: string): Promise<PriceAlertDigest> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/price-alert-digest`)
  return PriceAlertDigestSchema.parse(data)
}

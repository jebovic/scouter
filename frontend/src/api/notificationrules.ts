import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const AlertSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  ruleId: z.string(),
  severity: z.enum(['critical', 'warning']),
  message: z.string(),
  triggeredAt: z.string(),
})

export type Alert = z.infer<typeof AlertSchema>

export const AlertsResponseSchema = z.object({
  missionId: z.string(),
  alerts: z.array(AlertSchema),
  criticalCount: z.number(),
  warningCount: z.number(),
})

export type AlertsResponse = z.infer<typeof AlertsResponseSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function fetchSmartAlerts(missionId: string): Promise<AlertsResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/smart-alerts`)
  return AlertsResponseSchema.parse(data)
}

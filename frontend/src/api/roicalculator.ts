import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schema ────────────────────────────────────────────────────────────────

export const MissionROISchema = z.object({
  missionId: z.string(),
  totalSpend: z.number(),
  totalEstimate: z.number(),
  totalTargetSavings: z.number(),
  savingsVsEstimate: z.number(),
  savingsPercent: z.number(),
  priceChecks: z.number().int(),
  daysTracking: z.number().int(),
  avgChecksPerDay: z.number(),
  roiScore: z.number(),
  verdict: z.string(),
  generatedAt: z.string(),
})

export type MissionROI = z.infer<typeof MissionROISchema>

// ── API function ──────────────────────────────────────────────────────────────

export async function fetchMissionROI(missionId: string): Promise<MissionROI> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/roi`)
  return MissionROISchema.parse(data)
}

import { z } from 'zod'
import { apiFetch } from './client'

// ── Legacy Shopping Summary (Phase 43) ───────────────────────────────────────

export const ShoppingSummarySchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  content: z.string(),
  generatedAt: z.string(),
})

export type ShoppingSummaryDTO = z.infer<typeof ShoppingSummarySchema>

export async function generateSummary(missionId: string): Promise<ShoppingSummaryDTO> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/summary`, {
    method: 'POST',
  })
  return ShoppingSummarySchema.parse(data)
}

export async function fetchSummary(missionId: string): Promise<ShoppingSummaryDTO | null> {
  try {
    const data = await apiFetch<unknown>(`/api/missions/${missionId}/summary`)
    return ShoppingSummarySchema.parse(data)
  } catch {
    return null
  }
}

// ── Mission Summary Card (Phase 72) ──────────────────────────────────────────

export const VerdictSchema = z.enum(['go', 'wait', 'research_more'])

export type Verdict = z.infer<typeof VerdictSchema>

export const MissionSummarySchema = z.object({
  missionSlug: z.string(),
  bullets: z.array(z.string()).min(1),
  verdict: VerdictSchema,
  cachedAt: z.number().int(), // Unix timestamp (seconds)
})

export type MissionSummaryDTO = z.infer<typeof MissionSummarySchema>

export async function getMissionSummary(slug: string): Promise<MissionSummaryDTO> {
  const data = await apiFetch<unknown>(`/api/missions/${slug}/summary`)
  return MissionSummarySchema.parse(data)
}

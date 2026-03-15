import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const RegretSignalSchema = z.object({
  type: z.enum(['overpaid', 'impulse', 'timing', 'ok']),
  severity: z.enum(['high', 'medium', 'low', 'none']),
  description: z.string(),
  potentialSaving: z.number(),
})

export const RegretAnalyzerResponseSchema = z.object({
  missionId: z.string().uuid(),
  totalSpent: z.number(),
  signals: z.array(RegretSignalSchema),
  regretScore: z.number().int().min(0).max(100),
  summary: z.string(),
})

export type RegretSignal = z.infer<typeof RegretSignalSchema>
export type RegretAnalyzerResponse = z.infer<typeof RegretAnalyzerResponseSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function getRegretAnalysis(missionId: string): Promise<RegretAnalyzerResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/regret-analysis`)
  return RegretAnalyzerResponseSchema.parse(data)
}

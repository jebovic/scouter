import { z } from 'zod'
import { apiFetch } from './client'

export const ModelStatusSchema = z.object({
  name: z.string(),
  healthy: z.boolean(),
  circuit_state: z.enum(['closed', 'open', 'half_open']),
  last_latency_ms: z.number(),
  capabilities: z.array(z.string()),
})

export const PoolHealthSchema = z.object({
  models: z.array(ModelStatusSchema),
  checked_at: z.string().optional(),
})

export type ModelStatus = z.infer<typeof ModelStatusSchema>
export type PoolHealth = z.infer<typeof PoolHealthSchema>

export async function getLLMHealth(): Promise<PoolHealth> {
  const data = await apiFetch<unknown>('/api/health/llm')
  return PoolHealthSchema.parse(data)
}

// ── Mission Health Score (Phase 57) ──────────────────────────────────────────

export const SignalScoreSchema = z.object({
  name: z.string(),
  score: z.number(),
  weight: z.number(),
  description: z.string(),
  icon: z.string(),
})

export const HealthReportSchema = z.object({
  missionId: z.string(),
  overallScore: z.number(),
  grade: z.enum(['A', 'B', 'C', 'D', 'F']),
  signals: z.array(SignalScoreSchema),
  summary: z.string(),
  advice: z.string(),
  generatedAt: z.string(),
})

export type SignalScore = z.infer<typeof SignalScoreSchema>
export type HealthReport = z.infer<typeof HealthReportSchema>

export async function fetchHealthReport(missionSlug: string): Promise<HealthReport> {
  const data = await apiFetch<unknown>(`/api/missions/${missionSlug}/health`)
  return HealthReportSchema.parse(data)
}

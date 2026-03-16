import { z } from 'zod'
import { apiFetch, ApiError } from './client'
import type { Decision } from '../types/decision'

// ── Zod schemas ─────────────────────────────────────────────────────────────

const ScoreResultSchema = z.object({
  optionId: z.string().uuid(),
  optionName: z.string(),
  score: z.number(),
  priceScore: z.number(),
  qualScore: z.number(),
  featScore: z.number(),
  eliminated: z.boolean(),
  eliminReason: z.string().optional(),
})

const DecisionSchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  scores: z.array(ScoreResultSchema),
  summary: z.string(),
  createdAt: z.string(),
})

export type DecisionDTO = z.infer<typeof DecisionSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function runDecision(missionId: string): Promise<Decision> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/decision`, { method: 'POST' })
  return DecisionSchema.parse(data) as Decision
}

export async function getDecision(missionId: string): Promise<Decision | null> {
  try {
    const data = await apiFetch<unknown>(`/api/missions/${missionId}/decision`)
    return DecisionSchema.parse(data) as Decision
  } catch (e: unknown) {
    if (e instanceof ApiError && e.status === 404) {
      return null
    }
    throw e
  }
}

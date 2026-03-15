import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const WeightSchema = z.object({
  id: z.string(),
  missionId: z.string(),
  criteria: z.string(),
  weight: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const OptionScoreSchema = z.object({
  optionId: z.string(),
  optionName: z.string(),
  score: z.number(),
  rank: z.number(),
})

export const MatrixResponseSchema = z.object({
  weights: z.array(WeightSchema),
  scores: z.array(OptionScoreSchema),
})

// ── TypeScript types ─────────────────────────────────────────────────────────

export type Weight = z.infer<typeof WeightSchema>
export type OptionScore = z.infer<typeof OptionScoreSchema>
export type MatrixResponse = z.infer<typeof MatrixResponseSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function fetchWeights(missionId: string): Promise<Weight[]> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/comparison-weights`)
  return z.array(WeightSchema).parse(data)
}

export async function setWeight(missionId: string, criteria: string, weight: number): Promise<Weight> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/comparison-weights`, {
    method: 'PUT',
    body: JSON.stringify({ criteria, weight }),
  })
  return WeightSchema.parse(data)
}

export async function deleteWeight(missionId: string, criteria: string): Promise<void> {
  await apiFetch<void>(`/api/missions/${missionId}/comparison-weights/${encodeURIComponent(criteria)}`, {
    method: 'DELETE',
  })
}

export async function fetchMatrix(missionId: string): Promise<MatrixResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/matrix`)
  return MatrixResponseSchema.parse(data)
}

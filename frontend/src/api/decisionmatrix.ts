import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ───────────────────────────────────────────────────────────────

export const ItemScoresSchema = z.object({
  prix: z.number().int(),
  urgence: z.number().int(),
  valeur: z.number().int(),
  disponibilite: z.number().int(),
  confiance: z.number().int(),
})

export const RankedItemSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  scores: ItemScoresSchema,
  totalScore: z.number().int(),
  recommendation: z.string(),
  rank: z.number().int(),
})

export const MatrixResponseSchema = z.object({
  items: z.array(RankedItemSchema),
  topRecommendation: RankedItemSchema.nullable(),
})

export type ItemScores = z.infer<typeof ItemScoresSchema>
export type RankedItem = z.infer<typeof RankedItemSchema>
export type MatrixResponse = z.infer<typeof MatrixResponseSchema>

// ── API function ──────────────────────────────────────────────────────────────

export async function fetchDecisionMatrix(missionId: string): Promise<MatrixResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/decision-matrix`)
  return MatrixResponseSchema.parse(data)
}

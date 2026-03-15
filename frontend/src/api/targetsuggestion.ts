import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const StrategySchema = z.object({
  type: z.enum(['conservative', 'moderate', 'aggressive']),
  price: z.number(),
  label: z.string(),
  rationale: z.string(),
})

export const TargetSuggestionSchema = z.object({
  itemId: z.string(),
  strategies: z.array(StrategySchema),
})

export type Strategy = z.infer<typeof StrategySchema>
export type TargetSuggestion = z.infer<typeof TargetSuggestionSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function getTargetSuggestion(
  missionId: string,
  itemId: string,
): Promise<TargetSuggestion> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/target-suggestion`,
  )
  return TargetSuggestionSchema.parse(data)
}

export async function applyTargetSuggestion(
  missionId: string,
  itemId: string,
  target: 'conservative' | 'moderate' | 'aggressive',
): Promise<unknown> {
  return apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/target-suggestion/apply`,
    { method: 'POST', body: JSON.stringify({ target }) },
  )
}

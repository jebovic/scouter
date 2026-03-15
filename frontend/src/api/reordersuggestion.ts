import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const SuggestionSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  price: z.number(),
  status: z.string(),
  buyScore: z.number(),
  priority: z.enum(['immédiat', 'cette semaine', 'ce mois', 'flexible']),
  reasoning: z.string(),
  merchant: z.string(),
})

export type Suggestion = z.infer<typeof SuggestionSchema>

export const MerchantBatchSchema = z.object({
  merchant: z.string(),
  items: z.array(z.string()),
  batchNote: z.string(),
})

export type MerchantBatch = z.infer<typeof MerchantBatchSchema>

export const ReorderSuggestionsResponseSchema = z.object({
  missionId: z.string(),
  suggestions: z.array(SuggestionSchema),
  merchantBatches: z.array(MerchantBatchSchema),
  totalToBuy: z.number(),
})

export type ReorderSuggestionsResponse = z.infer<typeof ReorderSuggestionsResponseSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function fetchReorderSuggestions(missionId: string): Promise<ReorderSuggestionsResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/reorder-suggestions`)
  return ReorderSuggestionsResponseSchema.parse(data)
}

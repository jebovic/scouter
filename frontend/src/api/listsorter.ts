import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const RankedItemSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  price: z.number(),
  status: z.string(),
  score: z.number(),
  reason: z.string(),
  rank: z.number(),
})

export const SortedResponseSchema = z.object({
  items: z.array(RankedItemSchema),
  sortedBy: z.string(),
})

export type RankedItem = z.infer<typeof RankedItemSchema>
export type SortedResponse = z.infer<typeof SortedResponseSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function getSortedItems(missionId: string): Promise<SortedResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/sorted-items`)
  return SortedResponseSchema.parse(data)
}

import { z } from 'zod'
import { apiFetch } from './client'

export const PurchaseAdviceSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  decision: z.enum(['buy_now', 'wait', 'skip', 'negotiate']),
  confidence: z.number().int().min(0).max(100),
  reasoning: z.string(),
  alternativeSuggestion: z.string(),
  bestTimeToAct: z.string(),
  estimatedSavingsByWaiting: z.number(),
})

export type PurchaseAdvice = z.infer<typeof PurchaseAdviceSchema>

export async function getAdvice(itemId: string): Promise<PurchaseAdvice> {
  const data = await apiFetch<unknown>(`/api/items/${itemId}/purchase-advice`, {
    method: 'POST',
  })
  return PurchaseAdviceSchema.parse(data)
}

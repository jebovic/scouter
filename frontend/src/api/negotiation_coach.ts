import { z } from 'zod'
import { apiFetch } from './client'

export const NegotiationTipSchema = z.object({
  title: z.string(),
  description: z.string(),
  difficulty: z.enum(['easy', 'medium', 'hard']),
})

export const NegotiationAdviceSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  currentPrice: z.number(),
  targetPrice: z.number(),
  tips: z.array(NegotiationTipSchema),
  script: z.string(),
  cachedAt: z.number(),
})

export type NegotiationTip = z.infer<typeof NegotiationTipSchema>
export type NegotiationAdvice = z.infer<typeof NegotiationAdviceSchema>

export async function getNegotiationAdvice(itemId: string): Promise<NegotiationAdvice> {
  const data = await apiFetch<unknown>(`/api/shopping-items/${itemId}/negotiate`)
  return NegotiationAdviceSchema.parse(data)
}

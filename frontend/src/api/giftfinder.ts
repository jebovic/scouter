import { z } from 'zod'
import { apiFetch } from './client'

export const GiftSuggestionSchema = z.object({
  id: z.string(),
  name: z.string(),
  category: z.string(),
  priceRange: z.string(),
  estimatedPrice: z.number(),
  description: z.string(),
  occasion: z.string(),
  recipient: z.string(),
  platforms: z.array(z.string()),
  score: z.number(),
})

export const GiftResultSchema = z.object({
  missionId: z.string(),
  budget: z.number(),
  suggestions: z.array(GiftSuggestionSchema),
  totalFound: z.number(),
  generatedAt: z.string(),
})

export type GiftSuggestion = z.infer<typeof GiftSuggestionSchema>
export type GiftResult = z.infer<typeof GiftResultSchema>

export interface GiftSuggestionsParams {
  occasion?: string
  recipient?: string
  maxPrice?: number
}

export async function getGiftSuggestions(
  missionId: string,
  params: GiftSuggestionsParams = {},
): Promise<GiftResult> {
  const query = new URLSearchParams()
  if (params.occasion) query.set('occasion', params.occasion)
  if (params.recipient) query.set('recipient', params.recipient)
  if (params.maxPrice != null) query.set('maxPrice', String(params.maxPrice))
  const qs = query.toString()
  const raw = await apiFetch<unknown>(
    `/api/missions/${missionId}/gift-suggestions${qs ? `?${qs}` : ''}`,
  )
  return GiftResultSchema.parse(raw)
}

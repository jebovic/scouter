import { z } from 'zod'
import { apiFetch } from './client'

export const ReviewSummarySchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  overallScore: z.number(),
  totalReviews: z.number(),
  pros: z.array(z.string()),
  cons: z.array(z.string()),
  verdict: z.string(),
  recommendedFor: z.array(z.string()),
  avoidIf: z.array(z.string()),
  lastUpdated: z.string(),
})

export type ItemReviewSummary = z.infer<typeof ReviewSummarySchema>

export async function getReviewSummary(
  itemId: string,
  itemName: string,
): Promise<ItemReviewSummary> {
  const params = new URLSearchParams({ itemName })
  const data = await apiFetch<unknown>(
    `/api/items/${encodeURIComponent(itemId)}/review-summary?${params}`,
  )
  return ReviewSummarySchema.parse(data)
}

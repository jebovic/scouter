import { z } from 'zod'
import { apiFetch } from './client'

export const ReviewSummarySchema = z.object({
  optionId: z.string(),
  sentiment: z.number(),
  positivePct: z.number(),
  neutralPct: z.number(),
  negativePct: z.number(),
  pros: z.array(z.string()),
  cons: z.array(z.string()),
  totalReviews: z.number(),
  source: z.string(),
  refreshedAt: z.string(),
})

export type ReviewSummary = z.infer<typeof ReviewSummarySchema>

export async function getOptionReviews(optionId: string): Promise<ReviewSummary> {
  const data = await apiFetch<unknown>(`/api/options/${optionId}/reviews`)
  return ReviewSummarySchema.parse(data)
}

export async function refreshOptionReviews(optionId: string): Promise<ReviewSummary> {
  const data = await apiFetch<unknown>(`/api/options/${optionId}/reviews/refresh`, {
    method: 'POST',
  })
  return ReviewSummarySchema.parse(data)
}

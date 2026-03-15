import { z } from 'zod'
import { apiFetch } from './client'

export const DealExplanationSchema = z.object({
  itemId: z.string(),
  explanation: z.string(),
  tags: z.array(z.string()),
  generatedAt: z.string(),
})

export type DealExplanation = z.infer<typeof DealExplanationSchema>

export async function fetchDealExplanation(
  itemId: string,
  score?: number,
  trend?: string,
): Promise<DealExplanation> {
  const params = new URLSearchParams()
  if (score !== undefined) params.set('score', String(score))
  if (trend) params.set('trend', trend)
  const query = params.toString()
  const data = await apiFetch<unknown>(
    `/api/shopping-items/${itemId}/explain${query ? `?${query}` : ''}`,
  )
  return DealExplanationSchema.parse(data)
}

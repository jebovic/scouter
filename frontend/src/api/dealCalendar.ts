import { z } from 'zod'
import { apiFetch } from './client'

export const DealEventSchema = z.object({
  name: z.string(),
  startDate: z.string(),
  endDate: z.string(),
  categoryHint: z.string(),
  discountLevel: z.enum(['low', 'medium', 'high']),
  description: z.string(),
})
export type DealEvent = z.infer<typeof DealEventSchema>

const ResponseSchema = z.object({
  events: z.array(DealEventSchema),
  count: z.number(),
})

export async function listDealEvents(params?: {
  category?: string
  upcoming?: boolean
}): Promise<DealEvent[]> {
  const qs = new URLSearchParams()
  if (params?.category) qs.set('category', params.category)
  if (params?.upcoming !== undefined) qs.set('upcoming', String(params.upcoming))
  const query = qs.toString() ? `?${qs.toString()}` : ''
  const raw = await apiFetch<unknown>(`/api/deal-calendar${query}`)
  const parsed = ResponseSchema.parse(raw)
  return parsed.events
}

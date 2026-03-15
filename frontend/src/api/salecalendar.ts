import { z } from 'zod'
import { apiFetch } from './client'

export const SaleEventSchema = z.object({
  id: z.string(),
  name: z.string(),
  type: z.enum(['soldes', 'flash', 'fete', 'back_to_school']),
  startDate: z.string(),
  endDate: z.string(),
  daysUntil: z.number(),
  isActive: z.boolean(),
  discountRange: z.string(),
  categories: z.array(z.string()),
  tip: z.string(),
})
export type SaleEvent = z.infer<typeof SaleEventSchema>

const SaleCalendarSchema = z.object({
  events: z.array(SaleEventSchema),
  nextEvent: SaleEventSchema.nullable().optional(),
  activeEvent: SaleEventSchema.nullable().optional(),
  generatedAt: z.string(),
})
export type SaleCalendar = z.infer<typeof SaleCalendarSchema>

export async function getSaleCalendar(): Promise<SaleCalendar> {
  const raw = await apiFetch<unknown>('/api/sale-calendar')
  return SaleCalendarSchema.parse(raw)
}

import { z } from 'zod'
import { apiFetch } from './client'

const ItemRefSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  price: z.number(),
  status: z.string(),
  priority: z.number(),
})

const MerchantGroupSchema = z.object({
  merchant: z.string(),
  items: z.array(ItemRefSchema),
  totalPrice: z.number(),
  tripSaving: z.number(),
})

export const OptimizedListSchema = z.object({
  missionId: z.string(),
  merchantGroups: z.array(MerchantGroupSchema),
  totalItems: z.number(),
  totalPrice: z.number(),
  estimatedTrips: z.number(),
  summary: z.string(),
})

export type OptimizedList = z.infer<typeof OptimizedListSchema>
export type MerchantGroup = z.infer<typeof MerchantGroupSchema>
export type ItemRef = z.infer<typeof ItemRefSchema>

export async function getOptimizedList(missionId: string): Promise<OptimizedList> {
  return apiFetch<OptimizedList>(`/api/missions/${missionId}/optimized-list`)
}

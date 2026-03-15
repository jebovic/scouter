import { z } from 'zod'
import { apiFetch } from './client'

export const StockStatusSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  stockStatus: z.enum(['rupture', 'stock_faible', 'disponible', 'stock_élevé']),
  stockLabel: z.string(),
  estimatedUnitsLeft: z.number().nullable().optional(),
  restockDays: z.number().nullable().optional(),
  alert: z.string(),
  urgency: z.enum(['critical', 'high', 'low']),
})

export type StockStatus = z.infer<typeof StockStatusSchema>

export async function getStockStatus(
  missionId: string,
  itemId: string,
): Promise<StockStatus> {
  const raw = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/stock-status`,
  )
  return StockStatusSchema.parse(raw)
}

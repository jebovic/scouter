import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const StockStatusSchema = z.object({
  status: z.enum(['in_stock', 'limited', 'out_of_stock']),
  retailer: z.string(),
  updatedAt: z.coerce.date(),
})

export type StockStatusDTO = z.infer<typeof StockStatusSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function fetchStockStatus(
  product: string,
  retailer: string,
): Promise<StockStatusDTO> {
  const params = new URLSearchParams({ product, retailer })
  const data = await apiFetch<unknown>(`/api/stock-check?${params}`)
  return StockStatusSchema.parse(data)
}

import { z } from 'zod'
import { apiFetch, ApiError } from './client'

export const ProductInfoSchema = z.object({
  barcode: z.string(),
  name: z.string(),
  brand: z.string(),
  category: z.string(),
  imageUrl: z.string(),
  nutriScore: z.string(),
  ecoScore: z.string(),
  stores: z.array(z.string()),
  quantity: z.string(),
  source: z.string(),
})

export type ProductInfo = z.infer<typeof ProductInfoSchema>

export async function lookupProduct(barcode: string): Promise<ProductInfo> {
  try {
    const data = await apiFetch<unknown>(`/api/products/${barcode}`)
    return ProductInfoSchema.parse(data)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      throw new Error('NOT_FOUND')
    }
    throw err
  }
}

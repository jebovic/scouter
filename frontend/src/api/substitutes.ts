import { z } from 'zod'
import { apiFetch } from './client'

export const SubstituteSchema = z.object({
  name: z.string(),
  brand: z.string(),
  price: z.number(),
  currency: z.string(),
  retailer: z.string(),
  reason: z.string(),
  advantage: z.enum(['cheaper', 'better_rated', 'eco_friendly', 'local']),
  url: z.string().optional(),
  // Phase 79 enrichment fields
  savingsPercent: z.number().optional(),
  pros: z.array(z.string()).optional(),
  cons: z.array(z.string()).optional(),
  whyConsider: z.string().optional(),
})

export const SubstituteResponseSchema = z.object({
  optionId: z.string(),
  productName: z.string(),
  originalPrice: z.number().optional(),
  substitutes: z.array(SubstituteSchema),
  generatedAt: z.string(),
  cachedAt: z.number().optional(),
})

export type Substitute = z.infer<typeof SubstituteSchema>
export type SubstituteResponse = z.infer<typeof SubstituteResponseSchema>

export async function fetchSubstitutes(optionId: string): Promise<SubstituteResponse> {
  const data = await apiFetch<unknown>(`/api/options/${optionId}/substitutes`)
  return SubstituteResponseSchema.parse(data)
}

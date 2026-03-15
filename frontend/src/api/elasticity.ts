import { z } from 'zod'
import { apiFetch } from './client'

export const ElasticityResultSchema = z.object({
  itemId: z.string(),
  elasticity: z.number().nullable(),
  classification: z.string(),
  insight: z.string(),
  optimalBuyThreshold: z.number().nullable(),
  pricePoints: z.number().int(),
})

export type ElasticityResult = z.infer<typeof ElasticityResultSchema>

export async function getElasticity(
  missionId: string,
  itemId: string,
): Promise<ElasticityResult> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/elasticity`,
  )
  return ElasticityResultSchema.parse(data)
}

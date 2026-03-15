import { z } from 'zod'
import { apiFetch } from './client'

export const MerchantScoreSchema = z.object({
  merchant: z.string(),
  score: z.number(),
  avgPrice: z.number(),
  priceCount: z.number(),
  recommendation: z.string(),
})

export const MerchantRecommenderResponseSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  rankings: z.array(MerchantScoreSchema),
})

export type MerchantScore = z.infer<typeof MerchantScoreSchema>
export type MerchantRecommenderResponse = z.infer<typeof MerchantRecommenderResponseSchema>

export async function getMerchantRecommendations(
  missionId: string,
  itemId: string,
): Promise<MerchantRecommenderResponse> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/merchant-recommendations`,
  )
  return MerchantRecommenderResponseSchema.parse(data)
}

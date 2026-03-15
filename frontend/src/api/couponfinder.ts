import { z } from 'zod'
import { apiFetch } from './client'

export const CouponSchema = z.object({
  code: z.string(),
  description: z.string(),
  discountPct: z.number(),
  discountAmt: z.number(),
  source: z.string(),
  expiresIn: z.string(),
  confidence: z.string(),
})

export const CashbackOfferSchema = z.object({
  platform: z.string(),
  cashbackPct: z.number(),
  minPurchase: z.number(),
  description: z.string(),
})

export const CouponResultSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  merchant: z.string(),
  coupons: z.array(CouponSchema),
  cashbackOffers: z.array(CashbackOfferSchema),
  totalPotentialSaving: z.number(),
  generatedAt: z.string(),
})

export type Coupon = z.infer<typeof CouponSchema>
export type CashbackOffer = z.infer<typeof CashbackOfferSchema>
export type CouponResult = z.infer<typeof CouponResultSchema>

export async function getCoupons(
  missionId: string,
  itemId: string,
): Promise<CouponResult> {
  const raw = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/coupons`,
  )
  return CouponResultSchema.parse(raw)
}

import { z } from 'zod'

export const NegotiationScriptSchema = z.object({
  openingOffer: z.number(),
  walkAwayPrice: z.number(),
  suggestedDiscount: z.number(),
  talkingPoints: z.array(z.string()),
  counterOfferScript: z.array(z.string()),
  confidenceScore: z.number(),
})

export type NegotiationScript = z.infer<typeof NegotiationScriptSchema>

export async function fetchNegotiationScript(optionId: string): Promise<NegotiationScript> {
  const res = await fetch(`/api/options/${optionId}/negotiate`, { method: 'POST' })
  if (!res.ok) throw new Error('Failed to fetch negotiation script')
  return NegotiationScriptSchema.parse(await res.json())
}

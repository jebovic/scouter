import { z } from 'zod'
import { apiFetch } from './client'

export const ScriptSectionsSchema = z.object({
  opening: z.string(),
  argument: z.string(),
  counterOffer: z.string(),
  closing: z.string(),
})

export const NegotiationScriptResponseSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  script: ScriptSectionsSchema,
  tips: z.array(z.string()),
  estimatedSavings: z.number(),
})

export type ScriptSections = z.infer<typeof ScriptSectionsSchema>
export type NegotiationScriptResponse = z.infer<typeof NegotiationScriptResponseSchema>

export async function getNegotiationScriptContextual(
  missionId: string,
  itemId: string,
): Promise<NegotiationScriptResponse> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/negotiation-script`,
  )
  return NegotiationScriptResponseSchema.parse(data)
}

import { z } from 'zod'
import { apiFetch } from './client'

export const ScriptStepSchema = z.object({
  phase: z.enum(['opening', 'negotiation', 'closing']),
  speaker: z.enum(['vous', 'vendeur']),
  text: z.string(),
  tip: z.string(),
})

export const NegotiationScriptSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  currentPrice: z.number(),
  targetPrice: z.number(),
  potentialSavings: z.number(),
  scripts: z.array(ScriptStepSchema),
  tips: z.array(z.string()),
  generatedAt: z.string(),
})

export type ScriptStep = z.infer<typeof ScriptStepSchema>
export type NegotiationScript = z.infer<typeof NegotiationScriptSchema>

export async function getNegotiationScript(
  missionId: string,
  itemId: string,
): Promise<NegotiationScript> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/negotiation-script`,
  )
  return NegotiationScriptSchema.parse(data)
}

import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const EnvelopeSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  monthlyAmount: z.number(),
  currency: z.string(),
  createdAt: z.string(),
})

export type EnvelopeDTO = z.infer<typeof EnvelopeSchema>

const EnvelopeListSchema = z.object({
  items: z.array(EnvelopeSchema),
})

// ── API functions ────────────────────────────────────────────────────────────

export async function listEnvelopes(): Promise<EnvelopeDTO[]> {
  const data = await apiFetch<unknown>('/api/envelopes')
  const { items } = EnvelopeListSchema.parse(data)
  return items
}

export async function createEnvelope(payload: {
  name: string
  monthlyAmount: number
  currency: string
}): Promise<EnvelopeDTO> {
  const data = await apiFetch<unknown>('/api/envelopes', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return EnvelopeSchema.parse(data)
}

export async function updateEnvelope(
  id: string,
  payload: { name?: string; monthlyAmount?: number; currency?: string },
): Promise<EnvelopeDTO> {
  const data = await apiFetch<unknown>(`/api/envelopes/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
  return EnvelopeSchema.parse(data)
}

export async function deleteEnvelope(id: string): Promise<void> {
  await apiFetch<unknown>(`/api/envelopes/${id}`, { method: 'DELETE' })
}

// ── Envelope summary ─────────────────────────────────────────────────────────

export const EnvelopeSummarySchema = z.object({
  envelope: EnvelopeSchema,
  missionCount: z.number(),
  totalBudget: z.number(),
  totalSpent: z.number(),
})

export type EnvelopeSummaryDTO = z.infer<typeof EnvelopeSummarySchema>

export async function getEnvelopeSummary(id: string): Promise<EnvelopeSummaryDTO> {
  const data = await apiFetch<unknown>(`/api/envelopes/${id}/summary`)
  return EnvelopeSummarySchema.parse(data)
}

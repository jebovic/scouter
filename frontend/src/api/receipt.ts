import { z } from 'zod'
import { apiFetch } from './client'

export const ReceiptItemSchema = z.object({
  name: z.string(),
  quantity: z.number(),
  unitPrice: z.number(),
  total: z.number(),
})

export const ReceiptAnalysisSchema = z.object({
  id: z.string(),
  missionSlug: z.string(),
  rawText: z.string(),
  items: z.array(ReceiptItemSchema),
  total: z.number(),
  currency: z.string(),
  analyzedAt: z.string(),
})

export type ReceiptItem = z.infer<typeof ReceiptItemSchema>
export type ReceiptAnalysis = z.infer<typeof ReceiptAnalysisSchema>

export async function analyzeReceipt(
  slug: string,
  rawText: string,
  currency = 'EUR',
): Promise<ReceiptAnalysis> {
  const data = await apiFetch<unknown>(`/api/missions/${slug}/receipts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ rawText, currency }),
  })
  return ReceiptAnalysisSchema.parse(data)
}

export async function listReceipts(slug: string): Promise<ReceiptAnalysis[]> {
  const data = await apiFetch<unknown>(`/api/missions/${slug}/receipts`)
  return z.array(ReceiptAnalysisSchema).parse(data)
}

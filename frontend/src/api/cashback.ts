import { z } from 'zod'
import { apiFetch } from './client'

export const CashbackEntrySchema = z.object({
  id: z.string().uuid(),
  userRef: z.string(),
  itemName: z.string(),
  merchant: z.string(),
  amount: z.number(),
  percentRate: z.number(),
  status: z.enum(['pending', 'confirmed', 'paid']),
  notes: z.string().optional(),
  claimedAt: z.string().nullish(),
  paidAt: z.string().nullish(),
  createdAt: z.string(),
})

export const CashbackSummarySchema = z.object({
  totalPending: z.number(),
  totalConfirmed: z.number(),
  totalPaid: z.number(),
  entryCount: z.number(),
})

export type CashbackEntry = z.infer<typeof CashbackEntrySchema>
export type CashbackSummary = z.infer<typeof CashbackSummarySchema>

export interface CreateCashbackPayload {
  itemName: string
  merchant: string
  amount: number
  percentRate: number
  status?: string
  notes?: string
}

export interface UpdateCashbackPayload {
  status?: string
  paidAt?: string
  notes?: string
}

export async function listCashbackEntries(userRef: string): Promise<CashbackEntry[]> {
  const raw = await apiFetch<unknown>(`/api/cashback/?userRef=${encodeURIComponent(userRef)}`)
  return z.array(CashbackEntrySchema).parse(raw)
}

export async function getCashbackSummary(userRef: string): Promise<CashbackSummary> {
  const raw = await apiFetch<unknown>(`/api/cashback/summary?userRef=${encodeURIComponent(userRef)}`)
  return CashbackSummarySchema.parse(raw)
}

export async function createCashbackEntry(
  userRef: string,
  payload: CreateCashbackPayload,
): Promise<CashbackEntry> {
  const raw = await apiFetch<unknown>('/api/cashback/', {
    method: 'POST',
    headers: { 'X-Collaborator-ID': userRef },
    body: JSON.stringify(payload),
  })
  return CashbackEntrySchema.parse(raw)
}

export async function updateCashbackEntry(
  id: string,
  payload: UpdateCashbackPayload,
): Promise<CashbackEntry> {
  const raw = await apiFetch<unknown>(`/api/cashback/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
  return CashbackEntrySchema.parse(raw)
}

export async function deleteCashbackEntry(id: string): Promise<void> {
  await apiFetch<void>(`/api/cashback/${id}`, { method: 'DELETE' })
}

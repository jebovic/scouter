import { z } from 'zod'
import { apiFetch, ApiError } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const PersonaSchema = z.object({
  id: z.string(),
  archetype: z.string(),
  traits: z.array(z.string()),
  tips: z.array(z.string()),
  summary: z.string(),
  createdAt: z.string(),
})

export type PersonaResult = z.infer<typeof PersonaSchema>

export const ShoppingPersonaSchema = z.object({
  archetype: z.string(),
  icon: z.string(),
  description: z.string(),
  strengths: z.array(z.string()),
  blindspots: z.array(z.string()),
  tips: z.array(z.string()),
  scoreProfile: z.record(z.string(), z.number()),
  cachedAt: z.number(),
})

export type ShoppingPersonaResult = z.infer<typeof ShoppingPersonaSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function fetchPersona(): Promise<PersonaResult | null> {
  try {
    const data = await apiFetch<unknown>('/api/persona')
    return PersonaSchema.parse(data)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      return null
    }
    throw err
  }
}

export async function runPersona(): Promise<PersonaResult> {
  const data = await apiFetch<unknown>('/api/persona', {
    method: 'POST',
  })
  return PersonaSchema.parse(data)
}

export async function getShoppingPersona(): Promise<ShoppingPersonaResult> {
  const data = await apiFetch<unknown>('/api/persona/shopping')
  return ShoppingPersonaSchema.parse(data)
}

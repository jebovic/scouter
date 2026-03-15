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

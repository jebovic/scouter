import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const SettingsSchema = z.object({
  currency: z.string().default('EUR'),
  locale: z.string().default('fr-FR'),
  llm_provider: z.string().default('ollama'),
}).catchall(z.unknown())

export type Settings = z.infer<typeof SettingsSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function getSettings(): Promise<Settings> {
  const data = await apiFetch<unknown>('/api/settings')
  return SettingsSchema.parse(data)
}

export async function updateSettings(
  patch: Partial<Record<string, unknown>>,
): Promise<Settings> {
  const body: Record<string, unknown> = {}
  for (const [k, v] of Object.entries(patch)) {
    body[k] = v
  }
  const data = await apiFetch<unknown>('/api/settings', {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
  return SettingsSchema.parse(data)
}

export async function deleteAllData(): Promise<void> {
  await apiFetch<void>('/api/data', {
    method: 'DELETE',
    headers: { 'X-Confirm': 'delete-all' },
  })
}

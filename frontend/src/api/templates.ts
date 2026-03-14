import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

const ConstraintSchema = z.object({
  key: z.string(),
  label: z.string(),
  value: z.unknown(),
  type: z.enum(['hard', 'soft']),
})

const WeightProfileSchema = z.object({
  price: z.number(),
  quality: z.number(),
  feature: z.number(),
})

const BudgetRangeSchema = z.object({
  min: z.number(),
  max: z.number(),
})

export const TemplateSchema = z.object({
  slug: z.string(),
  name: z.string(),
  icon: z.string(),
  category: z.enum(['travel', 'electronics', 'computing', 'renovation', 'custom']),
  description: z.string(),
  constraints: z.array(ConstraintSchema),
  costCategories: z.array(z.string()),
  weightProfile: WeightProfileSchema,
  suggestedBudget: BudgetRangeSchema.optional(),
})

// ── API functions ────────────────────────────────────────────────────────────

export async function listTemplates(): Promise<z.infer<typeof TemplateSchema>[]> {
  const data = await apiFetch<unknown>('/api/templates')
  const { items } = z.object({ items: z.array(TemplateSchema) }).parse(data)
  return items
}

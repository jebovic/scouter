import { z } from 'zod'
import { apiFetch } from './client'

export const ItemTagsSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  suggestedCategory: z.string(),
  tags: z.array(z.string()),
  brand: z.string(),
  productType: z.string(),
  isLuxury: z.boolean(),
  isTech: z.boolean(),
  warrantyHint: z.string(),
  refurbishedAvailable: z.boolean(),
  generatedAt: z.string(),
})

export type ItemTags = z.infer<typeof ItemTagsSchema>

export async function getItemTags(missionId: string, itemId: string): Promise<ItemTags> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/items/${itemId}/tags`)
  return ItemTagsSchema.parse(data)
}

import { z } from 'zod'
import { apiFetch } from './client'

export const AnnotationTypeSchema = z.enum([
  'new_low',
  'rebound',
  'spike_up',
  'spike_down',
  'stable',
])

export const AnnotationSchema = z.object({
  recordedAt: z.string(),
  price: z.number(),
  type: AnnotationTypeSchema,
  label: z.string(),
  description: z.string(),
})

export const PriceAnnotationsSchema = z.object({
  itemId: z.string(),
  annotations: z.array(AnnotationSchema),
})

export type AnnotationType = z.infer<typeof AnnotationTypeSchema>
export type Annotation = z.infer<typeof AnnotationSchema>
export type PriceAnnotations = z.infer<typeof PriceAnnotationsSchema>

export async function getPriceAnnotations(
  missionId: string,
  itemId: string,
): Promise<PriceAnnotations> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/price-annotations`,
  )
  return PriceAnnotationsSchema.parse(data)
}

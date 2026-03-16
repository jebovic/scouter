// frontend/src/api/images.ts
import { z } from "zod";

export const OptionImageSchema = z.object({
  id: z.string().uuid(),
  contentType: z.string(),
  width: z.number().optional(),
  height: z.number().optional(),
  sortOrder: z.number(),
  url: z.string().url(),
});

export type OptionImage = z.infer<typeof OptionImageSchema>;

export async function listOptionImages(
  missionId: string,
  optionId: string
): Promise<OptionImage[]> {
  const res = await fetch(
    `/api/missions/${missionId}/options/${optionId}/images`
  );
  if (!res.ok) throw new Error(`listOptionImages: ${res.status}`);
  const data = await res.json();
  return z.array(OptionImageSchema).parse(data);
}

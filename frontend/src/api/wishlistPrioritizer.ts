import { z } from 'zod';

export const PrioritizedItemSchema = z.object({
  id: z.string(),
  name: z.string(),
  score: z.number(),
  rank: z.number(),
  verdict: z.string(),
  urgency: z.number(),
  trend: z.number(),
  budgetFit: z.number(),
});

export const TopPickSchema = z.object({
  id: z.string(),
  name: z.string(),
});

export const PrioritizedListSchema = z.object({
  items: z.array(PrioritizedItemSchema),
  topPick: TopPickSchema.nullable(),
  summary: z.string(),
});

export type PrioritizedItem = z.infer<typeof PrioritizedItemSchema>;
export type TopPick = z.infer<typeof TopPickSchema>;
export type PrioritizedList = z.infer<typeof PrioritizedListSchema>;

export async function fetchWishlistPrioritized(): Promise<PrioritizedList> {
  const res = await fetch('/api/wishlist/prioritized');
  if (!res.ok) throw new Error(`fetchWishlistPrioritized: ${res.status}`);
  const data = await res.json();
  return PrioritizedListSchema.parse(data);
}

import { z } from 'zod';

export const QuantityTierSchema = z.object({
  quantity: z.number(),
  unitPrice: z.number(),
  totalPrice: z.number(),
  discountPct: z.number(),
  budgetFit: z.boolean(),
  recommendation: z.string(),
});

export const OptimizerReportSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  basePrice: z.number(),
  budget: z.number().nullable(),
  optimalQty: z.number(),
  tiers: z.array(QuantityTierSchema),
  summary: z.string(),
});

export type QuantityTier = z.infer<typeof QuantityTierSchema>;
export type OptimizerReport = z.infer<typeof OptimizerReportSchema>;

export async function fetchQuantityOptimizer(
  missionId: string,
  itemId: string
): Promise<OptimizerReport> {
  const res = await fetch(`/api/missions/${missionId}/items/${itemId}/quantity-optimizer`);
  if (!res.ok) throw new Error(`fetchQuantityOptimizer: ${res.status}`);
  const data = await res.json();
  return OptimizerReportSchema.parse(data);
}

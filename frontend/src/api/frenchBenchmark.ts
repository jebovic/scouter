import { z } from 'zod';

export const BenchmarkResultSchema = z.object({
  missionId: z.string(),
  missionName: z.string(),
  medianPrice: z.number(),
  yourAvgPrice: z.number(),
  verdict: z.string(),
  verdictCode: z.enum(['bon_prix', 'prix_moyen', 'au_dessus_du_marche']),
  sampleSize: z.number(),
  priceRange: z.tuple([z.number(), z.number()]),
  tips: z.array(z.string()),
});

export type BenchmarkResult = z.infer<typeof BenchmarkResultSchema>;

export async function fetchFrenchBenchmark(missionId: string): Promise<BenchmarkResult> {
  const res = await fetch(`/api/missions/${missionId}/french-benchmark`);
  if (!res.ok) throw new Error(`fetchFrenchBenchmark: ${res.status}`);
  const data = await res.json();
  return BenchmarkResultSchema.parse(data);
}

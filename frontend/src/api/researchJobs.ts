import { z } from 'zod'

export const ResearchJobSchema = z.object({
  id: z.string().uuid(),
  status: z.enum(['pending', 'running', 'done', 'failed']),
  options_count: z.number().int().nullable(),
  error: z.string().nullable(),
  feedback: z.string().nullable(),
  started_at: z.string().nullable(),
  completed_at: z.string().nullable(),
  created_at: z.string(),
})

export type ResearchJob = z.infer<typeof ResearchJobSchema>

export const ResearchJobListSchema = z.object({
  jobs: z.array(ResearchJobSchema),
})

export type ResearchJobList = z.infer<typeof ResearchJobListSchema>

export const TriggerResearchResponseSchema = z.object({
  job_id: z.string().uuid(),
  status: z.literal('pending'),
})

export type TriggerResearchResponse = z.infer<typeof TriggerResearchResponseSchema>

export async function triggerResearch(missionId: string, feedback?: string): Promise<TriggerResearchResponse> {
  const res = await fetch(`/api/missions/${missionId}/research`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ feedback: feedback ?? '' }),
  })
  if (res.status === 409) {
    throw new Error('already_running')
  }
  if (!res.ok) {
    throw new Error(`Failed to trigger research: ${res.status}`)
  }
  return TriggerResearchResponseSchema.parse(await res.json())
}

export async function fetchResearchJobs(missionId: string): Promise<ResearchJobList> {
  const res = await fetch(`/api/missions/${missionId}/research/jobs`)
  if (!res.ok) {
    throw new Error(`Failed to fetch research jobs: ${res.status}`)
  }
  return ResearchJobListSchema.parse(await res.json())
}

export async function fetchResearchJob(missionId: string, jobId: string): Promise<ResearchJob> {
  const res = await fetch(`/api/missions/${missionId}/research/jobs/${jobId}`)
  if (!res.ok) {
    throw new Error(`Failed to fetch research job: ${res.status}`)
  }
  return ResearchJobSchema.parse(await res.json())
}

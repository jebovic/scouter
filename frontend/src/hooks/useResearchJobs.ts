import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { fetchResearchJobs } from '../api/researchJobs'
import { useToast } from '../components/scouter'
import type { ResearchJob } from '../api/researchJobs'

export function useResearchJobs(missionId: string) {
  const { t } = useTranslation()
  const { toast } = useToast()
  const queryClient = useQueryClient()
  const prevStatusRef = useRef<Record<string, string>>({})

  const query = useQuery({
    queryKey: ['research-jobs', missionId],
    queryFn: () => fetchResearchJobs(missionId),
    refetchInterval: (query) => {
      const jobs = query.state.data?.jobs
      if (!jobs) return false
      return jobs.some(j => j.status === 'running' || j.status === 'pending') ? 3000 : false
    },
  })

  useEffect(() => {
    const jobs = query.data?.jobs
    if (!jobs) return
    const prev = prevStatusRef.current
    jobs.forEach((job: ResearchJob) => {
      const prevStatus = prev[job.id]
      if (prevStatus && prevStatus !== job.status) {
        if (job.status === 'done') {
          queryClient.invalidateQueries({ queryKey: ['options', missionId] })
          toast(t('research.complete', { count: job.options_count ?? 0 }), 'success')
        } else if (job.status === 'failed') {
          toast(t('research.failed', { error: job.error ?? '' }), 'error')
        }
      }
    })
    prevStatusRef.current = Object.fromEntries(jobs.map(j => [j.id, j.status]))
  }, [query.data?.jobs, missionId, queryClient, t, toast])

  return query
}

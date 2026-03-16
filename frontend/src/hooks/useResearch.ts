import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { triggerResearch } from '../api/researchJobs'
import { useToast } from '../components/scouter'

export function useTriggerResearch(missionId: string) {
  const queryClient = useQueryClient()
  const { t } = useTranslation()
  const { toast } = useToast()

  return useMutation({
    mutationFn: (feedback?: string) => triggerResearch(missionId, feedback),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['research-jobs', missionId] })
    },
    onError: (err: Error) => {
      if (err.message === 'already_running') {
        toast(t('research.alreadyRunning'), 'error')
      } else {
        toast(t('research.triggerFailed', { message: err.message }), 'error')
      }
    },
  })
}

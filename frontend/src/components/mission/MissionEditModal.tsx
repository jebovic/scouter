import { useTranslation } from 'react-i18next'
import { MissionForm } from './MissionForm'
import type { Mission, MissionCreateRequest } from '../../types'
import styles from './MissionEditModal.module.css'

interface MissionEditModalProps {
  mission: Mission
  onSave: (updates: Partial<MissionCreateRequest>) => void
  onClose: () => void
  loading?: boolean
  error?: string
}

export function MissionEditModal({ mission, onSave, onClose, loading, error }: MissionEditModalProps) {
  const { t } = useTranslation()

  function handleBackdropClick(e: React.MouseEvent<HTMLDivElement>) {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  return (
    <div className={styles.backdrop} onClick={handleBackdropClick}>
      <div className={styles.modal}>
        <div className={styles.header}>
          <h2 className={styles.title}>{t('mission.actions.edit')}</h2>
          <button className={styles.closeBtn} onClick={onClose} aria-label={t('common.close')}>
            ×
          </button>
        </div>
        <MissionForm
          initialValues={{
            name: mission.name,
            icon: mission.icon,
            category: mission.category,
            budget: mission.budget,
            currency: mission.currency,
            constraints: mission.constraints,
            envelopeId: mission.envelopeId,
          }}
          onSubmit={onSave}
          onCancel={onClose}
          loading={loading}
          error={error}
        />
      </div>
    </div>
  )
}

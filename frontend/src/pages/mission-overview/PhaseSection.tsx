import { useTranslation } from 'react-i18next'
import type { Mission, MissionPhase } from '../../types'
import styles from '../MissionOverview.module.css'

const PHASES: MissionPhase[] = ['researching', 'comparing', 'buying', 'done']

interface PhaseSectionProps {
  mission: Mission
  updatePending: boolean
  handlePhase: (phase: MissionPhase) => void
}

export function PhaseSection({ mission, updatePending, handlePhase }: PhaseSectionProps) {
  const { t } = useTranslation()
  return (
    <div className={styles.card}>
      <h3 className={styles.cardLabel}>{t('mission.phase').toUpperCase()}</h3>
      <div className={styles.phaseButtons}>
        {PHASES.map((p) => (
          <button
            key={p}
            onClick={() => handlePhase(p)}
            disabled={updatePending}
            className={`${styles.phaseBtn} ${mission.phase === p ? styles.phaseBtnActive : ''}`}
          >
            {t(`mission.phases.${p}`)}
          </button>
        ))}
      </div>
    </div>
  )
}

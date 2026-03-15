import { useTranslation } from 'react-i18next'
import { useTimingAdvice } from '../../hooks'
import styles from './TimingAdvisorCard.module.css'

interface Props {
  missionId: string
}

const REC_ICON: Record<string, string> = {
  BUY_NOW: '✅',
  WAIT: '⏳',
  UNSURE: '❓',
}

const REC_CLASS: Record<string, string> = {
  BUY_NOW: styles.recBuyNow,
  WAIT: styles.recWait,
  UNSURE: styles.recUnsure,
}

export function TimingAdvisorCard({ missionId }: Props) {
  const { t } = useTranslation()
  const { fetchAdvice, isPending, advice } = useTimingAdvice(missionId)

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.title}>{t('timing.title')}</span>
        <button
          className={styles.analyzeBtn}
          disabled={isPending}
          onClick={() => fetchAdvice()}
        >
          {isPending ? t('timing.analyzing') : advice ? t('timing.reanalyze') : t('timing.analyze')}
        </button>
      </div>

      {!advice && !isPending && (
        <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
          {t('timing.prompt')}
        </div>
      )}

      {isPending && (
        <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
          {t('timing.analyzing')}
        </div>
      )}

      {advice && !isPending && (
        <>
          <div className={`${styles.recBadge} ${REC_CLASS[advice.recommendation] ?? styles.recUnsure}`}>
            {REC_ICON[advice.recommendation]} {t(`timing.rec.${advice.recommendation}`)}
          </div>

          <div className={styles.rationale}>{advice.rationale}</div>

          <div className={styles.meta}>
            {advice.waitUntil && (
              <span className={styles.metaChip}>
                📅 {t('timing.waitUntil')}: {advice.waitUntil}
              </span>
            )}
            {advice.nextSaleEvent && (
              <span className={styles.metaChip}>
                🏷️ {advice.nextSaleEvent}
              </span>
            )}
          </div>

          <div className={styles.confidence}>
            {t('timing.confidence')}: {Math.round(advice.confidence * 100)}%
          </div>
        </>
      )}
    </div>
  )
}

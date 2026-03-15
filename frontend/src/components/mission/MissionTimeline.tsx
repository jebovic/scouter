import { useTranslation } from 'react-i18next'
import { useSettings } from '../../hooks/useSettings'
import { formatDate } from '../../utils/format'
import styles from './MissionTimeline.module.css'

interface TimelineEvent {
  label: string
  date?: string
  done: boolean
}

interface MissionTimelineProps {
  createdAt: string
  hasDecision: boolean
  decisionAt?: string
  hasPurchase: boolean
  purchasedAt?: string
  hasReview: boolean
}

export function MissionTimeline({
  createdAt,
  hasDecision,
  decisionAt,
  hasPurchase,
  purchasedAt,
  hasReview,
}: MissionTimelineProps) {
  const { t } = useTranslation()
  const { data: settings } = useSettings()
  const locale = settings?.locale ?? 'fr-FR'

  const events: TimelineEvent[] = [
    { label: t('timeline.researchStarted'), date: createdAt, done: true },
    { label: t('timeline.decisionRun'), date: decisionAt, done: hasDecision },
    { label: t('timeline.purchaseRecorded'), date: purchasedAt, done: hasPurchase },
    { label: t('timeline.reviewed'), done: hasReview },
  ]

  return (
    <div className={styles.timeline}>
      {events.map((ev, i) => (
        <div key={i} className={`${styles.step} ${ev.done ? styles.done : styles.pending}`}>
          <div className={styles.iconWrap}>
            <div className={styles.dot}>{ev.done ? '✓' : String(i + 1)}</div>
            {i < events.length - 1 && <div className={styles.line} />}
          </div>
          <div className={styles.content}>
            <span className={styles.label}>{ev.label}</span>
            {ev.date && <span className={styles.date}>{formatDate(ev.date, locale)}</span>}
          </div>
        </div>
      ))}
    </div>
  )
}

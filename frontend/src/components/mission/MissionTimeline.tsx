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

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

export function MissionTimeline({
  createdAt,
  hasDecision,
  decisionAt,
  hasPurchase,
  purchasedAt,
  hasReview,
}: MissionTimelineProps) {
  const events: TimelineEvent[] = [
    { label: 'Research started', date: createdAt, done: true },
    { label: 'Decision run', date: decisionAt, done: hasDecision },
    { label: 'Purchase recorded', date: purchasedAt, done: hasPurchase },
    { label: 'Reviewed', done: hasReview },
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
            {ev.date && <span className={styles.date}>{formatDate(ev.date)}</span>}
          </div>
        </div>
      ))}
    </div>
  )
}

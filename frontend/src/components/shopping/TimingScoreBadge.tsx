import { useState } from 'react'
import { useTimingScore } from '../../hooks/useTimingScore'
import type { Signal } from '../../api/timingrecommender'
import styles from './TimingScoreBadge.module.css'

interface TimingScoreBadgeProps {
  missionId: string
  itemId: string
}

const REC_EMOJI: Record<string, string> = {
  buy_now: '🟢',
  watch: '🟡',
  wait: '🔴',
}

export function TimingScoreBadge({ missionId, itemId }: TimingScoreBadgeProps) {
  const [open, setOpen] = useState(false)
  const { data, isLoading } = useTimingScore(missionId, itemId)

  if (isLoading || !data) return null

  const urgencyClass = styles[data.urgencyLevel as keyof typeof styles] ?? styles.low
  const emoji = REC_EMOJI[data.recommendation] ?? '⏳'

  return (
    <div className={`${styles.wrapper} ${urgencyClass}`}>
      <button
        className={styles.badge}
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-label={`Score d'achat : ${data.timingScore}/100 — ${data.verdict}`}
        title={data.verdict}
      >
        <span className={styles.circle}>{data.timingScore}</span>
        {emoji}
      </button>

      {open && (
        <div className={styles.panel} role="region" aria-label="Détail du score d'achat">
          <div className={styles.verdict}>{data.verdict}</div>
          <div className={styles.window}>🗓 {data.bestBuyWindow}</div>
          <div className={styles.signals}>
            {data.signals.map((s: Signal) => (
              <div key={s.name} className={styles.signalRow}>
                <span className={styles.signalLabel} title={s.label}>{s.label}</span>
                <div className={styles.signalBarTrack}>
                  <div
                    className={styles.signalBar}
                    style={{ '--bar-width': `${s.score}%` } as React.CSSProperties}
                  />
                </div>
                <span className={styles.signalScore}>{s.score}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

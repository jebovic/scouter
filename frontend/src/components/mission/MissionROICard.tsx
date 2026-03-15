import React, { useEffect, useRef, useState } from 'react'
import { useMissionROI } from '../../hooks/useMissionROI'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import styles from './MissionROICard.module.css'

interface Props {
  missionId: string | undefined
}

const RADIUS = 32
const CIRCUMFERENCE = 2 * Math.PI * RADIUS

function scoreColor(score: number): string {
  if (score >= 80) return 'var(--green)'
  if (score >= 50) return 'var(--cyan)'
  if (score >= 20) return 'var(--gold)'
  if (score >= 0) return 'var(--orange)'
  return 'var(--coral)'
}

function fmtDecimal(v: number): string {
  return v.toFixed(1)
}

export function MissionROICard({ missionId }: Props) {
  const { data, isLoading, isError } = useMissionROI(missionId)
  const { fmt } = useFormatCurrency()
  const [animated, setAnimated] = useState(false)
  const mountedRef = useRef(false)

  useEffect(() => {
    if (data && !mountedRef.current) {
      mountedRef.current = true
      // Trigger animation on next frame so CSS transition fires
      requestAnimationFrame(() => setAnimated(true))
    }
  }, [data])

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="ROI de la mission">
        <div className={styles.header}>
          <span className={styles.title}>ROI Prospection</span>
        </div>
        <div className={styles.hero}>
          <div className={`${styles.skeleton} ${styles.skeletonCircle}`} />
          <div className={styles.skeletonRight}>
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className={`${styles.skeleton} ${styles.skeletonLine}`} />
            ))}
          </div>
        </div>
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="ROI de la mission">
        <div className={styles.header}>
          <span className={styles.title}>ROI Prospection</span>
        </div>
        <p className={styles.errorState}>Impossible de calculer le ROI.</p>
      </section>
    )
  }

  const clampedScore = Math.max(-100, Math.min(100, data.roiScore))
  // Normalise to 0-100 for the ring fill (negative scores show empty ring)
  const fillPct = Math.max(0, clampedScore)
  const dashOffset = CIRCUMFERENCE * (1 - (animated ? fillPct / 100 : 0))
  const color = scoreColor(data.roiScore)

  return (
    <section className={styles.card} aria-label="ROI de la mission">
      <div className={styles.header}>
        <span className={styles.title}>ROI Prospection</span>
      </div>

      <div className={styles.hero}>
        {/* Circular score badge */}
        <div className={styles.ringWrap} aria-label={`Score ROI : ${Math.round(data.roiScore)}`}>
          <svg width="80" height="80" viewBox="0 0 80 80" aria-hidden="true">
            <circle
              className={styles.ringBg}
              cx="40"
              cy="40"
              r={RADIUS}
              strokeWidth="7"
            />
            <circle
              className={styles.ringFill}
              cx="40"
              cy="40"
              r={RADIUS}
              strokeWidth="7"
              strokeDasharray={CIRCUMFERENCE}
              strokeDashoffset={dashOffset}
              style={{ stroke: color } as React.CSSProperties}
            />
          </svg>
          <div className={styles.ringLabel}>
            <span className={styles.ringScore} style={{ color } as React.CSSProperties}>
              {Math.round(data.roiScore)}
            </span>
            <span className={styles.ringUnit}>/100</span>
          </div>
        </div>

        {/* Metrics grid */}
        <div className={styles.grid}>
          <div className={styles.metric}>
            <span className={styles.metricValue} data-positive={data.savingsVsEstimate >= 0 ? 'true' : 'false'}>
              {data.savingsVsEstimate >= 0 ? '+' : ''}{fmt(data.savingsVsEstimate)}
            </span>
            <span className={styles.metricLabel}>Économies vs estimé</span>
          </div>
          <div className={styles.metric}>
            <span className={styles.metricValue}>{data.daysTracking}j</span>
            <span className={styles.metricLabel}>Jours suivis</span>
          </div>
          <div className={styles.metric}>
            <span className={styles.metricValue}>{data.priceChecks}</span>
            <span className={styles.metricLabel}>Vérif. de prix</span>
          </div>
          <div className={styles.metric}>
            <span className={styles.metricValue}>{fmtDecimal(data.avgChecksPerDay)}/j</span>
            <span className={styles.metricLabel}>Moy. par jour</span>
          </div>
        </div>
      </div>

      {/* Verdict */}
      <div className={styles.verdict} style={{ '--verdict-color': color } as React.CSSProperties}>
        {data.verdict}
      </div>
    </section>
  )
}

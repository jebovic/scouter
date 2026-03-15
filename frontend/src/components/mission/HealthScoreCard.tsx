import React from 'react'
import { useMissionHealth } from '../../hooks/useMissionHealth'
import styles from './HealthScoreCard.module.css'

interface Props {
  missionSlug: string | undefined
}

export function HealthScoreCard({ missionSlug }: Props) {
  const { data, isLoading, isError } = useMissionHealth(missionSlug)

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Santé de la mission">
        <div className={styles.header}>
          <span className={styles.title}>Santé de la mission</span>
        </div>
        <div className={styles.loading} data-testid="health-loading">
          <div className={styles.scoreHero}>
            <div className={`${styles.skeleton} ${styles.skeletonCircle}`} />
            <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div className={`${styles.skeleton} ${styles.skeletonLine}`} style={{ width: '40%' }} />
              <div className={`${styles.skeleton} ${styles.skeletonLine}`} style={{ width: '60%' }} />
            </div>
          </div>
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className={`${styles.skeleton} ${styles.skeletonLine}`} />
          ))}
        </div>
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="Santé de la mission">
        <div className={styles.header}>
          <span className={styles.title}>Santé de la mission</span>
        </div>
        <p className={styles.errorState}>Impossible de charger le score de santé.</p>
      </section>
    )
  }

  return (
    <section className={styles.card} aria-label="Santé de la mission">
      <div className={styles.header}>
        <span className={styles.title}>Santé de la mission</span>
      </div>

      {/* Score Hero */}
      <div className={styles.scoreHero}>
        <div className={styles.scoreCircle}>
          <span className={styles.scoreValue}>{data.overallScore}</span>
          <span className={styles.scoreDenominator}>/100</span>
        </div>
        <div className={styles.gradeLabel}>
          <span className={styles.grade} data-grade={data.grade}>{data.grade}</span>
          <span className={styles.gradeSubLabel}>Score global</span>
        </div>
      </div>

      {/* Signals */}
      <div className={styles.signals}>
        {data.signals.map((signal) => (
          <div key={signal.name} className={styles.signal}>
            <span className={styles.signalIcon}>{signal.icon}</span>
            <div className={styles.signalBar}>
              <span className={styles.signalName}>{signal.name}</span>
              <div className={styles.signalTrack}>
                <div
                  className={styles.signalFill}
                  style={{ '--pct': signal.score } as React.CSSProperties}
                />
              </div>
            </div>
            <span className={styles.signalScore}>{signal.score}</span>
          </div>
        ))}
      </div>

      {/* AI Summary + Advice */}
      {(data.summary || data.advice) && (
        <div className={styles.aiBlock}>
          {data.summary && (
            <>
              <span className={styles.aiLabel}>Analyse IA</span>
              <p className={styles.aiText}>{data.summary}</p>
            </>
          )}
          {data.advice && (
            <p className={styles.adviceText}>{data.advice}</p>
          )}
        </div>
      )}
    </section>
  )
}

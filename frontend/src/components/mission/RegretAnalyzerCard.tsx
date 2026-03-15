import React, { useMemo } from 'react'
import { useRegretAnalyzer } from '../../hooks/useRegretAnalyzer'
import styles from './RegretAnalyzerCard.module.css'

interface Props {
  missionId: string | undefined
}

export function RegretAnalyzerCard({ missionId }: Props) {
  const { data, isLoading, error } = useRegretAnalyzer(missionId)

  // Calculate arc angles for gauge based on regret score
  const arcDegrees = useMemo(() => {
    if (!data) return 0
    // Map 0-100 score to 0-180 degrees (semicircle)
    return (data.regretScore / 100) * 180
  }, [data])

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Analyse de regret">
        <div className={styles.header}>
          <span className={styles.title}>Analyse de regret</span>
        </div>
        <div className={styles.loading}>
          <div className={`${styles.skeleton} ${styles.skeletonCircle}`} />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div className={`${styles.skeleton} ${styles.skeletonLine}`} style={{ width: '40%' }} />
            <div className={`${styles.skeleton} ${styles.skeletonLine}`} style={{ width: '60%' }} />
          </div>
        </div>
      </section>
    )
  }

  if (error || !data) {
    return (
      <section className={styles.card} aria-label="Analyse de regret">
        <div className={styles.header}>
          <span className={styles.title}>Analyse de regret</span>
        </div>
        <p className={styles.errorState}>Impossible de charger l'analyse de regret.</p>
      </section>
    )
  }

  const getSeverityBadgeClass = (severity: string) => {
    switch (severity) {
      case 'high':
        return styles.severityHigh
      case 'medium':
        return styles.severityMedium
      case 'low':
        return styles.severityLow
      default:
        return styles.severityNone
    }
  }

  const getScoreClass = (score: number) => {
    if (score <= 30) return styles.scoreGood
    if (score <= 60) return styles.scoreWarning
    return styles.scoreBad
  }

  return (
    <section className={styles.card} aria-label="Analyse de regret">
      <div className={styles.header}>
        <span className={styles.title}>Analyse de regret</span>
      </div>

      {/* Score Gauge */}
      <div className={styles.gaugeContainer}>
        <div className={styles.gauge}>
          {/* Background track */}
          <svg
            className={styles.gaugeTrack}
            viewBox="0 0 200 100"
            preserveAspectRatio="xMidYMid meet"
          >
            <path
              d="M 20 80 A 60 60 0 0 1 180 80"
              fill="none"
              stroke="var(--border)"
              strokeWidth="8"
              strokeLinecap="round"
            />
          </svg>

          {/* Colored arc */}
          <svg
            className={`${styles.gaugeFill} ${getScoreClass(data.regretScore)}`}
            viewBox="0 0 200 100"
            preserveAspectRatio="xMidYMid meet"
            style={{
              '--arc-degrees': `${arcDegrees}deg`,
            } as React.CSSProperties}
          >
            <path
              d="M 20 80 A 60 60 0 0 1 180 80"
              fill="none"
              strokeWidth="8"
              strokeLinecap="round"
            />
          </svg>

          {/* Score text in center */}
          <div className={styles.scoreText}>
            <span className={styles.scoreValue}>{data.regretScore}</span>
            <span className={styles.scoreLabel}>/100</span>
          </div>
        </div>
      </div>

      {/* Summary */}
      <p className={styles.summary}>{data.summary}</p>

      {/* Total Spent */}
      <div className={styles.totalSpent}>
        <span className={styles.label}>Total dépensé</span>
        <span className={styles.amount}>
          {new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(
            data.totalSpent
          )}
        </span>
      </div>

      {/* Signals List */}
      {data.signals.length > 0 && (
        <div className={styles.signals}>
          <h3 className={styles.signalsTitle}>Signaux détectés</h3>
          {data.signals.map((signal, idx) => (
            <div key={idx} className={styles.signal}>
              <div className={styles.signalHeader}>
                <span className={styles.signalType}>
                  {signal.type === 'overpaid' && 'Surpayé'}
                  {signal.type === 'impulse' && 'Achat impulsif'}
                  {signal.type === 'timing' && 'Mauvais timing'}
                </span>
                <span className={`${styles.severityBadge} ${getSeverityBadgeClass(signal.severity)}`}>
                  {signal.severity === 'high' && 'Élevé'}
                  {signal.severity === 'medium' && 'Moyen'}
                  {signal.severity === 'low' && 'Faible'}
                </span>
              </div>
              <p className={styles.signalDescription}>{signal.description}</p>
              {signal.potentialSaving > 0 && (
                <span className={styles.savings}>
                  Économie potentielle: {new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(signal.potentialSaving)}
                </span>
              )}
            </div>
          ))}
        </div>
      )}

      {data.signals.length === 0 && (
        <div className={styles.noSignals}>
          <p>Aucun signal de regret détecté. Excellentes décisions d'achat!</p>
        </div>
      )}
    </section>
  )
}

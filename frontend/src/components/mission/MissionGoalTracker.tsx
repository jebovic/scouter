import { useMemo } from 'react'
import { computeMissionProgress } from '../../utils/missionProgress'
import styles from './MissionGoalTracker.module.css'

interface MissionGoalTrackerProps {
  mission: { budget?: number | null; currency?: string; createdAt: string }
  items: Array<{ currentPrice: number; targetPrice?: number | null }>
  targetDays?: number
}

export function MissionGoalTracker({ mission, items, targetDays }: MissionGoalTrackerProps) {
  const progress = useMemo(
    () => computeMissionProgress(mission, items, targetDays),
    [mission, items, targetDays]
  )

  // Status label in French
  const statusLabel = {
    ahead: 'AVANCE',
    on_track: 'EN COURS',
    at_risk: 'ATTENTION',
    overdue: 'EN RETARD',
  }[progress.overallStatus]

  // Status color class
  const statusClass = {
    ahead: styles.statusAhead,
    on_track: styles.statusOnTrack,
    at_risk: styles.statusAtRisk,
    overdue: styles.statusOverdue,
  }[progress.overallStatus]

  // Budget status colors
  const budgetStatusClass = {
    on_track: styles.budgetOnTrack,
    under_budget: styles.budgetUnderBudget,
    over_budget: styles.budgetOverBudget,
  }[progress.budgetStatus]

  // Urgency badge
  const urgencyLabel = {
    low: 'Normal',
    medium: 'Modérée',
    high: 'Élevée',
  }[progress.urgency]

  const urgencyClass = {
    low: styles.urgencyLow,
    medium: styles.urgencyMedium,
    high: styles.urgencyHigh,
  }[progress.urgency]

  return (
    <div className={styles.container}>
      {/* Overall status banner */}
      <div className={`${styles.statusBanner} ${statusClass}`}>
        <div className={styles.statusIcon}>
          {progress.overallStatus === 'ahead' && '🚀'}
          {progress.overallStatus === 'on_track' && '📍'}
          {progress.overallStatus === 'at_risk' && '⚠️'}
          {progress.overallStatus === 'overdue' && '🚨'}
        </div>
        <div className={styles.statusContent}>
          <h3 className={styles.statusTitle}>{statusLabel}</h3>
          <p className={styles.statusDesc}>
            {progress.overallStatus === 'ahead' && 'Vous avez de l\'avance sur votre planning'}
            {progress.overallStatus === 'on_track' && 'Tout progresse comme prévu'}
            {progress.overallStatus === 'at_risk' && 'Vous êtes à la traîne, dépêchez-vous'}
            {progress.overallStatus === 'overdue' && 'Délai dépassé, achetez maintenant'}
          </p>
        </div>
      </div>

      {/* Progress cards grid */}
      <div className={styles.cardGrid}>
        {/* Budget card */}
        <div className={`${styles.card} ${styles.budgetCard}`}>
          <div className={styles.cardHeader}>
            <h4 className={styles.cardTitle}>Budget</h4>
            <span className={styles.budgetAmount}>
              {progress.budgetSpent.toFixed(0)} / {progress.budgetTotal.toFixed(0)}
            </span>
          </div>

          {/* Circular progress ring */}
          <div className={styles.ringContainer}>
            <svg className={styles.ring} viewBox="0 0 120 120">
              <circle
                className={styles.ringBg}
                cx="60"
                cy="60"
                r="54"
              />
              <circle
                className={`${styles.ringProgress} ${budgetStatusClass}`}
                cx="60"
                cy="60"
                r="54"
                style={{
                  strokeDasharray: `${(progress.budgetPercent / 100) * (2 * Math.PI * 54)} ${2 * Math.PI * 54}`,
                }}
              />
            </svg>
            <div className={styles.ringLabel}>
              <div className={styles.percent}>{progress.budgetPercent.toFixed(0)}%</div>
            </div>
          </div>

          <div className={styles.budgetStatus}>
            {progress.budgetStatus === 'over_budget' && '❌ Dépassement'}
            {progress.budgetStatus === 'under_budget' && '✅ Économies'}
            {progress.budgetStatus === 'on_track' && '📊 En cours'}
          </div>
        </div>

        {/* Readiness card */}
        <div className={`${styles.card} ${styles.readinessCard}`}>
          <div className={styles.cardHeader}>
            <h4 className={styles.cardTitle}>Préparation</h4>
            <span className={styles.itemCount}>
              {progress.readyItems} / {progress.totalItems} articles prêts
            </span>
          </div>

          {/* Progress bar */}
          <div className={styles.progressBar}>
            <div
              className={styles.progressFill}
              style={{
                width: `${progress.readinessPercent}%`,
              }}
            />
          </div>

          <div className={styles.readinessPercent}>
            {progress.readinessPercent.toFixed(0)}% prêts
          </div>
        </div>

        {/* Timeline card */}
        <div className={`${styles.card} ${styles.timelineCard}`}>
          <div className={styles.cardHeader}>
            <h4 className={styles.cardTitle}>Temps</h4>
            <span className={`${styles.urgencyBadge} ${urgencyClass}`}>
              {urgencyLabel}
            </span>
          </div>

          {/* Progress bar */}
          <div className={styles.progressBar}>
            <div
              className={styles.progressFill}
              style={{
                width: `${Math.min(100, progress.timePercent)}%`,
              }}
            />
          </div>

          <div className={styles.timelineLabel}>
            <div>J+{progress.daysElapsed}</div>
            <div className={styles.timelineDesc}>
              {progress.daysElapsed} / {progress.daysTotal} jours
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

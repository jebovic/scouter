import { useState } from 'react'
import { useBudgetAnalysis } from '../../hooks/useBudgetAnalysis'
import type { BudgetRecommendation } from '../../api/budgetrec'
import styles from './BudgetRecommendations.module.css'

interface Props {
  missionId: string
}

// Returns a data-health attribute value driving gauge stroke colour.
function healthColor(score: number): 'green' | 'cyan' | 'gold' | 'orange' | 'coral' {
  if (score >= 80) return 'green'
  if (score >= 60) return 'cyan'
  if (score >= 40) return 'gold'
  if (score >= 20) return 'orange'
  return 'coral'
}

const CIRCUMFERENCE = 2 * Math.PI * 28 // r=28

function Gauge({ score }: { score: number }) {
  const offset = CIRCUMFERENCE * (1 - score / 100)
  const color = healthColor(score)
  return (
    <div className={styles.gaugeWrap}>
      <svg className={styles.gaugeSvg} viewBox="0 0 72 72" aria-hidden="true">
        <circle className={styles.gaugeTrack} cx="36" cy="36" r="28" />
        <circle
          className={styles.gaugeFill}
          cx="36"
          cy="36"
          r="28"
          strokeDasharray={CIRCUMFERENCE}
          strokeDashoffset={offset}
          data-health={color}
        />
      </svg>
      <div className={styles.gaugeLabel}>{score}</div>
    </div>
  )
}

function RecCard({ rec }: { rec: BudgetRecommendation }) {
  const [expanded, setExpanded] = useState(false)
  return (
    <div
      className={styles.recItem}
      onClick={() => setExpanded((v) => !v)}
      role="button"
      tabIndex={0}
      aria-expanded={expanded}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') setExpanded((v) => !v) }}
    >
      <div className={styles.recHeader}>
        <span className={styles.recCategory}>{rec.category || 'Général'}</span>
        <span className={styles.recPriority} data-priority={rec.priority}>
          {rec.priority}
        </span>
      </div>
      <p className={styles.recReason}>{rec.reason}</p>
      {expanded && (
        <div className={styles.recDetail}>
          <div className={styles.recDetailItem}>
            Budget actuel&nbsp;: <span>{rec.currentBudget.toFixed(2)} €</span>
          </div>
          <div className={styles.recDetailItem}>
            Budget recommandé&nbsp;: <span>{rec.recommendedBudget.toFixed(2)} €</span>
          </div>
          <div className={styles.recDetailItem}>
            Économie potentielle&nbsp;: <span>{rec.saveAmount.toFixed(2)} €</span>
          </div>
        </div>
      )}
    </div>
  )
}

export function BudgetRecommendations({ missionId }: Props) {
  const { data, isLoading, isError } = useBudgetAnalysis(missionId)

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Recommandations budgétaires">
        <div className={styles.header}>
          <span className={styles.title}>Recommandations budgétaires</span>
        </div>
        <div className={styles.gaugeRow} data-testid="budget-loading">
          <div className={`${styles.skeleton} ${styles.skeletonCircle}`} />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div className={`${styles.skeleton} ${styles.skeletonLine}`} style={{ width: '40%' }} />
            <div className={`${styles.skeleton} ${styles.skeletonLine}`} style={{ width: '60%' }} />
          </div>
        </div>
        {[1, 2].map((i) => (
          <div key={i} className={`${styles.skeleton} ${styles.skeletonLine}`} />
        ))}
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="Recommandations budgétaires">
        <div className={styles.header}>
          <span className={styles.title}>Recommandations budgétaires</span>
        </div>
        <p className={styles.errorState}>Impossible de charger l'analyse budgétaire.</p>
      </section>
    )
  }

  return (
    <section className={styles.card} aria-label="Recommandations budgétaires">
      <div className={styles.header}>
        <span className={styles.title}>Recommandations budgétaires</span>
      </div>

      {/* Health score gauge + totals */}
      <div className={styles.gaugeRow}>
        <Gauge score={data.healthScore} />
        <div className={styles.gaugeMeta}>
          <span className={styles.scoreValue}>{data.healthScore}<small>/100</small></span>
          <span className={styles.scoreLabel}>Score budgétaire</span>
          <div className={styles.totalsRow}>
            <div className={styles.totalItem}>
              <span className={styles.totalKey}>Budget</span>
              <span className={styles.totalVal}>{data.totalBudget.toFixed(2)} €</span>
            </div>
            <div className={styles.totalItem}>
              <span className={styles.totalKey}>Dépensé</span>
              <span className={styles.totalVal}>{data.totalSpent.toFixed(2)} €</span>
            </div>
          </div>
        </div>
      </div>

      {/* Potential savings summary */}
      {data.potentialSavings > 0 && (
        <div className={styles.savingsRow}>
          <span className={styles.savingsLabel}>Économies potentielles identifiées</span>
          <span className={styles.savingsAmount}>+{data.potentialSavings.toFixed(2)} €</span>
        </div>
      )}

      {/* Recommendation cards */}
      {data.recommendations.length > 0 ? (
        <div className={styles.recList}>
          {data.recommendations.map((rec, i) => (
            <RecCard key={`${rec.category}-${i}`} rec={rec} />
          ))}
        </div>
      ) : (
        <p className={styles.emptyState}>Budget équilibré — aucune recommandation pour le moment.</p>
      )}
    </section>
  )
}

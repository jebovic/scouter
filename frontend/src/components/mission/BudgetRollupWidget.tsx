import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useBudgetRollup } from '../../hooks/useBudgetRollup'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import styles from './BudgetRollupWidget.module.css'

const BUDGET_REFERENCE_CAP = 10_000

const CATEGORY_COLORS = [
  'var(--accent)',
  'var(--gold)',
  'var(--green)',
  'var(--coral)',
  'var(--purple, #a78bfa)',
  'var(--cyan, #67e8f9)',
]

const PHASE_LABELS: Record<string, string> = {
  researching: 'Recherche',
  comparing: 'Comparaison',
  buying: 'Décidé',
  done: 'Acheté',
}

const PHASE_COLORS: Record<string, string> = {
  researching: 'var(--text-dim)',
  comparing: 'var(--gold)',
  buying: 'var(--purple, #a78bfa)',
  done: 'var(--green)',
}

function getProgressColorClass(pct: number): string {
  if (pct < 60) return styles.progressFillGreen
  if (pct < 90) return styles.progressFillGold
  return styles.progressFillCoral
}

export function BudgetRollupWidget() {
  const { rollup } = useBudgetRollup()
  const { fmt } = useFormatCurrency()
  const [expanded, setExpanded] = useState(true)

  if (rollup.entries.length === 0) return null

  const pct = Math.min(100, (rollup.totalBudget / BUDGET_REFERENCE_CAP) * 100)
  const progressColorClass = getProgressColorClass(pct)

  const categories = Object.entries(rollup.byCategory)
  const maxCategoryBudget = Math.max(...categories.map(([, v]) => v), 1)

  const phases = Object.entries(rollup.byPhase).sort(([a], [b]) => {
    const order = ['researching', 'comparing', 'buying', 'done']
    return order.indexOf(a) - order.indexOf(b)
  })

  return (
    <div className={styles.widget}>
      <button
        className={styles.header}
        onClick={() => setExpanded((prev) => !prev)}
        aria-expanded={expanded}
        aria-label="Budget Total Actif"
      >
        <div className={styles.headerLeft}>
          <span className={styles.headerLabel}>Budget Total Actif</span>
          <span className={styles.headerTotal}>{fmt(rollup.totalBudget)}</span>
        </div>
        <div className={styles.headerRight}>
          <span className={styles.missionCount}>
            {rollup.activeMissions} mission{rollup.activeMissions !== 1 ? 's' : ''} active{rollup.activeMissions !== 1 ? 's' : ''}
          </span>
          <span className={`${styles.chevron}${expanded ? ` ${styles.chevronOpen}` : ''}`}>
            ▼
          </span>
        </div>
      </button>

      {expanded && (
        <div className={styles.body}>
          {/* Progress bar */}
          <div className={styles.progressSection}>
            <div className={styles.progressLabel}>
              <span>Exposition totale</span>
              <span>{fmt(BUDGET_REFERENCE_CAP)} de référence</span>
            </div>
            <div className={styles.progressTrack}>
              <div
                className={`${styles.progressFill} ${progressColorClass}`}
                style={{ width: `${pct}%` } as React.CSSProperties}
                data-testid="budget-progress-fill"
              />
            </div>
          </div>

          {/* Category breakdown */}
          <div className={styles.section}>
            <div className={styles.sectionTitle}>Par catégorie</div>
            <div className={styles.categoryBars}>
              {categories.map(([cat, amount], idx) => {
                const barPct = (amount / maxCategoryBudget) * 100
                const color = CATEGORY_COLORS[idx % CATEGORY_COLORS.length]
                return (
                  <div key={cat} className={styles.categoryRow}>
                    <span className={styles.categoryName}>{cat}</span>
                    <div className={styles.categoryBarTrack}>
                      <div
                        className={styles.categoryBarFill}
                        style={{ width: `${barPct}%`, background: color } as React.CSSProperties}
                      />
                    </div>
                    <span className={styles.categoryAmount}>{fmt(amount)}</span>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Phase funnel */}
          <div className={styles.section}>
            <div className={styles.sectionTitle}>Par phase</div>
            <div className={styles.phaseFunnel}>
              {phases.map(([phase, amount]) => (
                <div key={phase} className={styles.phaseBadge}>
                  <span
                    className={styles.phaseLabel}
                    style={{ color: PHASE_COLORS[phase] ?? 'var(--text-dim)' } as React.CSSProperties}
                  >
                    {PHASE_LABELS[phase] ?? phase}
                  </span>
                  <span className={styles.phaseAmount}>{fmt(amount)}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Mission list */}
          <div className={styles.section}>
            <div className={styles.sectionTitle}>Missions</div>
            <div className={styles.missionTable}>
              {rollup.entries.map((entry) => (
                <Link
                  key={entry.missionId}
                  to={`/missions/${entry.missionSlug}`}
                  className={styles.missionRow}
                >
                  <span className={styles.missionName}>{entry.missionName}</span>
                  <span className={styles.missionCategory}>{entry.category}</span>
                  <span
                    className={styles.missionPhase}
                    style={{ color: PHASE_COLORS[entry.phase] ?? 'var(--text-dim)' } as React.CSSProperties}
                  >
                    {PHASE_LABELS[entry.phase] ?? entry.phase}
                  </span>
                  <span className={styles.missionBudget}>{fmt(entry.budget)}</span>
                </Link>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

import { useEnvelopes } from '../../hooks/useEnvelopes'
import { useBudgetVariance } from '../../hooks/useBudgetVariance'
import { EmptyState } from './EmptyState'
import styles from './BudgetHeatmap.module.css'

interface BudgetHeatmapProps {
  monthsBack?: number
}

function getHeatmapColor(variancePct: number): string {
  if (variancePct === Infinity) {
    return 'var(--surface-raised)' // No budget set
  }
  if (variancePct < -20) {
    return 'var(--green)' // Well under budget
  }
  if (variancePct < 0) {
    return 'var(--green)' // Under budget
  }
  if (variancePct < 80) {
    return 'var(--green)' // Under budget
  }
  if (variancePct < 100) {
    return 'var(--gold)' // Attention zone
  }
  // Over budget
  return 'var(--coral)'
}

function getHeatmapOpacity(variancePct: number): number {
  if (variancePct === Infinity) {
    return 0.3
  }
  if (Math.abs(variancePct) < 20) {
    return 0.3
  }
  if (Math.abs(variancePct) < 50) {
    return 0.6
  }
  return 0.9
}

export function BudgetHeatmap({ monthsBack = 3 }: BudgetHeatmapProps) {
  const { envelopes } = useEnvelopes()
  const { months, categories, cells } = useBudgetVariance(envelopes, monthsBack)

  if (categories.length === 0) {
    return (
      <EmptyState
        icon="📊"
        title="Heatmap Budgétaire"
        description="Créez des enveloppes pour voir l'analyse de variance budgétaire"
      />
    )
  }

  // Create a map of cells for easy lookup: month + category → cell
  const cellMap = new Map(
    cells.map((cell) => [`${cell.month}|${cell.category}`, cell]),
  )

  const formatVariance = (pct: number) => {
    if (pct === Infinity) return 'N/A'
    if (isNaN(pct)) return 'N/A'
    return `${pct > 0 ? '+' : ''}${pct.toFixed(0)}%`
  }

  return (
    <div className={styles.container}>
      <div className={styles.heatmapWrapper}>
        {/* Header row with month labels */}
        <div className={styles.grid}>
          <div className={styles.headerCell} style={{ gridColumn: 1, gridRow: 1 }} />
          {months.map((month, idx) => {
            const cell = cellMap.get(`${month}|${categories[0]}`)
            return (
              <div
                key={month}
                className={styles.headerCell}
                style={{ gridColumn: idx + 2, gridRow: 1 }}
              >
                {cell?.monthLabel}
              </div>
            )
          })}

          {/* Data rows */}
          {categories.map((category, rowIdx) => (
            <div key={category}>
              <div
                className={styles.categoryLabel}
                style={{ gridColumn: 1, gridRow: rowIdx + 2 }}
              >
                {category}
              </div>
              {months.map((month, colIdx) => {
                const cell = cellMap.get(`${month}|${category}`)
                if (!cell) return null

                const color = getHeatmapColor(cell.variancePct)
                const opacity = getHeatmapOpacity(cell.variancePct)

                return (
                  <div
                    key={`${month}|${category}`}
                    className={styles.cell}
                    style={{
                      gridColumn: colIdx + 2,
                      gridRow: rowIdx + 2,
                      '--cell-color': color,
                      '--cell-opacity': opacity,
                    } as React.CSSProperties}
                    title={`${category} - ${cell.monthLabel}: ${formatVariance(cell.variancePct)}`}
                  >
                    {formatVariance(cell.variancePct)}
                  </div>
                )
              })}
            </div>
          ))}
        </div>
      </div>

      {/* Legend */}
      <div className={styles.legend}>
        <div className={styles.legendRow}>
          <div className={styles.legendItem}>
            <div className={styles.legendColor} style={{ backgroundColor: 'var(--green)' }} />
            <span>En dessous du budget</span>
          </div>
          <div className={styles.legendItem}>
            <div className={styles.legendColor} style={{ backgroundColor: 'var(--gold)' }} />
            <span>Attention</span>
          </div>
          <div className={styles.legendItem}>
            <div className={styles.legendColor} style={{ backgroundColor: 'var(--coral)' }} />
            <span>Dépassé</span>
          </div>
        </div>
      </div>
    </div>
  )
}

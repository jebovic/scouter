import styles from './BudgetHeatmap.module.css'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { BudgetHeatmap as BudgetHeatmapData, HeatmapCell } from '../../api/budgetheatmap'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Map intensity 0-1 → HSL lightness: 0=light (no spend), 1=dark (max spend). */
function intensityToColor(intensity: number): string {
  // light blue (90% L) → saturated dark blue (20% L)
  const lightness = Math.round(90 - intensity * 70)
  return `hsl(220, 70%, ${lightness}%)`
}

function buildCellMap(cells: HeatmapCell[]): Map<string, HeatmapCell> {
  const m = new Map<string, HeatmapCell>()
  for (const c of cells) {
    m.set(`${c.category}::${c.month}`, c)
  }
  return m
}

// ---------------------------------------------------------------------------
// Legend
// ---------------------------------------------------------------------------

const LEGEND_STEPS = [0, 0.2, 0.4, 0.6, 0.8, 1]

function Legend() {
  return (
    <div className={styles.legend}>
      <span>Faible</span>
      <div className={styles.legendScale}>
        {LEGEND_STEPS.map((v) => (
          <div
            key={v}
            className={styles.legendSwatch}
            style={{ background: intensityToColor(v) }}
          />
        ))}
      </div>
      <span>Élevé</span>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

interface Props {
  data: BudgetHeatmapData
}

export default function BudgetHeatmap({ data }: Props) {
  const { fmt } = useFormatCurrency()

  if (data.cells.length === 0) {
    return (
      <div className={styles.empty}>
        Aucune donnée d'historique de prix disponible pour les 6 derniers mois.
      </div>
    )
  }

  const cellMap = buildCellMap(data.cells)

  return (
    <div className={styles.wrap}>
      <table className={styles.table}>
        <thead>
          <tr>
            {/* Empty top-left corner */}
            <th className={styles.headerCell} />
            {data.months.map((m) => (
              <th key={m} className={styles.headerCell}>
                {m}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.categories.map((cat) => (
            <tr key={cat}>
              <td className={styles.rowLabel}>{cat || 'Sans catégorie'}</td>
              {data.months.map((month) => {
                const cell = cellMap.get(`${cat}::${month}`)
                if (!cell) {
                  return <td key={month} className={`${styles.cell} ${styles.emptyCell}`} />
                }
                const tip = `${cat || 'Sans catégorie'} / ${month}: ${fmt(cell.total)} (${cell.count} entrée${cell.count > 1 ? 's' : ''})`
                return (
                  <td
                    key={month}
                    className={styles.cell}
                    style={
                      { background: intensityToColor(cell.intensity) } as React.CSSProperties
                    }
                    data-tip={tip}
                    title={tip}
                  />
                )
              })}
            </tr>
          ))}
        </tbody>
      </table>
      <Legend />
    </div>
  )
}

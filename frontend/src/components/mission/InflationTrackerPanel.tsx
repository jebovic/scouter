import { useInflationImpact } from '../../hooks/useInflationImpact'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { InflationItem } from '../../api/inflationtracker'
import styles from './InflationTrackerPanel.module.css'

interface Props {
  missionId: string | undefined
}

function fmtPct(rate: number): string {
  const sign = rate >= 0 ? '+' : ''
  return `${sign}${(rate * 100).toFixed(1)}%`
}

function rateClass(rate: number, s: typeof styles): string {
  if (rate < 0) return s.rateGreen
  if (rate <= 0.03) return s.rateYellow
  return s.rateRed
}

function ItemRow({ item, s, fmt }: { item: InflationItem; s: typeof styles; fmt: (n: number) => string }) {
  return (
    <tr>
      <td>{item.name}</td>
      <td className={rateClass(item.annualInflationRate, s)}>
        {fmtPct(item.annualInflationRate)}
      </td>
      <td className={s.mono}>{fmt(item.inflationPer30Days)}</td>
      <td className={s.mono}>{fmt(item.waitCost3Months)}</td>
      <td className={s.rec}>{item.recommendation}</td>
    </tr>
  )
}

export function InflationTrackerPanel({ missionId }: Props) {
  const { data, isLoading, isError } = useInflationImpact(missionId)
  const { fmt } = useFormatCurrency()

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Impact de l'inflation">
        <div className={styles.header}>
          <span className={styles.title}>Impact Inflation</span>
        </div>
        {[1, 2, 3].map((i) => (
          <div key={i} className={`${styles.skeleton} ${styles.skeletonRow}`} />
        ))}
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="Impact de l'inflation">
        <div className={styles.header}>
          <span className={styles.title}>Impact Inflation</span>
        </div>
        <p className={styles.errorState}>Impossible de calculer l'impact inflation.</p>
      </section>
    )
  }

  const isPositiveCost = data.totalWaitCost3Months > 0

  return (
    <section className={styles.card} aria-label="Impact de l'inflation">
      <div className={styles.header}>
        <span className={styles.title}>Impact Inflation (IPC Français)</span>
      </div>

      {/* Summary banner */}
      <div className={styles.summary}>
        <span className={styles.summaryLabel}>Coût d'attente 3 mois&nbsp;:</span>
        <span
          className={styles.summaryValue}
          data-positive={isPositiveCost ? 'true' : 'false'}
        >
          {isPositiveCost ? '+' : ''}{fmt(data.totalWaitCost3Months)}
        </span>
        <span className={styles.summaryAvg}>
          Inflation moy.&nbsp;: {fmtPct(data.avgInflationRate)}
        </span>
      </div>

      {/* Items table */}
      {data.items.length === 0 ? (
        <p className={styles.emptyState}>Aucun article dans cette mission.</p>
      ) : (
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>Article</th>
                <th>Taux annuel</th>
                <th>Coût / mois</th>
                <th>Coût 3 mois</th>
                <th>Recommandation</th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((item) => (
                <ItemRow key={item.itemId} item={item} s={styles} fmt={fmt} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}

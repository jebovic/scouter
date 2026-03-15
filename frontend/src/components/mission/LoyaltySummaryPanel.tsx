import { useLoyaltySummary } from '../../hooks/useLoyaltySummary'
import type { MerchantSummary } from '../../api/loyaltytracker'
import styles from './LoyaltySummaryPanel.module.css'

function formatEur(value: number): string {
  return new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency: 'EUR',
    maximumFractionDigits: 2,
  }).format(value)
}

function MerchantRow({ m }: { m: MerchantSummary }) {
  const hasCashback = m.estimatedCashback > 0
  return (
    <tr className={hasCashback ? styles.rowHighlight : undefined}>
      <td>
        <div className={styles.merchantName}>{m.merchant}</div>
        <div className={styles.programName}>{m.program}</div>
        {m.tier && <div className={styles.tierBadge}>{m.tier}</div>}
      </td>
      <td>{formatEur(m.total)}</td>
      <td className={hasCashback ? styles.cashbackCell : styles.zeroCell}>
        {hasCashback ? formatEur(m.estimatedCashback) : '—'}
      </td>
      <td className={hasCashback ? undefined : styles.zeroCell}>
        {hasCashback ? m.estimatedPoints.toLocaleString('fr-FR') : '—'}
      </td>
    </tr>
  )
}

interface Props {
  missionId: string
}

export function LoyaltySummaryPanel({ missionId }: Props) {
  const { summary, isLoading } = useLoyaltySummary(missionId)

  if (isLoading) {
    return (
      <div className={styles.panel}>
        <div className={styles.header}>
          <span className={styles.headerTitle}>FIDÉLITÉ MARCHANDS</span>
        </div>
        <div className={styles.skeletonBody}>
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className={styles.skeletonRow} />
          ))}
        </div>
      </div>
    )
  }

  if (!summary || summary.merchants.length === 0) {
    return (
      <div className={styles.panel}>
        <div className={styles.header}>
          <span className={styles.headerTitle}>FIDÉLITÉ MARCHANDS</span>
        </div>
        <div className={styles.empty}>
          <span className={styles.emptyIcon}>🏪</span>
          <span>Aucun achat avec marchand enregistré</span>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <span className={styles.headerTitle}>FIDÉLITÉ MARCHANDS</span>
      </div>

      <div className={styles.cashbackHero}>
        <span className={styles.cashbackLabel}>Cashback estimé total</span>
        <span className={styles.cashbackAmount}>
          {formatEur(summary.totalEstimatedCashback)}
        </span>
      </div>

      <div className={styles.tableWrapper}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>Marchand / Programme</th>
              <th>Dépenses</th>
              <th>Cashback</th>
              <th>Points</th>
            </tr>
          </thead>
          <tbody>
            {summary.merchants.map((m) => (
              <MerchantRow key={m.merchant} m={m} />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

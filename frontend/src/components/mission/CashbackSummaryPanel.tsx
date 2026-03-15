import { useState } from 'react'
import { useCashbackTracker } from '../../hooks/useCashbackTracker'
import styles from './CashbackSummaryPanel.module.css'

interface Props {
  missionId: string
  currency?: string
}

function formatCurrency(amount: number, currency = 'EUR'): string {
  return new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency,
    maximumFractionDigits: 2,
  }).format(amount)
}

export function CashbackSummaryPanel({ missionId, currency = 'EUR' }: Props) {
  const { data, isLoading, isError } = useCashbackTracker(missionId)
  const [isExpanded, setIsExpanded] = useState(true)

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Estimé de cashback">
        <div className={styles.header}>
          <span className={styles.title}>💰 Cashback Estimé</span>
        </div>
        <div className={`${styles.skeleton} ${styles.skeletonLine}`} />
        <div className={styles.skeletonGrid}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={`${styles.skeleton} ${styles.skeletonCell}`} />
          ))}
        </div>
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="Estimé de cashback">
        <div className={styles.header}>
          <span className={styles.title}>💰 Cashback Estimé</span>
        </div>
        <p className={styles.errorState}>Impossible de charger l'estimé de cashback.</p>
      </section>
    )
  }

  return (
    <section className={styles.card} aria-label="Estimé de cashback">
      <div className={styles.header}>
        <button
          className={styles.toggleBtn}
          onClick={() => setIsExpanded(!isExpanded)}
          aria-expanded={isExpanded}
        >
          <span className={styles.title}>💰 Cashback Estimé</span>
          <span className={`${styles.toggleIcon} ${isExpanded ? styles.expanded : ''}`}>▼</span>
        </button>
      </div>

      {isExpanded && (
        <>
          <div className={styles.totalAmount}>
            <span className={styles.label}>Estimé total</span>
            <span className={styles.amount} style={{ '--success-color': 'var(--success)' } as React.CSSProperties}>
              {formatCurrency(data.totalEstimated, currency)}
            </span>
          </div>

          <div className={styles.bestPlatform}>
            <span className={styles.platformLabel}>Meilleure plateforme:</span>
            <span className={styles.platformBadge}>{data.bestPlatform}</span>
          </div>

          <p className={styles.tip}>{data.tip}</p>

          {data.merchantCashback.length > 0 && (
            <div className={styles.tableContainer}>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th className={styles.thMerchant}>Marchand</th>
                    <th className={styles.thPlatform}>Plateforme</th>
                    <th className={styles.thPct}>% Cashback</th>
                    <th className={styles.thAmount}>Estimé</th>
                    <th className={styles.thCount}>Articles</th>
                  </tr>
                </thead>
                <tbody>
                  {data.merchantCashback.map((m) => (
                    <tr key={m.merchant} className={styles.row}>
                      <td className={styles.tdMerchant}>{m.merchant}</td>
                      <td className={styles.tdPlatform}>{m.platform}</td>
                      <td className={styles.tdPct}>{m.cashbackPct.toFixed(1)}%</td>
                      <td className={styles.tdAmount}>{formatCurrency(m.estimatedAmount, currency)}</td>
                      <td className={styles.tdCount}>{m.itemCount}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </section>
  )
}

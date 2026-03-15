import { useStats } from '../hooks/usePurchase'
import { StarRating } from '../components/scouter/StarRating'
import styles from './StatsPage.module.css'

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(amount)
}

export default function StatsPage() {
  const { data: stats, isLoading } = useStats()

  if (isLoading) {
    return (
      <main className={styles.page}>
        <h1 className={styles.heading}>Spending Stats</h1>
        <div className={styles.loading}>Loading stats...</div>
      </main>
    )
  }

  if (!stats || stats.purchaseCount === 0) {
    return (
      <main className={styles.page}>
        <h1 className={styles.heading}>Spending Stats</h1>
        <div className={styles.empty}>
          <p className={styles.emptyIcon}>📊</p>
          <p className={styles.emptyTitle}>No purchases recorded yet</p>
          <p className={styles.emptyDesc}>Record purchases on completed missions to see your spending stats.</p>
        </div>
      </main>
    )
  }

  const maxCategorySpend = Math.max(...stats.categoryBreakdown.map((c) => c.totalSpent), 1)

  return (
    <main className={styles.page}>
      <h1 className={styles.heading}>Spending Stats</h1>

      <div className={styles.summaryGrid}>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Total Spent</span>
          <span className={styles.statValue}>{formatCurrency(stats.totalSpent)}</span>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Total Budget</span>
          <span className={styles.statValue}>{formatCurrency(stats.totalBudget)}</span>
        </div>
        <div className={`${styles.statCard} ${stats.savings >= 0 ? styles.positive : styles.negative}`}>
          <span className={styles.statLabel}>{stats.savings >= 0 ? 'Saved' : 'Over Budget'}</span>
          <span className={styles.statValue}>{formatCurrency(Math.abs(stats.savings))}</span>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Purchases</span>
          <span className={styles.statValue}>{stats.purchaseCount}</span>
        </div>
      </div>

      {stats.categoryBreakdown.length > 0 && (
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>By Category</h2>
          <div className={styles.categories}>
            {stats.categoryBreakdown.map((cat) => (
              <div key={cat.category} className={styles.categoryRow}>
                <div className={styles.categoryMeta}>
                  <span className={styles.categoryName}>{cat.category}</span>
                  <span className={styles.categoryCount}>{cat.count} purchase{cat.count !== 1 ? 's' : ''}</span>
                  <span className={styles.categorySpend}>{formatCurrency(cat.totalSpent)}</span>
                  {cat.avgSatisfaction && (
                    <StarRating value={Math.round(cat.avgSatisfaction)} readOnly size={14} />
                  )}
                </div>
                <div className={styles.barWrap}>
                  <div
                    className={styles.bar}
                    style={{ '--pct': `${(cat.totalSpent / maxCategorySpend) * 100}%` } as React.CSSProperties}
                  />
                </div>
              </div>
            ))}
          </div>
        </section>
      )}
    </main>
  )
}

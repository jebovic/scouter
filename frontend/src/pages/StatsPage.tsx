import { useTranslation } from 'react-i18next'
import { useStats, useMonthlyStats } from '../hooks/usePurchase'
import { useSettings } from '../hooks/useSettings'
import { StarRating } from '../components/scouter/StarRating'
import { SpendTrendChart } from '../components/charts/SpendTrendChart'
import { CategoryDonutChart } from '../components/charts/CategoryDonutChart'
import { BudgetVsActualChart } from '../components/charts/BudgetVsActualChart'
import { formatCurrencyLocale } from '../utils/format'
import styles from './StatsPage.module.css'

export default function StatsPage() {
  const { t } = useTranslation()
  const { data: settings } = useSettings()
  const locale = settings?.locale ?? 'fr-FR'
  const currency = settings?.currency ?? 'EUR'

  const { data: stats, isLoading } = useStats()
  const { data: monthlyStats } = useMonthlyStats()

  const fmt = (amount: number) => formatCurrencyLocale(amount, locale, currency)

  if (isLoading) {
    return (
      <main className={styles.page}>
        <h1 className={styles.heading}>{t('stats.title')}</h1>
        <div className={styles.loading}>{t('stats.loading')}</div>
      </main>
    )
  }

  if (!stats || stats.purchaseCount === 0) {
    return (
      <main className={styles.page}>
        <h1 className={styles.heading}>{t('stats.title')}</h1>
        <div className={styles.empty}>
          <p className={styles.emptyIcon}>📊</p>
          <p className={styles.emptyTitle}>{t('stats.noData')}</p>
          <p className={styles.emptyDesc}>{t('stats.noDataDesc')}</p>
        </div>
      </main>
    )
  }

  const maxCategorySpend = Math.max(...stats.categoryBreakdown.map((c) => c.totalSpent), 1)

  return (
    <main className={styles.page}>
      <h1 className={styles.heading}>{t('stats.title')}</h1>

      <div className={styles.summaryGrid}>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>{t('stats.totalSpent')}</span>
          <span className={styles.statValue}>{fmt(stats.totalSpent)}</span>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>{t('stats.totalBudget')}</span>
          <span className={styles.statValue}>{fmt(stats.totalBudget)}</span>
        </div>
        <div className={`${styles.statCard} ${stats.savings >= 0 ? styles.positive : styles.negative}`}>
          <span className={styles.statLabel}>{stats.savings >= 0 ? t('stats.saved') : t('stats.overBudget')}</span>
          <span className={styles.statValue}>{fmt(Math.abs(stats.savings))}</span>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>{t('stats.purchases')}</span>
          <span className={styles.statValue}>{stats.purchaseCount}</span>
        </div>
      </div>

      {stats.categoryBreakdown.length > 0 && (
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{t('stats.byCategory')}</h2>
          <div className={styles.categories}>
            {stats.categoryBreakdown.map((cat) => (
              <div key={cat.category} className={styles.categoryRow}>
                <div className={styles.categoryMeta}>
                  <span className={styles.categoryName}>{cat.category}</span>
                  <span className={styles.categoryCount}>
                    {t('stats.purchaseCount', { count: cat.count })}
                  </span>
                  <span className={styles.categorySpend}>{fmt(cat.totalSpent)}</span>
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

      {monthlyStats && monthlyStats.length >= 2 && (
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{t('stats.spendTrend')}</h2>
          <div className={styles.card}>
            <SpendTrendChart data={monthlyStats} currency={currency} locale={locale} />
          </div>
        </section>
      )}

      {monthlyStats && monthlyStats.length >= 2 && (
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{t('stats.budgetVsActual')}</h2>
          <div className={styles.card}>
            <BudgetVsActualChart data={monthlyStats} currency={currency} locale={locale} />
          </div>
        </section>
      )}

      {stats.categoryBreakdown.length >= 2 && (
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{t('stats.byCategory')}</h2>
          <div className={styles.card}>
            <CategoryDonutChart
              data={stats.categoryBreakdown}
              currency={currency}
              locale={locale}
            />
          </div>
        </section>
      )}
    </main>
  )
}

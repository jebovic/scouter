import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useStats, useMonthlyStats } from '../hooks/usePurchase'
import { useSettings } from '../hooks/useSettings'
import { useMissions } from '../hooks/useMission'
import { useBadges } from '../hooks/useBadges'
import { BadgeRow, BadgeToast, BudgetHeatmap, ShoppingPersonaCard } from '../components/scouter'
import { StarRating } from '../components/scouter/StarRating'
import { SpendTrendChart } from '../components/charts/SpendTrendChart'
import { CategoryDonutChart } from '../components/charts/CategoryDonutChart'
import { BudgetVsActualChart } from '../components/charts/BudgetVsActualChart'
import { PersonaCard } from '../components/persona'
import { formatCurrencyLocale } from '../utils/format'
import { unlockIfEarned, ALL_BADGES } from '../utils/badges'
import type { Badge } from '../utils/badges'
import styles from './StatsPage.module.css'

export default function StatsPage() {
  const { t } = useTranslation()
  const { data: settings } = useSettings()
  const locale = settings?.locale ?? 'fr-FR'
  const currency = settings?.currency ?? 'EUR'

  const { data: stats, isLoading } = useStats()
  const { data: monthlyStats } = useMonthlyStats()
  const { missions } = useMissions()
  const { badges, unlock } = useBadges()
  const [newBadge, setNewBadge] = useState<Badge | null>(null)

  const fmt = (amount: number) => formatCurrencyLocale(amount, locale, currency)

  // Check for earned badges on load
  useEffect(() => {
    if (!stats) return

    const missionCount = missions.length
    const savedTotal = Math.max(0, stats.savings)
    const researchCount = 0 // Default to 0; can be enhanced with actual research tracking

    const earnedBadgeIds = unlockIfEarned(missionCount, savedTotal, researchCount)

    // Unlock new badges and show toast for the first new one
    let firstNewBadge: string | null = null
    earnedBadgeIds.forEach((badgeId) => {
      if (!badges.some((b) => b.id === badgeId)) {
        unlock(badgeId)
        if (!firstNewBadge) {
          firstNewBadge = badgeId
        }
      }
    })

    // Show toast only for the first new badge in this render
    if (firstNewBadge && firstNewBadge in ALL_BADGES) {
      const badgeKey = firstNewBadge as keyof typeof ALL_BADGES
      const def = ALL_BADGES[badgeKey]
      const badge: Badge = {
        id: def.id as any,
        name: def.name,
        description: def.description,
        icon: def.icon,
        unlockedAt: new Date(),
      }
      setNewBadge(badge)
    }
  }, [stats, missions.length, badges, unlock])

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

      <section className={styles.section}>
        <ShoppingPersonaCard />
      </section>

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

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>{t('stats.budgetVariance')}</h2>
        <div className={styles.card}>
          <BudgetHeatmap monthsBack={3} />
        </div>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Vos badges</h2>
        <BadgeRow badges={badges} />
      </section>

      <section className={styles.section}>
        <PersonaCard />
      </section>

      {newBadge && <BadgeToast badge={newBadge} onDismiss={() => setNewBadge(null)} />}
    </main>
  )
}

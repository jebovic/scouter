import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useStats, useMonthlyStats } from '../hooks/usePurchase'
import { useSettings } from '../hooks/useSettings'
import { useFormatCurrency } from '../hooks/useFormatCurrency'
import { useMissions } from '../hooks/useMission'
import { useBadges } from '../hooks/useBadges'
import { useSpendingAnalytics } from '../hooks/useSpendingAnalytics'
import { useBudgetHeatmap } from '../hooks/useBudgetHeatmap'
import { useCrossMission } from '../hooks/useCrossMission'
import type { CategoryStats, MerchantStats, TopItem, MonthStats } from '../api/spendinganalytics'
import BudgetHeatmapMission from '../components/mission/BudgetHeatmap'
import CrossMissionPanel from '../components/mission/CrossMissionPanel'
import { BadgeRow, BadgeToast, BudgetHeatmap, ShoppingPersonaCard } from '../components/scouter'
import { Topnav } from '../components/scouter/Topnav'
import { StarRating } from '../components/scouter/StarRating'
import { SpendTrendChart } from '../components/charts/SpendTrendChart'
import { CategoryDonutChart } from '../components/charts/CategoryDonutChart'
import { BudgetVsActualChart } from '../components/charts/BudgetVsActualChart'
import { PersonaCard } from '../components/persona'
import { formatCurrencyLocale } from '../utils/format'
import { unlockIfEarned, ALL_BADGES } from '../utils/badges'
import type { Badge } from '../utils/badges'
import styles from './StatsPage.module.css'

// ---------------------------------------------------------------------------
// Analytics sub-components (previously in AnalyticsPage.tsx)
// ---------------------------------------------------------------------------

function BudgetBar({ spending, budget, pct, fmt }: { spending: number; budget: number; pct: number; fmt: (n: number) => string }) {
  const { t } = useTranslation()
  const clampedPct = Math.min(100, pct)
  const isOver = pct > 100
  return (
    <div className={styles.budgetSection}>
      <div className={styles.budgetLabels}>
        <span>{fmt(spending)} {t('stats.spent')}</span>
        <span>{fmt(budget)} {t('stats.budget')}</span>
      </div>
      <div className={styles.budgetTrack}>
        <div
          className={isOver ? `${styles.budgetFill} ${styles.budgetOver}` : styles.budgetFill}
          style={{ '--w': `${clampedPct}%` } as React.CSSProperties}
        />
      </div>
      <div className={styles.budgetPct}>{t('stats.pctUsed', { pct: pct.toFixed(1) })}</div>
    </div>
  )
}

function CategoryDonut({ categories, fmt }: { categories: CategoryStats[]; fmt: (n: number) => string }) {
  const { t } = useTranslation()
  if (categories.length === 0) {
    return <div className={styles.analyticsEmpty}>{t('stats.noCategoryData')}</div>
  }

  const palette = [
    'var(--accent)',
    'var(--green)',
    'var(--orange)',
    'var(--purple)',
    '#06b6d4',
    '#f59e0b',
    '#ec4899',
    '#10b981',
  ]

  let cumulative = 0
  const stops = categories.map((c, i) => {
    const from = cumulative
    cumulative += c.pct
    const color = palette[i % palette.length]
    return `${color} ${from.toFixed(1)}% ${cumulative.toFixed(1)}%`
  })

  const gradient = `conic-gradient(${stops.join(', ')})`

  return (
    <div className={styles.donutWrap}>
      <div
        className={styles.donut}
        style={{ background: gradient } as React.CSSProperties}
        role="img"
        aria-label="Category spending breakdown"
      />
      {/* Screen-reader accessible table */}
      <table className={styles.srOnly} aria-label="Category spending breakdown table">
        <thead>
          <tr>
            <th scope="col">{t('stats.categoryCol')}</th>
            <th scope="col">{t('stats.amountCol')}</th>
            <th scope="col">{t('stats.pctCol')}</th>
          </tr>
        </thead>
        <tbody>
          {categories.map((c) => (
            <tr key={c.category}>
              <td>{c.category || t('stats.uncategorized')}</td>
              <td>{fmt(c.total)}</td>
              <td>{c.pct.toFixed(1)}%</td>
            </tr>
          ))}
        </tbody>
      </table>
      <ul className={styles.legend}>
        {categories.map((c, i) => (
          <li key={c.category} className={styles.legendItem}>
            <span
              className={styles.legendDot}
              style={{ background: palette[i % palette.length] } as React.CSSProperties}
            />
            <span className={styles.legendLabel}>{c.category || t('stats.uncategorized')}</span>
            <span className={styles.legendValue}>{fmt(c.total)}</span>
            <span className={styles.legendPct}>{c.pct.toFixed(1)}%</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

function MerchantBars({ merchants, fmt }: { merchants: MerchantStats[]; fmt: (n: number) => string }) {
  const { t } = useTranslation()
  if (merchants.length === 0) {
    return <div className={styles.analyticsEmpty}>{t('stats.noMerchantData')}</div>
  }
  const max = merchants[0].total
  return (
    <ul className={styles.merchantList}>
      {merchants.map((m) => (
        <li key={m.merchant} className={styles.merchantRow}>
          <span className={styles.merchantName}>{m.merchant || t('stats.unknownMerchant')}</span>
          <div className={styles.merchantTrack}>
            <div
              className={styles.merchantFill}
              style={{ '--w': `${max > 0 ? (m.total / max) * 100 : 0}%` } as React.CSSProperties}
            />
          </div>
          <span className={styles.merchantTotal}>{fmt(m.total)}</span>
        </li>
      ))}
    </ul>
  )
}

function TopItemsList({ items, fmt }: { items: TopItem[]; fmt: (n: number) => string }) {
  const { t } = useTranslation()
  if (items.length === 0) {
    return <div className={styles.analyticsEmpty}>{t('stats.noItemData')}</div>
  }
  return (
    <ol className={styles.topList}>
      {items.map((item, i) => (
        <li key={item.itemId} className={styles.topItem}>
          <span className={styles.topRank}>#{i + 1}</span>
          <div className={styles.topInfo}>
            <span className={styles.topName}>{item.itemName}</span>
            <span className={styles.topMission}>{item.missionName}</span>
          </div>
          <span className={styles.topPrice}>{fmt(item.price)}</span>
        </li>
      ))}
    </ol>
  )
}

function MonthlyChart({ months, fmt }: { months: MonthStats[]; fmt: (n: number) => string }) {
  const { t } = useTranslation()
  if (months.length === 0) {
    return <div className={styles.analyticsEmpty}>{t('stats.noMonthlyData')}</div>
  }
  const max = Math.max(...months.map((m) => m.avgPrice), 1)
  return (
    <div
      className={styles.monthlyChart}
      role="img"
      aria-label="Monthly spending trend chart"
    >
      {/* Screen-reader accessible table */}
      <table className={styles.srOnly} aria-label="Monthly spending trend table">
        <thead>
          <tr>
            <th scope="col">{t('stats.monthCol')}</th>
            <th scope="col">{t('stats.avgPriceCol')}</th>
          </tr>
        </thead>
        <tbody>
          {months.map((m) => (
            <tr key={m.month}>
              <td>{m.month}</td>
              <td>{fmt(m.avgPrice)}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {months.map((m) => (
        <div key={m.month} className={styles.monthBar} aria-hidden="true">
          <div
            className={styles.monthFill}
            style={{ '--h': `${(m.avgPrice / max) * 100}%` } as React.CSSProperties}
            title={`${m.month}: ${fmt(m.avgPrice)} ${t('stats.avgTooltip')}`}
          />
          <span className={styles.monthLabel}>{m.month.slice(5)}</span>
        </div>
      ))}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Analytics tab panel
// ---------------------------------------------------------------------------

function AnalyticsTab() {
  const { t } = useTranslation()
  const { data, isLoading, error } = useSpendingAnalytics()
  const { data: heatmapData } = useBudgetHeatmap()
  const { data: crossData } = useCrossMission()
  const { fmt, locale } = useFormatCurrency()

  if (isLoading) {
    return <div className={styles.analyticsLoading}>{t('stats.analyticsLoading')}</div>
  }

  if (error || !data) {
    return (
      <div className={styles.analyticsEmptyState}>
        <div className={styles.analyticsEmptyIcon}>📊</div>
        <h2>{t('stats.analyticsUnavailable')}</h2>
        <p>{t('stats.analyticsUnavailableDesc')}</p>
      </div>
    )
  }

  return (
    <div className={styles.analyticsContainer}>
      <div className={styles.kpiRow}>
        <div className={styles.kpi}>
          <div className={styles.kpiValue}>{fmt(data.totalBudget)}</div>
          <div className={styles.kpiLabel}>{t('stats.totalBudgetKpi')}</div>
        </div>
        <div className={styles.kpi}>
          <div className={styles.kpiValue}>{fmt(data.totalSpending)}</div>
          <div className={styles.kpiLabel}>{t('stats.totalSpendingKpi')}</div>
        </div>
        <div className={styles.kpi}>
          <div className={styles.kpiValue}>{data.budgetUsagePct.toFixed(1)}%</div>
          <div className={styles.kpiLabel}>{t('stats.budgetUsagePct')}</div>
        </div>
        <div className={styles.kpi}>
          <div className={styles.kpiValue}>{data.byCategory.length}</div>
          <div className={styles.kpiLabel}>{t('stats.categoriesKpi')}</div>
        </div>
      </div>

      <section className={styles.analyticsCard}>
        <h2 className={styles.analyticsCardTitle}>{t('stats.budgetUsageTitle')}</h2>
        <BudgetBar
          spending={data.totalSpending}
          budget={data.totalBudget}
          pct={data.budgetUsagePct}
          fmt={fmt}
        />
      </section>

      <section className={styles.analyticsCard}>
        <h2 className={styles.analyticsCardTitle}>{t('stats.categoryBreakdownTitle')}</h2>
        <CategoryDonut categories={data.byCategory} fmt={fmt} />
      </section>

      <section className={styles.analyticsCard}>
        <h2 className={styles.analyticsCardTitle}>{t('stats.topMerchantsTitle')}</h2>
        <MerchantBars merchants={data.byMerchant} fmt={fmt} />
      </section>

      <section className={styles.analyticsCard}>
        <h2 className={styles.analyticsCardTitle}>{t('stats.topItemsTitle')}</h2>
        <TopItemsList items={data.topItems} fmt={fmt} />
      </section>

      <section className={styles.analyticsCard}>
        <h2 className={styles.analyticsCardTitle}>{t('stats.monthlyTrendTitle')}</h2>
        <MonthlyChart months={data.monthlySummary} fmt={fmt} />
      </section>

      {heatmapData && (
        <section className={styles.analyticsCard}>
          <h2 className={styles.analyticsCardTitle}>{t('stats.heatmapTitle')}</h2>
          <BudgetHeatmapMission data={heatmapData} />
        </section>
      )}

      {crossData && crossData.missions.length > 0 && (
        <section className={styles.analyticsCard}>
          <h2 className={styles.analyticsCardTitle}>{t('stats.crossMissionTitle')}</h2>
          <CrossMissionPanel data={crossData} />
        </section>
      )}

      <div className={styles.generatedAt}>
        {t('stats.generatedAt')}{' '}
        {new Date(data.generatedAt).toLocaleString(locale, {
          dateStyle: 'medium',
          timeStyle: 'short',
        })}
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

type Tab = 'stats' | 'analytics'

export default function StatsPage() {
  const { t } = useTranslation()
  const { data: settings } = useSettings()
  const locale = settings?.locale ?? 'fr-FR'
  const currency = settings?.currency ?? 'EUR'

  const [activeTab, setActiveTab] = useState<Tab>('stats')

  const { data: stats, isLoading } = useStats()
  const { data: monthlyStats } = useMonthlyStats()
  const { missions } = useMissions()
  const { badges, unlock } = useBadges()
  const [newBadge, setNewBadge] = useState<Badge | null>(null)

  const fmt = (amount: number) => formatCurrencyLocale(amount, locale, currency)

  useEffect(() => {
    if (!stats) return

    const missionCount = missions.length
    const savedTotal = Math.max(0, stats.savings)
    const researchCount = 0

    const earnedBadgeIds = unlockIfEarned(missionCount, savedTotal, researchCount)

    let firstNewBadge: string | null = null
    earnedBadgeIds.forEach((badgeId) => {
      if (!badges.some((b) => b.id === badgeId)) {
        unlock(badgeId)
        if (!firstNewBadge) {
          firstNewBadge = badgeId
        }
      }
    })

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

  return (
    <>
      <Topnav />
      <main className={`page ${styles.page}`}>
      <h1 className={styles.heading}>{t('stats.title')}</h1>

      {/* Tab switcher */}
      <div className={styles.tabs} role="tablist" aria-label="Stats sections">
        <button
          role="tab"
          aria-selected={activeTab === 'stats'}
          aria-controls="panel-stats"
          id="tab-stats"
          className={`${styles.tab} ${activeTab === 'stats' ? styles.tabActive : ''}`}
          onClick={() => setActiveTab('stats')}
        >
          {t('stats.tabStats')}
        </button>
        <button
          role="tab"
          aria-selected={activeTab === 'analytics'}
          aria-controls="panel-analytics"
          id="tab-analytics"
          className={`${styles.tab} ${activeTab === 'analytics' ? styles.tabActive : ''}`}
          onClick={() => setActiveTab('analytics')}
        >
          {t('stats.tabAnalytics')}
        </button>
      </div>

      {/* Stats tab panel */}
      <div
        role="tabpanel"
        id="panel-stats"
        aria-labelledby="tab-stats"
        hidden={activeTab !== 'stats'}
      >
        {isLoading ? (
          <div className={styles.loading}>{t('stats.loading')}</div>
        ) : !stats || stats.purchaseCount === 0 ? (
          <div className={styles.empty}>
            <p className={styles.emptyIcon}>📊</p>
            <p className={styles.emptyTitle}>{t('stats.noData')}</p>
            <p className={styles.emptyDesc}>{t('stats.noDataDesc')}</p>
          </div>
        ) : (
          <>
            {(() => {
              const maxCategorySpend = Math.max(...stats.categoryBreakdown.map((c) => c.totalSpent), 1)
              return (
                <>
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
                    <h2 className={styles.sectionTitle}>{t('stats.badges')}</h2>
                    <BadgeRow badges={badges} />
                  </section>

                  <section className={styles.section}>
                    <PersonaCard />
                  </section>
                </>
              )
            })()}
          </>
        )}
      </div>

      {/* Analytics tab panel */}
      <div
        role="tabpanel"
        id="panel-analytics"
        aria-labelledby="tab-analytics"
        hidden={activeTab !== 'analytics'}
      >
        <AnalyticsTab />
      </div>

      {newBadge && <BadgeToast badge={newBadge} onDismiss={() => setNewBadge(null)} />}
    </main>
    </>
  )
}

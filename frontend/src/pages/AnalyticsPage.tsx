import { useSpendingAnalytics } from '../hooks/useSpendingAnalytics'
import { useBudgetHeatmap } from '../hooks/useBudgetHeatmap'
import { useCrossMission } from '../hooks/useCrossMission'
import type { CategoryStats, MerchantStats, TopItem, MonthStats } from '../api/spendinganalytics'
import BudgetHeatmap from '../components/mission/BudgetHeatmap'
import CrossMissionPanel from '../components/mission/CrossMissionPanel'
import styles from './AnalyticsPage.module.css'

const fmt = (v: number) =>
  new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(v)

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

function BudgetBar({ spending, budget, pct }: { spending: number; budget: number; pct: number }) {
  const clampedPct = Math.min(100, pct)
  const isOver = pct > 100
  return (
    <div className={styles.budgetSection}>
      <div className={styles.budgetLabels}>
        <span>{fmt(spending)} dépensé</span>
        <span>{fmt(budget)} budget</span>
      </div>
      <div className={styles.budgetTrack}>
        <div
          className={isOver ? `${styles.budgetFill} ${styles.budgetOver}` : styles.budgetFill}
          style={{ '--w': `${clampedPct}%` } as React.CSSProperties}
        />
      </div>
      <div className={styles.budgetPct}>{pct.toFixed(1)}% utilisé</div>
    </div>
  )
}

function CategoryDonut({ categories }: { categories: CategoryStats[] }) {
  if (categories.length === 0) {
    return <div className={styles.empty}>Aucune donnée de catégorie</div>
  }

  // Build conic-gradient stops
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
      <div className={styles.donut} style={{ background: gradient } as React.CSSProperties} />
      <ul className={styles.legend}>
        {categories.map((c, i) => (
          <li key={c.category} className={styles.legendItem}>
            <span
              className={styles.legendDot}
              style={{ background: palette[i % palette.length] } as React.CSSProperties}
            />
            <span className={styles.legendLabel}>{c.category || 'Sans catégorie'}</span>
            <span className={styles.legendValue}>{fmt(c.total)}</span>
            <span className={styles.legendPct}>{c.pct.toFixed(1)}%</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

function MerchantBars({ merchants }: { merchants: MerchantStats[] }) {
  if (merchants.length === 0) {
    return <div className={styles.empty}>Aucune donnée marchand</div>
  }
  const max = merchants[0].total
  return (
    <ul className={styles.merchantList}>
      {merchants.map((m) => (
        <li key={m.merchant} className={styles.merchantRow}>
          <span className={styles.merchantName}>{m.merchant || 'Inconnu'}</span>
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

function TopItemsList({ items }: { items: TopItem[] }) {
  if (items.length === 0) {
    return <div className={styles.empty}>Aucun article</div>
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

function MonthlyChart({ months }: { months: MonthStats[] }) {
  if (months.length === 0) {
    return <div className={styles.empty}>Aucune donnée mensuelle</div>
  }
  const max = Math.max(...months.map((m) => m.avgPrice), 1)
  return (
    <div className={styles.monthlyChart}>
      {months.map((m) => (
        <div key={m.month} className={styles.monthBar}>
          <div
            className={styles.monthFill}
            style={{ '--h': `${(m.avgPrice / max) * 100}%` } as React.CSSProperties}
            title={`${m.month}: ${fmt(m.avgPrice)} moy.`}
          />
          <span className={styles.monthLabel}>{m.month.slice(5)}</span>
        </div>
      ))}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function AnalyticsPage() {
  const { data, isLoading, error } = useSpendingAnalytics()
  const { data: heatmapData } = useBudgetHeatmap()
  const { data: crossData } = useCrossMission()

  if (isLoading) {
    return (
      <main className={styles.main}>
        <div className={styles.loading}>Chargement des analytics…</div>
      </main>
    )
  }

  if (error || !data) {
    return (
      <main className={styles.main}>
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>📊</div>
          <h2>Données indisponibles</h2>
          <p>Impossible de charger les analytics de dépenses.</p>
        </div>
      </main>
    )
  }

  return (
    <main className={styles.main}>
      <div className={styles.container}>
        <h1 className={styles.title}>Tableau de Bord Analytics</h1>

        {/* KPI row */}
        <div className={styles.kpiRow}>
          <div className={styles.kpi}>
            <div className={styles.kpiValue}>{fmt(data.totalBudget)}</div>
            <div className={styles.kpiLabel}>Budget total</div>
          </div>
          <div className={styles.kpi}>
            <div className={styles.kpiValue}>{fmt(data.totalSpending)}</div>
            <div className={styles.kpiLabel}>Dépenses (achetés)</div>
          </div>
          <div className={styles.kpi}>
            <div className={styles.kpiValue}>{data.budgetUsagePct.toFixed(1)}%</div>
            <div className={styles.kpiLabel}>Utilisation budget</div>
          </div>
          <div className={styles.kpi}>
            <div className={styles.kpiValue}>{data.byCategory.length}</div>
            <div className={styles.kpiLabel}>Catégories</div>
          </div>
        </div>

        {/* Budget usage bar */}
        <section className={styles.card}>
          <h2 className={styles.cardTitle}>Utilisation du budget</h2>
          <BudgetBar
            spending={data.totalSpending}
            budget={data.totalBudget}
            pct={data.budgetUsagePct}
          />
        </section>

        {/* Category donut */}
        <section className={styles.card}>
          <h2 className={styles.cardTitle}>Répartition par catégorie</h2>
          <CategoryDonut categories={data.byCategory} />
        </section>

        {/* Merchant bars */}
        <section className={styles.card}>
          <h2 className={styles.cardTitle}>Top marchands</h2>
          <MerchantBars merchants={data.byMerchant} />
        </section>

        {/* Top items */}
        <section className={styles.card}>
          <h2 className={styles.cardTitle}>Articles les plus chers</h2>
          <TopItemsList items={data.topItems} />
        </section>

        {/* Monthly trend */}
        <section className={styles.card}>
          <h2 className={styles.cardTitle}>Tendance mensuelle (6 mois)</h2>
          <MonthlyChart months={data.monthlySummary} />
        </section>

        {/* Budget heatmap */}
        {heatmapData && (
          <section className={styles.card}>
            <h2 className={styles.cardTitle}>Heatmap des dépenses par catégorie</h2>
            <BudgetHeatmap data={heatmapData} />
          </section>
        )}

        {/* Cross-mission comparison */}
        {crossData && crossData.missions.length > 0 && (
          <section className={styles.card}>
            <h2 className={styles.cardTitle}>Comparaison inter-missions</h2>
            <CrossMissionPanel data={crossData} />
          </section>
        )}

        <div className={styles.generatedAt}>
          Données générées le{' '}
          {new Date(data.generatedAt).toLocaleString('fr-FR', {
            dateStyle: 'medium',
            timeStyle: 'short',
          })}
        </div>
      </div>
    </main>
  )
}

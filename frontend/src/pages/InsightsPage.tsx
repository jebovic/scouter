import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listMissions } from '../api/missions'
import { analyzeSpending } from '../utils/spendingInsights'
import { useFormatCurrency } from '../hooks/useFormatCurrency'
import styles from './InsightsPage.module.css'

function SpendingBar({ pct, label }: { pct: number; label: string }) {
  return (
    <div className={styles.barWrap}>
      <div className={styles.barLabel}>{label}</div>
      <div className={styles.barTrack}>
        <div
          className={styles.barFill}
          style={{ '--w': `${Math.min(100, pct)}%` } as React.CSSProperties}
        />
      </div>
      <div className={styles.barPct}>{pct.toFixed(1)}%</div>
    </div>
  )
}

export default function InsightsPage() {
  const { data: missions, isLoading } = useQuery({
    queryKey: ['missions'],
    queryFn: listMissions,
    staleTime: 60_000,
  })
  const { currency, locale } = useFormatCurrency()

  // Aggregate items from all missions (use budget data only, no per-item fetch needed)
  const report = useMemo(() => {
    if (!missions || missions.length === 0) return null
    // Use mission-level data as proxy (no cross-mission item fetch for MVP)
    return analyzeSpending(
      missions.map((m) => ({ name: m.name, category: m.category, budget: m.budget })),
      // empty items — insights based on missions only
      [],
      currency,
      locale,
    )
  }, [missions, currency, locale])

  if (isLoading) {
    return (
      <main className={styles.main}>
        <div className={styles.loading}>Analyse en cours…</div>
      </main>
    )
  }

  if (!report || !missions || missions.length === 0) {
    return (
      <main className={styles.main}>
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>📊</div>
          <h2>Aucune donnée</h2>
          <p>Créez des missions pour voir vos insights.</p>
        </div>
      </main>
    )
  }

  const { fmt } = useFormatCurrency()

  return (
    <main className={styles.main}>
      <div className="container">
        <h1 className={styles.title}>📊 Insights Budget</h1>

        {/* KPI row */}
        <div className={styles.kpiRow}>
          <div className={styles.kpi}>
            <div className={styles.kpiValue}>{missions.length}</div>
            <div className={styles.kpiLabel}>Missions actives</div>
          </div>
          <div className={styles.kpi}>
            <div className={styles.kpiValue}>{fmt(report.totalBudget)}</div>
            <div className={styles.kpiLabel}>Budget total</div>
          </div>
          <div className={styles.kpi}>
            <div className={styles.kpiValue}>{report.utilizationPct.toFixed(0)}%</div>
            <div className={styles.kpiLabel}>Utilisation budget</div>
          </div>
        </div>

        {/* Insights */}
        {report.insights.length > 0 && (
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Recommandations</h2>
            <div className={styles.insightGrid}>
              {report.insights.map((ins, i) => (
                <div key={i} className={`${styles.insightCard} ${styles[ins.type]}`}>
                  <div className={styles.insightEmoji}>{ins.emoji}</div>
                  <div>
                    <div className={styles.insightTitle}>{ins.title}</div>
                    <div className={styles.insightDesc}>{ins.description}</div>
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}

        {/* Category breakdown */}
        {report.patterns.length > 0 && (
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Par catégorie</h2>
            {report.patterns.map((p) => (
              <SpendingBar key={p.category} pct={p.pctOfBudget} label={p.category} />
            ))}
          </section>
        )}

        {/* Missions table */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Toutes les missions</h2>
          <div className={styles.missionTable}>
            {missions.map((m) => (
              <div key={m.id} className={styles.missionRow}>
                <span className={styles.missionIcon}>{m.icon}</span>
                <span className={styles.missionName}>{m.name}</span>
                <span className={styles.missionCat}>{m.category}</span>
                <span className={styles.missionBudget}>{fmt(m.budget)}</span>
              </div>
            ))}
          </div>
        </section>
      </div>
    </main>
  )
}

import { useQueries } from '@tanstack/react-query'
import { listShoppingItems } from '../../api'
import { useMissions } from '../../hooks/useMission'
import type { Mission } from '../../types'
import type { ShoppingItem, ItemStatus } from '../../types/shopping'
import styles from './MissionComparison.module.css'

interface Props {
  missionIds: string[]
}

interface MissionStats {
  mission: Mission
  items: ShoppingItem[]
  totalBudget: number
  totalSpent: number
  totalEstimate: number
  itemCount: number
  completionPct: number
  statusCounts: Record<ItemStatus, number>
  avgDealScore: number | null
}

const STATUS_LABELS: Record<ItemStatus, string> = {
  buy: 'Buy',
  'flash-sale': 'Flash',
  preorder: 'Pre',
  defer: 'Defer',
  watch: 'Watch',
  crisis: 'Crisis',
  recommended: 'Rec',
  rejected: 'Rej',
}

const STATUS_COLORS: Record<ItemStatus, string> = {
  buy: 'var(--green)',
  'flash-sale': 'var(--orange)',
  preorder: 'var(--gold)',
  defer: 'var(--text-dim)',
  watch: 'var(--purple)',
  crisis: 'var(--coral)',
  recommended: 'var(--cyan)',
  rejected: 'var(--coral)',
}

const ALL_STATUSES: ItemStatus[] = [
  'buy', 'recommended', 'watch', 'flash-sale', 'preorder', 'defer', 'crisis', 'rejected',
]

function computeStats(mission: Mission, items: ShoppingItem[]): MissionStats {
  const totalSpent = items.reduce((sum, i) => sum + i.price, 0)
  const totalEstimate = items.reduce((sum, i) => sum + (i.originalEstimate ?? i.price), 0)
  const boughtItems = items.filter(i => i.status === 'buy' || i.status === 'flash-sale')
  const completionPct = items.length === 0 ? 0 : Math.round((boughtItems.length / items.length) * 100)

  const statusCounts = ALL_STATUSES.reduce(
    (acc, s) => ({ ...acc, [s]: 0 }),
    {} as Record<ItemStatus, number>,
  )
  for (const item of items) {
    statusCounts[item.status] = (statusCounts[item.status] ?? 0) + 1
  }

  return {
    mission,
    items,
    totalBudget: mission.budget,
    totalSpent,
    totalEstimate,
    itemCount: items.length,
    completionPct,
    statusCounts,
    avgDealScore: null,
  }
}

function formatCurrency(amount: number, currency = 'EUR', locale = 'fr-FR'): string {
  try {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency,
      maximumFractionDigits: 0,
    }).format(amount)
  } catch {
    return `${amount} ${currency}`
  }
}

function SpentBar({ spent, budget }: { spent: number; budget: number }) {
  const pct = budget === 0 ? 0 : Math.min(100, Math.round((spent / budget) * 100))
  const color = pct > 100 ? 'var(--coral)' : pct > 80 ? 'var(--gold)' : 'var(--green)'
  return (
    <div className={styles.spentBar}>
      <div
        className={styles.spentBarFill}
        style={{ '--fill': `${pct}%`, '--color': color } as React.CSSProperties}
      />
      <span className={styles.spentBarPct}>{pct}%</span>
    </div>
  )
}

function StatusDots({ counts }: { counts: Record<ItemStatus, number> }) {
  const active = ALL_STATUSES.filter(s => counts[s] > 0)
  if (active.length === 0) return <span className={styles.empty}>—</span>
  return (
    <div className={styles.statusDots}>
      {active.map(s => (
        <span
          key={s}
          className={styles.statusDot}
          style={{ '--dot-color': STATUS_COLORS[s] } as React.CSSProperties}
          title={`${STATUS_LABELS[s]}: ${counts[s]}`}
        >
          {STATUS_LABELS[s]} {counts[s]}
        </span>
      ))}
    </div>
  )
}

function PhaseChip({ phase }: { phase: Mission['phase'] }) {
  const colors: Record<Mission['phase'], string> = {
    researching: 'var(--cyan)',
    comparing: 'var(--purple)',
    buying: 'var(--gold)',
    done: 'var(--green)',
  }
  return (
    <span
      className={styles.phaseChip}
      style={{ '--phase-color': colors[phase] } as React.CSSProperties}
    >
      {phase}
    </span>
  )
}

function SkeletonCol() {
  return (
    <td className={styles.skeletonCol}>
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className={styles.skeletonLine} />
      ))}
    </td>
  )
}

export function MissionComparison({ missionIds }: Props) {
  const { missions } = useMissions()

  const selectedMissions = missionIds
    .map(id => missions.find(m => m.id === id))
    .filter((m): m is Mission => m !== undefined)

  const shoppingQueries = useQueries({
    queries: selectedMissions.map(m => ({
      queryKey: ['shopping', m.id],
      queryFn: () => listShoppingItems(m.id),
      enabled: Boolean(m.id),
      staleTime: 5 * 60_000,
    })),
  })

  if (missionIds.length === 0) {
    return (
      <div className={styles.empty}>
        <div className={styles.emptyIcon}>⊞</div>
        <p className={styles.emptyTitle}>AUCUNE MISSION SÉLECTIONNÉE</p>
        <p className={styles.emptyDesc}>Sélectionnez 2 à 4 missions pour les comparer</p>
      </div>
    )
  }

  const statsPerMission = selectedMissions.map((mission, idx) => {
    const query = shoppingQueries[idx]
    if (!query || query.isLoading || !query.data) return null
    return computeStats(mission, query.data)
  })

  const isLoading = shoppingQueries.some(q => q.isLoading)

  return (
    <div className={styles.wrapper}>
      <div className={styles.tableScroll}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th className={styles.rowHeader}>MÉTRIQUE</th>
              {selectedMissions.map((m, i) => (
                <th key={m.id} className={styles.missionHeader}>
                  <div className={styles.missionHeaderContent}>
                    <span className={styles.missionIcon}>{m.icon}</span>
                    <span className={styles.missionName}>{m.name}</span>
                    <PhaseChip phase={m.phase} />
                  </div>
                  {isLoading && !statsPerMission[i] && <span className={styles.loadingDot} />}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className={styles.rowLabel}>Budget total</td>
              {selectedMissions.map((m, i) => {
                const s = statsPerMission[i]
                return (
                  <td key={m.id} className={styles.cell}>
                    {s ? formatCurrency(s.totalBudget, m.currency, m.locale) : <div className={styles.skeletonLine} />}
                  </td>
                )
              })}
            </tr>
            <tr>
              <td className={styles.rowLabel}>Dépensé</td>
              {selectedMissions.map((m, i) => {
                const s = statsPerMission[i]
                return (
                  <td key={m.id} className={styles.cell}>
                    {s ? formatCurrency(s.totalSpent, m.currency, m.locale) : <div className={styles.skeletonLine} />}
                  </td>
                )
              })}
            </tr>
            <tr>
              <td className={styles.rowLabel}>Budget vs dépensé</td>
              {selectedMissions.map((m, i) => {
                const s = statsPerMission[i]
                return (
                  <td key={m.id} className={styles.cell}>
                    {s ? <SpentBar spent={s.totalSpent} budget={s.totalBudget} /> : <div className={styles.skeletonLine} />}
                  </td>
                )
              })}
            </tr>
            <tr>
              <td className={styles.rowLabel}>Articles</td>
              {selectedMissions.map((m, i) => {
                const s = statsPerMission[i]
                return (
                  <td key={m.id} className={styles.cell}>
                    {s != null ? s.itemCount : <div className={styles.skeletonLine} />}
                  </td>
                )
              })}
            </tr>
            <tr>
              <td className={styles.rowLabel}>Complétion</td>
              {selectedMissions.map((m, i) => {
                const s = statsPerMission[i]
                return (
                  <td key={m.id} className={styles.cell}>
                    {s != null
                      ? (
                          <span
                            className={styles.completionPct}
                            style={{
                              '--pct-color': s.completionPct >= 80
                                ? 'var(--green)'
                                : s.completionPct >= 40
                                  ? 'var(--gold)'
                                  : 'var(--text-mid)',
                            } as React.CSSProperties}
                          >
                            {s.completionPct}%
                          </span>
                        )
                      : <div className={styles.skeletonLine} />}
                  </td>
                )
              })}
            </tr>
            <tr>
              <td className={styles.rowLabel}>Statuts articles</td>
              {selectedMissions.map((m, i) => {
                const s = statsPerMission[i]
                return (
                  <td key={m.id} className={styles.cell}>
                    {s ? <StatusDots counts={s.statusCounts} /> : <div className={styles.skeletonLine} />}
                  </td>
                )
              })}
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}

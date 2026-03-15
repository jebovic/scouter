import { Link } from 'react-router-dom'
import type { MissionComparison, CrossMissionResponse } from '../../api/crossmission'
import styles from './CrossMissionPanel.module.css'

const fmt = (v: number) =>
  new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(v)

function efficiencyClass(e: MissionComparison['efficiency']): string {
  switch (e) {
    case 'excellent': return styles.efficiencyExcellent
    case 'good': return styles.efficiencyGood
    case 'overspent': return styles.efficiencyOverspent
    default: return styles.efficiencyNoData
  }
}

function efficiencyLabel(e: MissionComparison['efficiency']): string {
  switch (e) {
    case 'excellent': return 'Excellent'
    case 'good': return 'Bon'
    case 'overspent': return 'Dépassé'
    default: return 'Sans données'
  }
}

interface Props {
  data: CrossMissionResponse
}

export default function CrossMissionPanel({ data }: Props) {
  if (data.missions.length === 0) {
    return <div className={styles.empty}>Aucune mission active pour la comparaison.</div>
  }

  return (
    <div className={styles.panel}>
      {data.totalSavingsAllMissions > 0 && (
        <div className={styles.summary}>
          Total économisé toutes missions&nbsp;: {fmt(data.totalSavingsAllMissions)}
        </div>
      )}
      <ul className={styles.missionList}>
        {data.missions.map((m) => {
          const isBest = data.bestMission?.missionId === m.missionId
          const clampedPct = Math.min(100, m.budgetUsedPct)
          const isOver = m.budgetUsedPct > 100

          return (
            <li key={m.missionId}>
              <Link to={`/missions/${m.slug}`} className={styles.missionRow}>
                <div className={styles.rowHeader}>
                  <span className={styles.missionName}>{m.name}</span>
                  {isBest && <span className={styles.bestBadge}>Meilleure</span>}
                  <span
                    className={`${styles.savingsBadge} ${
                      m.savingsAmount > 0
                        ? styles.savingsPositive
                        : m.savingsAmount < 0
                        ? styles.savingsNegative
                        : styles.savingsNeutral
                    }`}
                  >
                    {m.savingsAmount >= 0 ? '+' : ''}
                    {fmt(m.savingsAmount)}
                  </span>
                  <span className={`${styles.efficiencyPill} ${efficiencyClass(m.efficiency)}`}>
                    {efficiencyLabel(m.efficiency)}
                  </span>
                </div>

                {m.budget > 0 && (
                  <>
                    <div className={styles.barTrack}>
                      <div
                        className={`${styles.barFill} ${isOver ? styles.barFillOver : ''}`}
                        style={{ '--w': `${clampedPct}%` } as React.CSSProperties}
                      />
                    </div>
                    <div className={styles.rowMeta}>
                      <span>{fmt(m.totalPrice)} dépensé</span>
                      <span>{m.budgetUsedPct.toFixed(1)}% du budget</span>
                      <span>{m.itemCount} article{m.itemCount !== 1 ? 's' : ''}</span>
                    </div>
                  </>
                )}

                {m.budget === 0 && (
                  <div className={styles.rowMeta}>
                    <span>{fmt(m.totalPrice)} dépensé</span>
                    <span>{m.itemCount} article{m.itemCount !== 1 ? 's' : ''}</span>
                    {m.savingsPct !== 0 && (
                      <span>{m.savingsPct.toFixed(1)}% économisé</span>
                    )}
                  </div>
                )}
              </Link>
            </li>
          )
        })}
      </ul>
    </div>
  )
}

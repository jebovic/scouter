import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useBudgetAlertBanners } from '../../hooks/useBudgetAlertBanners'
import type { BudgetAlert } from '../../api/budgetalert'
import styles from './BudgetAlertBanner.module.css'

const LEVEL_ICONS: Record<BudgetAlert['alertLevel'], string> = {
  info: 'ℹ',
  warning: '⚠',
  critical: '🚨',
}

interface SingleBannerProps {
  alert: BudgetAlert
  onDismiss: (id: string) => void
}

function SingleBanner({ alert, onDismiss }: SingleBannerProps) {
  const icon = LEVEL_ICONS[alert.alertLevel]
  const firstRec = alert.recommendations[0]

  return (
    <div className={styles.banner} data-level={alert.alertLevel} role="alert">
      <div className={styles.iconCol}>
        <span className={styles.icon} aria-hidden="true">{icon}</span>
      </div>
      <div className={styles.body}>
        <div className={styles.header}>
          <span className={styles.missionName}>{alert.missionName}</span>
          <span className={styles.pct}>{Math.round(alert.percentage)}%</span>
        </div>
        <p className={styles.message}>{alert.message}</p>
        {firstRec && <p className={styles.rec}>{firstRec}</p>}
        <Link
          to={`/missions/${alert.missionId}`}
          className={styles.link}
        >
          Voir la mission →
        </Link>
      </div>
      <button
        className={styles.dismiss}
        onClick={() => onDismiss(alert.missionId)}
        aria-label={`Masquer l'alerte pour ${alert.missionName}`}
      >
        ✕
      </button>
    </div>
  )
}

export function BudgetAlertBanner() {
  const { alerts, isLoading } = useBudgetAlertBanners()
  const [dismissed, setDismissed] = useState<Set<string>>(new Set())

  if (isLoading || alerts.length === 0) return null

  const visible = alerts.filter((a) => !dismissed.has(a.missionId))
  if (visible.length === 0) return null

  function handleDismiss(missionId: string) {
    setDismissed((prev) => new Set([...prev, missionId]))
  }

  return (
    <div className={styles.stack} role="region" aria-label="Alertes budget">
      {visible.map((alert) => (
        <SingleBanner
          key={alert.missionId}
          alert={alert}
          onDismiss={handleDismiss}
        />
      ))}
    </div>
  )
}

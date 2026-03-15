export interface BudgetAlert {
  id: string
  type: 'overspend' | 'velocity' | 'milestone' | 'suggestion'
  severity: 'info' | 'warning' | 'danger'
  title: string
  message: string
  timestamp: number
}

export function detectBudgetAlerts(
  spent: number,
  budget: number,
  itemCount: number,
  previousSpent?: number,
): BudgetAlert[] {
  const alerts: BudgetAlert[] = []
  const pct = budget > 0 ? spent / budget : 0
  const now = Date.now()

  if (pct >= 1.0) {
    alerts.push({
      id: `overspend-${now}`,
      type: 'overspend',
      severity: 'danger',
      title: '⚠️ Budget dépassé',
      message: `Vous avez dépassé votre budget de ${((pct - 1) * 100).toFixed(0)}%`,
      timestamp: now,
    })
  } else if (pct >= 0.8) {
    alerts.push({
      id: `warning-${now}`,
      type: 'overspend',
      severity: 'warning',
      title: '🔔 Budget presque atteint',
      message: `${(pct * 100).toFixed(0)}% du budget utilisé`,
      timestamp: now,
    })
  }

  if (pct >= 0.5 && pct < 0.8) {
    alerts.push({
      id: `milestone-50-${now}`,
      type: 'milestone',
      severity: 'info',
      title: '📊 Mi-chemin',
      message: `50% du budget alloué avec ${itemCount} articles`,
      timestamp: now,
    })
  }

  if (previousSpent !== undefined && previousSpent > 0) {
    const velocityPct = (spent - previousSpent) / previousSpent
    if (velocityPct > 0.3) {
      alerts.push({
        id: `velocity-${now}`,
        type: 'velocity',
        severity: 'warning',
        title: '📈 Accélération des dépenses',
        message: `+${(velocityPct * 100).toFixed(0)}% d'augmentation détectée`,
        timestamp: now,
      })
    }
  }

  return alerts
}

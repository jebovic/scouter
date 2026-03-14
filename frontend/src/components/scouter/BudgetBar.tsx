import styles from './BudgetBar.module.css'

interface BudgetBarProps {
  spent: number
  budget: number
  currency?: string
}

function formatCurrency(value: number, currency = 'USD'): string {
  return new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(value)
}

export function BudgetBar({ spent, budget, currency = 'USD' }: BudgetBarProps) {
  const pct = budget > 0 ? Math.min((spent / budget) * 100, 100) : 0
  const over = spent > budget

  const barColor = over
    ? 'var(--budget-over)'
    : pct > 80
    ? 'var(--budget-warn)'
    : 'var(--budget-ok)'

  return (
    <div className={styles.wrapper}>
      <div className={styles.track}>
        <div
          className={`budget-bar-fill ${styles.fill}`}
          style={{
            '--bar-width': `${pct}%`,
            '--bar-color': barColor,
            '--bar-glow': `${barColor}80`,
          } as React.CSSProperties}
        />
      </div>
      <div className={styles.labels}>
        <span
          className={styles.spent}
          style={{ '--bar-color': barColor } as React.CSSProperties}
        >
          {formatCurrency(spent, currency)}
        </span>
        <span className={styles.budget}>
          / {formatCurrency(budget, currency)}
        </span>
      </div>
    </div>
  )
}

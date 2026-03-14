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
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <div
        style={{
          height: 6,
          borderRadius: 3,
          background: 'var(--raised)',
          overflow: 'hidden',
          position: 'relative',
        }}
      >
        <div
          className="budget-bar-fill"
          style={{
            height: '100%',
            width: `${pct}%`,
            background: barColor,
            borderRadius: 3,
            transition: 'width 0.4s ease, background 0.3s ease',
            boxShadow: `0 0 10px ${barColor}80`,
          }}
        />
      </div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          fontSize: '0.75rem',
          fontFamily: 'var(--font-mono)',
        }}
      >
        <span style={{ color: barColor }}>
          {formatCurrency(spent, currency)}
        </span>
        <span style={{ color: 'var(--text-dim)' }}>
          / {formatCurrency(budget, currency)}
        </span>
      </div>
    </div>
  )
}

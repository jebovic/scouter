import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import type { MonthlyStatPoint } from '../../api/purchase'
import styles from './BudgetVsActualChart.module.css'

interface Props {
  data: MonthlyStatPoint[]
  currency: string
  locale: string
}

export function BudgetVsActualChart({ data, currency, locale }: Props) {
  const fmt = (v: number) =>
    new Intl.NumberFormat(locale, { style: 'currency', currency, notation: 'compact' }).format(v)

  return (
    <div className={styles.wrap}>
      <ResponsiveContainer width="100%" height={200}>
        <BarChart data={data} margin={{ top: 5, right: 10, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis dataKey="month" tick={{ fill: 'var(--text-dim)', fontSize: 11 }} />
          <YAxis tickFormatter={fmt} tick={{ fill: 'var(--text-dim)', fontSize: 11 }} width={60} />
          <Tooltip
            formatter={(v) => (typeof v === 'number' ? fmt(v) : String(v))}
            contentStyle={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderRadius: '8px',
              color: 'var(--text)',
            }}
          />
          <Legend
            formatter={(v) => (
              <span style={{ color: 'var(--text-mid)', fontSize: 12 }}>
                {v === 'totalBudget' ? 'Budget' : 'Spent'}
              </span>
            )}
          />
          <Bar
            dataKey="totalBudget"
            fill="var(--border-mid)"
            radius={[3, 3, 0, 0]}
            name="totalBudget"
          />
          <Bar
            dataKey="totalSpent"
            fill="var(--cyan)"
            radius={[3, 3, 0, 0]}
            name="totalSpent"
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}

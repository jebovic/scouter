import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import type { MonthlyStatPoint } from '../../api/purchase'
import styles from './SpendTrendChart.module.css'

interface Props {
  data: MonthlyStatPoint[]
  currency: string
  locale: string
}

export function SpendTrendChart({ data, currency, locale }: Props) {
  const fmt = (v: number) =>
    new Intl.NumberFormat(locale, { style: 'currency', currency, notation: 'compact' }).format(v)

  return (
    <div className={styles.wrap}>
      <ResponsiveContainer width="100%" height={200}>
        <AreaChart data={data} margin={{ top: 5, right: 10, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id="spendGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="var(--cyan)" stopOpacity={0.3} />
              <stop offset="95%" stopColor="var(--cyan)" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis dataKey="month" tick={{ fill: 'var(--text-dim)', fontSize: 11 }} />
          <YAxis tickFormatter={fmt} tick={{ fill: 'var(--text-dim)', fontSize: 11 }} width={60} />
          <Tooltip
            formatter={(v) => [typeof v === 'number' ? fmt(v) : String(v), 'Spent']}
            contentStyle={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderRadius: '8px',
              color: 'var(--text)',
            }}
          />
          <Area
            type="monotone"
            dataKey="totalSpent"
            stroke="var(--cyan)"
            strokeWidth={2}
            fill="url(#spendGrad)"
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}

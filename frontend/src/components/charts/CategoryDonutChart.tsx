import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import type { CategoryStat } from '../../api/purchase'
import styles from './CategoryDonutChart.module.css'

const COLORS = [
  'var(--cyan)',
  'var(--gold)',
  'var(--purple)',
  'var(--coral)',
  'var(--green)',
  '#60a5fa',
  '#f472b6',
]

interface Props {
  data: CategoryStat[]
  currency: string
  locale: string
}

export function CategoryDonutChart({ data, currency, locale }: Props) {
  const fmt = (v: number) =>
    new Intl.NumberFormat(locale, { style: 'currency', currency, notation: 'compact' }).format(v)

  const chartData = data.map((c) => ({ name: c.category, value: c.totalSpent }))

  return (
    <div className={styles.wrap}>
      <ResponsiveContainer width="100%" height={220}>
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="45%"
            innerRadius={55}
            outerRadius={80}
            paddingAngle={3}
            dataKey="value"
          >
            {chartData.map((_, i) => (
              <Cell key={i} fill={COLORS[i % COLORS.length]} />
            ))}
          </Pie>
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
            iconType="circle"
            iconSize={8}
            formatter={(v) => (
              <span style={{ color: 'var(--text-mid)', fontSize: 12 }}>{v}</span>
            )}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  )
}

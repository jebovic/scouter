import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import type { ShoppingItem } from '../../types'

interface CostBreakdownProps {
  items: ShoppingItem[]
  currency?: string
}

const COLORS = [
  'var(--cyan)',
  'var(--gold)',
  'var(--green)',
  'var(--purple)',
  'var(--orange)',
  'var(--coral)',
]

export function CostBreakdown({ items, currency = 'USD' }: CostBreakdownProps) {
  const byCat = items.reduce<Record<string, number>>((acc, item) => {
    acc[item.costCategory] = (acc[item.costCategory] ?? 0) + item.price
    return acc
  }, {})

  const data = Object.entries(byCat).map(([name, value]) => ({ name, value }))

  if (data.length === 0) {
    return (
      <div style={{ padding: '1rem', textAlign: 'center', color: 'var(--text-dim)', fontSize: '0.8rem' }}>
        No items tracked
      </div>
    )
  }

  const fmt = (n: number) =>
    new Intl.NumberFormat(undefined, { style: 'currency', currency, maximumFractionDigits: 0 }).format(n)

  return (
    <ResponsiveContainer width="100%" height={220}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          innerRadius={50}
          outerRadius={80}
          paddingAngle={3}
          dataKey="value"
        >
          {data.map((_, i) => (
            <Cell key={i} fill={COLORS[i % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip
          contentStyle={{
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            borderRadius: 8,
            fontFamily: 'var(--font-mono)',
            fontSize: '0.8rem',
          }}
          formatter={(value) => [fmt(value as number), 'Cost']}
        />
        <Legend
          wrapperStyle={{ fontSize: '0.75rem', fontFamily: 'var(--font-mono)', color: 'var(--text-mid)' }}
        />
      </PieChart>
    </ResponsiveContainer>
  )
}

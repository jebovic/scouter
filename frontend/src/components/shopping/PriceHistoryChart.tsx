import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import type { PriceSnapshot } from '../../types'

interface PriceHistoryChartProps {
  snapshots: PriceSnapshot[]
  currency?: string
}

export function PriceHistoryChart({ snapshots, currency = 'USD' }: PriceHistoryChartProps) {
  if (snapshots.length === 0) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-dim)', fontSize: '0.8rem' }}>
        No price history yet
      </div>
    )
  }

  const fmt = (n: number) =>
    new Intl.NumberFormat(undefined, { style: 'currency', currency, maximumFractionDigits: 0 }).format(n)

  const data = [...snapshots]
    .sort((a, b) => new Date(a.recordedAt).getTime() - new Date(b.recordedAt).getTime())
    .map((s) => ({
      date: new Date(s.recordedAt).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
      price: s.price,
    }))

  const prices = data.map((d) => d.price)
  const minPrice = Math.min(...prices)
  const maxPrice = Math.max(...prices)

  return (
    <ResponsiveContainer width="100%" height={180}>
      <LineChart data={data} margin={{ top: 8, right: 16, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
        <XAxis
          dataKey="date"
          tick={{ fill: 'var(--text-dim)', fontSize: 10, fontFamily: 'var(--font-mono)' }}
        />
        <YAxis
          domain={[minPrice * 0.95, maxPrice * 1.05]}
          tick={{ fill: 'var(--text-dim)', fontSize: 10, fontFamily: 'var(--font-mono)' }}
          tickFormatter={fmt}
          width={60}
        />
        <Tooltip
          contentStyle={{
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            borderRadius: 8,
            fontFamily: 'var(--font-mono)',
            fontSize: '0.8rem',
          }}
          formatter={(value) => [fmt(value as number), 'Price']}
        />
        <Line
          type="monotone"
          dataKey="price"
          stroke="var(--cyan)"
          strokeWidth={2}
          dot={{ fill: 'var(--cyan)', r: 3 }}
          activeDot={{ r: 5 }}
        />
      </LineChart>
    </ResponsiveContainer>
  )
}

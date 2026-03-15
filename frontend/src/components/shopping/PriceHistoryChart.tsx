import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
  ReferenceDot,
} from 'recharts'
import type { PriceSnapshot } from '../../types'
import styles from './PriceHistoryChart.module.css'

interface PriceHistoryChartProps {
  snapshots: PriceSnapshot[]
  currency?: string
  targetPrice?: number
  showTrendLine?: boolean
}

// Linear regression: returns y-values at each x index
function linearRegression(values: number[]): number[] {
  const n = values.length
  if (n < 2) return values.slice()
  const sumX = (n * (n - 1)) / 2
  const sumX2 = (n * (n - 1) * (2 * n - 1)) / 6
  const sumY = values.reduce((a, v) => a + v, 0)
  const sumXY = values.reduce((a, v, i) => a + i * v, 0)
  const denom = n * sumX2 - sumX * sumX
  if (denom === 0) return values.slice()
  const slope = (n * sumXY - sumX * sumY) / denom
  const intercept = (sumY - slope * sumX) / n
  return values.map((_, i) => slope * i + intercept)
}

// Custom dot renderer that highlights min or max
interface AnnotatedDotProps {
  cx?: number
  cy?: number
  index?: number
  minIndex: number
  maxIndex: number
}

function AnnotatedDot({ cx, cy, index, minIndex, maxIndex }: AnnotatedDotProps) {
  if (cx === undefined || cy === undefined || index === undefined) return null
  const isMin = index === minIndex
  const isMax = index === maxIndex
  if (!isMin && !isMax) {
    return <circle cx={cx} cy={cy} r={3} fill="var(--cyan)" />
  }
  const color = isMin ? 'var(--success, #22c55e)' : 'var(--coral, #f87171)'
  const label = isMin ? 'Min' : 'Max'
  return (
    <g>
      <circle cx={cx} cy={cy} r={5} fill={color} stroke="var(--surface)" strokeWidth={1.5} />
      <text
        x={cx}
        y={isMin ? cy + 14 : cy - 8}
        textAnchor="middle"
        fill={color}
        fontSize={9}
        fontFamily="var(--font-mono)"
        fontWeight="600"
      >
        {label}
      </text>
    </g>
  )
}

export function PriceHistoryChart({
  snapshots,
  currency = 'USD',
  targetPrice,
  showTrendLine = true,
}: PriceHistoryChartProps) {
  const fmt = (n: number) =>
    new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
      maximumFractionDigits: 0,
    }).format(n)

  if (snapshots.length === 0) {
    return (
      <div className={styles.empty}>
        <span className={styles.emptyIcon} aria-hidden="true">📊</span>
        <p className={styles.emptyText}>
          No price history yet — prices will be tracked automatically
        </p>
      </div>
    )
  }

  const sorted = [...snapshots].sort(
    (a, b) => new Date(a.recordedAt).getTime() - new Date(b.recordedAt).getTime(),
  )

  const prices = sorted.map((s) => s.price)
  const trendValues = showTrendLine && prices.length >= 2 ? linearRegression(prices) : null

  const data = sorted.map((s, i) => ({
    date: new Date(s.recordedAt).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
    price: s.price,
    trend: trendValues ? Math.round(trendValues[i] * 100) / 100 : undefined,
  }))

  const minPrice = Math.min(...prices)
  const maxPrice = Math.max(...prices)
  const minIndex = prices.indexOf(minPrice)
  const maxIndex = prices.indexOf(maxPrice)

  const firstPrice = prices[0]
  const lastPrice = prices[prices.length - 1]
  const pctChange = firstPrice !== 0 ? ((lastPrice - firstPrice) / firstPrice) * 100 : 0
  const pctLabel = `${pctChange >= 0 ? '+' : ''}${pctChange.toFixed(1)}%`
  const pctColor = pctChange <= 0 ? 'var(--success, #22c55e)' : 'var(--coral, #f87171)'

  // Y-axis domain: accommodate target price line if outside data range
  const yMin = targetPrice !== undefined ? Math.min(minPrice, targetPrice) * 0.95 : minPrice * 0.95
  const yMax = targetPrice !== undefined ? Math.max(maxPrice, targetPrice) * 1.05 : maxPrice * 1.05

  return (
    <div className={styles.wrapper}>
      {/* Price change badge */}
      <div className={styles.pctBadge} style={{ color: pctColor }}>
        {pctChange <= 0 ? '▼' : '▲'} {pctLabel}
      </div>

      <ResponsiveContainer width="100%" height={180}>
        <LineChart data={data} margin={{ top: 8, right: 16, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis
            dataKey="date"
            tick={{ fill: 'var(--text-dim)', fontSize: 10, fontFamily: 'var(--font-mono)' }}
          />
          <YAxis
            domain={[yMin, yMax]}
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
            formatter={(value, name) => {
              if (name === 'trend') return [fmt(value as number), 'Trend']
              return [fmt(value as number), 'Price']
            }}
          />

          {/* Target price reference line */}
          {targetPrice !== undefined && (
            <ReferenceLine
              y={targetPrice}
              stroke="var(--gold, #f59e0b)"
              strokeDasharray="6 3"
              strokeWidth={1.5}
              label={{
                value: `🎯 Target: ${fmt(targetPrice)}`,
                position: 'insideTopRight',
                fill: 'var(--gold, #f59e0b)',
                fontSize: 10,
                fontFamily: 'var(--font-mono)',
              }}
            />
          )}

          {/* Trend line */}
          {showTrendLine && trendValues && (
            <Line
              type="linear"
              dataKey="trend"
              stroke="var(--purple, #a855f7)"
              strokeWidth={1.5}
              strokeDasharray="4 3"
              dot={false}
              activeDot={false}
              legendType="none"
            />
          )}

          {/* Main price line with annotated dots */}
          <Line
            type="monotone"
            dataKey="price"
            stroke="var(--cyan)"
            strokeWidth={2}
            dot={(props) => (
              <AnnotatedDot
                key={`dot-${props.index}`}
                cx={props.cx}
                cy={props.cy}
                index={props.index}
                minIndex={minIndex}
                maxIndex={maxIndex}
              />
            )}
            activeDot={{ r: 5 }}
          />

          {/* Min/Max reference dots for tooltip positioning */}
          {prices.length > 1 && minIndex !== maxIndex && (
            <>
              <ReferenceDot
                x={data[minIndex].date}
                y={minPrice}
                r={0}
              />
              <ReferenceDot
                x={data[maxIndex].date}
                y={maxPrice}
                r={0}
              />
            </>
          )}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

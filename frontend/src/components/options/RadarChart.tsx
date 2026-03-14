import {
  RadarChart as ReRadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import type { Option } from '../../types'
import styles from './RadarChart.module.css'

interface RadarChartProps {
  options: Option[]
}

const COLORS = ['#00e5ff', '#ffd93d', '#00d68f', '#a855f7', '#f7974f']

export function RadarChart({ options }: RadarChartProps) {
  // Only use score attributes
  const scoreKeys = Array.from(
    new Set(
      options.flatMap((o) =>
        o.attributes.filter((a) => a.type === 'score').map((a) => a.key),
      ),
    ),
  )

  if (scoreKeys.length < 3) return null

  const attrLabels = Object.fromEntries(
    options
      .flatMap((o) => o.attributes)
      .map((a) => [a.key, a.label]),
  )

  const data = scoreKeys.map((key) => {
    const entry: Record<string, string | number> = { subject: attrLabels[key] ?? key }
    options.forEach((o) => {
      const attr = o.attributes.find((a) => a.key === key)
      entry[o.name] = attr ? Number(attr.value) : 0
    })
    return entry
  })

  return (
    <div className={styles.container}>
      <ResponsiveContainer width="100%" height={300}>
        <ReRadarChart data={data}>
          <PolarGrid stroke="#2a3558" />
          <PolarAngleAxis
            dataKey="subject"
            tick={{ fill: '#6b7db3', fontSize: 11, fontFamily: 'var(--font-mono)' }}
          />
          {options.map((o, i) => (
            <Radar
              key={o.id}
              name={o.name}
              dataKey={o.name}
              stroke={COLORS[i % COLORS.length]}
              fill={COLORS[i % COLORS.length]}
              fillOpacity={0.15}
            />
          ))}
          <Legend
            wrapperStyle={{ fontSize: '0.75rem', fontFamily: 'var(--font-mono)', color: 'var(--text-mid)' }}
          />
        </ReRadarChart>
      </ResponsiveContainer>
    </div>
  )
}

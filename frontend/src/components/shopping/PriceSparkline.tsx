import { usePriceSnapshots } from '../../hooks'

interface PriceSparklineProps {
  missionId: string
  itemId: string
}

const W = 56
const H = 20

export function PriceSparkline({ missionId, itemId }: PriceSparklineProps) {
  const { snapshots } = usePriceSnapshots(missionId, itemId)

  if (snapshots.length < 2) return null

  const prices = snapshots.map((s) => s.price)
  const min = Math.min(...prices)
  const max = Math.max(...prices)
  const range = max - min || 1

  const pts = prices.map((p, i) => {
    const x = (i / (prices.length - 1)) * W
    const y = H - ((p - min) / range) * H
    return `${x.toFixed(1)},${y.toFixed(1)}`
  })

  const last = prices[prices.length - 1]
  const first = prices[0]
  const color = last < first ? 'var(--green)' : last > first ? 'var(--coral)' : 'var(--text-dim)'

  return (
    <svg width={W} height={H} viewBox={`0 0 ${W} ${H}`} aria-hidden="true">
      <polyline
        points={pts.join(' ')}
        fill="none"
        stroke={color}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

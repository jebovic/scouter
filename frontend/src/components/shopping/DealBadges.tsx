import { TrendBadge } from './TrendBadge'
import { DealScoreBadge } from './DealScoreBadge'
import { DealExplainerPopover } from './DealExplainerPopover'
import type { DealScore } from '../../types'

interface DealBadgesProps {
  itemId: string
  dealScore: DealScore
}

export function DealBadges({ itemId, dealScore }: DealBadgesProps) {
  return (
    <>
      <TrendBadge trend={dealScore.trend} />
      <DealScoreBadge score={dealScore} />
      <DealExplainerPopover
        itemId={itemId}
        dealScore={Math.round(Math.max(0, Math.min(100, dealScore.pctBelowAvg)))}
        trend={dealScore.trend}
      />
    </>
  )
}

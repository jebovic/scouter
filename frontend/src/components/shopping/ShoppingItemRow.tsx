import { useState } from 'react'
import { Badge } from '../scouter'
import { PriceCell } from './PriceCell'
import { DealBadges } from './DealBadges'
import { PriceSparkline } from './PriceSparkline'
import { PricePredictionBadge } from './PricePredictionBadge'
import { PriceBenchmarkBadge } from './PriceBenchmarkBadge'
import { PriceAlertRules } from './PriceAlertRules'
import { NegotiationCoach } from './NegotiationCoach'
import { VATCalculator } from './VATCalculator'
import { PriceForecastCard } from './PriceForecastCard'
import { PriceAnalyticsPanel } from './PriceAnalyticsPanel'
import { ReviewSummaryCard } from './ReviewSummaryCard'
import { PurchaseAdvisorCard } from './PurchaseAdvisorCard'
import { RetailerLinks } from '../options/RetailerLinks'
import { MarketplaceComparator } from './MarketplaceComparator'
import { PriceHistoryExportButton } from './PriceHistoryExportButton'
import { CarbonBadge } from './CarbonBadge'
import { SeasonalBadge } from './SeasonalBadge'
import { WatchlistButton } from './WatchlistButton'
import { FrenchMarketInsight } from './FrenchMarketInsight'
import { PriceInsightsCard } from './PriceInsightsCard'
import { NegotiationSimulator } from './NegotiationSimulator'
import { CompetitorPriceTable } from './CompetitorPriceTable'
import { CouponFinderCard } from './CouponFinderCard'
import { ItemTagsDisplay } from './ItemTagsDisplay'
import { PriceStreakBadge } from './PriceStreakBadge'
import { TimingScoreBadge } from './TimingScoreBadge'
import { PriceFloorCard } from './PriceFloorCard'
import { ConditionPricingCard } from './ConditionPricingCard'
import { TargetSuggestionCard } from './TargetSuggestionCard'
import { DropPredictionCard } from './DropPredictionCard'
import { PriceAnnotationList } from './PriceAnnotationList'
import { VolatilityCalendar } from './VolatilityCalendar'
import { ElasticityCard } from './ElasticityCard'
import { NegotiationScriptCard } from './NegotiationScriptCard'
import { StockStatusBadge } from './StockStatusBadge'
import { PriceComparisonTable } from './PriceComparisonTable'
import { DealAggregatorCard } from './DealAggregatorCard'
import { MerchantRecommenderCard } from './MerchantRecommenderCard'
import { useDealScore } from '../../hooks'
import type { ShoppingItem, ItemStatus } from '../../types'
import type { RankedItem } from '../../api/listsorter'
import styles from './ShoppingItemRow.module.css'

// Separate component so hooks only fire when the panel is actually rendered,
// avoiding N+1 requests when many rows are displayed simultaneously.
function DealIntelPanel({ missionId, itemId }: { missionId: string; itemId: string }) {
  const { score } = useDealScore(missionId, itemId)
  return (
    <>
      <PriceSparkline missionId={missionId} itemId={itemId} />
      {score && <DealBadges itemId={itemId} dealScore={score} />}
    </>
  )
}

interface ShoppingItemRowProps {
  item: ShoppingItem
  missionId: string
  currency?: string
  onStatusChange?: (status: ItemStatus) => void
  onPriceClick?: () => void
  onPin?: (itemId: string) => void
  /** When set, shows a rank badge with reason tooltip */
  rank?: RankedItem
}

const STATUSES: ItemStatus[] = ['buy', 'watch', 'flash-sale', 'preorder', 'defer', 'crisis']

export function ShoppingItemRow({ item, missionId, currency = 'USD', onStatusChange, onPriceClick, onPin, rank }: ShoppingItemRowProps) {
  const [showIntel, setShowIntel] = useState(false)
  const [showRetailers, setShowRetailers] = useState(false)
  const [showVAT, setShowVAT] = useState(false)

  return (
    <>
      <div className={styles.row}>
      {/* Name + merchant */}
      <div className={styles.info}>
        <div className={styles.name}>
          {rank && (
            <span
              className={styles.rankBadge}
              title={rank.reason}
              aria-label={`Rang ${rank.rank} — ${rank.reason}`}
            >
              #{rank.rank}
            </span>
          )}
          {item.name}
        </div>
        <div className={styles.meta}>
          {item.merchant} · {item.costCategory}
          <button
            onClick={() => setShowRetailers((v) => !v)}
            className={styles.retailerToggleBtn}
            aria-expanded={showRetailers}
            aria-label={showRetailers ? 'Masquer les liens marchands' : 'Rechercher en ligne'}
          >
            {showRetailers ? '▲ liens' : '🛒 liens'}
          </button>
        </div>
        {showRetailers && (
          <>
            <RetailerLinks query={item.name} />
            <MarketplaceComparator itemName={item.name} />
          </>
        )}
        <ItemTagsDisplay missionId={missionId} itemId={item.id} />
        <StockStatusBadge missionId={missionId} itemId={item.id} />
      </div>

      {/* Price + delta + target */}
      <PriceCell item={item} missionId={missionId} currency={currency} onPriceClick={onPriceClick} />

      {/* VAT Calculator — only show toggle if item has a price */}
      {item.price > 0 && (
        <button
          onClick={() => setShowVAT((v) => !v)}
          className={styles.intelToggleBtn}
          title={showVAT ? 'Hide VAT calculator' : 'Show VAT calculator'}
          aria-label={showVAT ? 'Hide VAT calculator' : 'Show VAT calculator'}
          aria-expanded={showVAT}
        >
          €
        </button>
      )}

      {/* Deal intelligence — loaded on demand to avoid N+1 requests per row */}
      <div className={styles.dealIntel}>
        <button
          onClick={() => setShowIntel((v) => !v)}
          className={styles.intelToggleBtn}
          title={showIntel ? 'Hide deal intelligence' : 'Show deal intelligence'}
          aria-label={showIntel ? 'Hide deal intelligence' : 'Show deal intelligence'}
          aria-expanded={showIntel}
        >
          📊
        </button>
        {showIntel && <DealIntelPanel missionId={missionId} itemId={item.id} />}
        <PricePredictionBadge itemId={item.id} />
        <PriceBenchmarkBadge itemId={item.id} itemName={item.name} />
        <PriceAlertRules itemId={item.id} currentPrice={item.price} />
        <NegotiationCoach itemId={item.id} />
        <NegotiationSimulator missionId={missionId} itemId={item.id} />
        <PriceHistoryExportButton missionId={missionId} itemId={item.id} itemName={item.name} />
        <CarbonBadge itemName={item.name} category={item.costCategory} />
        <SeasonalBadge itemName={item.name} category={item.costCategory} />
        <WatchlistButton itemId={item.id} itemName={item.name} currentPrice={item.price} />
        <PriceStreakBadge missionId={missionId} itemId={item.id} />
        <TimingScoreBadge missionId={missionId} itemId={item.id} />
      </div>

      <Badge variant={item.status} />

      {onPin && (
        <button
          onClick={() => onPin(item.id)}
          title={item.pinned ? 'Pinned — will survive re-run' : 'Pin this item'}
          className={`${styles.pinBtn} ${item.pinned ? styles.pinned : styles.unpinned}`}
        >
          📌
        </button>
      )}

      {onStatusChange && (
        <select
          value={item.status}
          onChange={(e) => onStatusChange(e.target.value as ItemStatus)}
          className={styles.statusSelect}
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      )}
      </div>
      {showVAT && (
        <VATCalculator priceHT={item.price} onClose={() => setShowVAT(false)} />
      )}
      {showIntel && (
        <PriceForecastCard missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <PriceAnalyticsPanel missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <PurchaseAdvisorCard itemId={item.id} itemName={item.name} />
      )}
      {showIntel && (
        <ReviewSummaryCard itemId={item.id} itemName={item.name} />
      )}
      {showIntel && (
        <FrenchMarketInsight itemName={item.name} category={item.costCategory} />
      )}
      {showIntel && (
        <PriceInsightsCard missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <CompetitorPriceTable missionId={missionId} itemId={item.id} />
      )}
      {showIntel && (
        <CouponFinderCard missionId={missionId} itemId={item.id} />
      )}
      {showIntel && (
        <PriceFloorCard missionId={missionId} itemId={item.id} />
      )}
      {showIntel && (
        <ConditionPricingCard missionId={missionId} itemId={item.id} />
      )}
      {showIntel && (
        <TargetSuggestionCard missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <DropPredictionCard missionId={missionId} itemId={item.id} />
      )}
      {showIntel && (
        <PriceAnnotationList missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <VolatilityCalendar missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <ElasticityCard missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <NegotiationScriptCard missionId={missionId} itemId={item.id} />
      )}
      {showIntel && (
        <PriceComparisonTable missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <DealAggregatorCard missionId={missionId} itemId={item.id} currency={currency} />
      )}
      {showIntel && (
        <MerchantRecommenderCard missionId={missionId} itemId={item.id} />
      )}
    </>
  )
}

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { BarChart2, Pin, ShoppingCart } from 'lucide-react'
import { Badge } from '../scouter'
import { PriceCell } from './PriceCell'
import { DealBadges } from './DealBadges'
import { PriceSparkline } from './PriceSparkline'
import { PricePredictionBadge } from './PricePredictionBadge'
import { PriceBenchmarkBadge } from './PriceBenchmarkBadge'
import { PriceAlertRules } from './PriceAlertRules'
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
import { NegotiationCoach } from './NegotiationCoach'
import { StockStatusBadge } from './StockStatusBadge'
import { PriceComparisonTable } from './PriceComparisonTable'
import { DealAggregatorCard } from './DealAggregatorCard'
import { MerchantRecommenderCard } from './MerchantRecommenderCard'
import { useDealScore } from '../../hooks'
import type { ShoppingItem, ItemStatus } from '../../types'
import type { RankedItem } from '../../api/listsorter'
import styles from './ShoppingItemRow.module.css'

type IntelTab = 'forecast' | 'analytics' | 'negotiate' | 'market' | 'reviews'

// Loaded on demand to avoid N+1 requests when many rows are displayed simultaneously.
function IntelTabPanel({
  missionId, itemId, currency, itemName, category, activeTab, onTabChange,
}: {
  missionId: string
  itemId: string
  currency: string
  itemName: string
  category: string
  activeTab: IntelTab
  onTabChange: (tab: IntelTab) => void
}) {
  const { t } = useTranslation()
  const { score } = useDealScore(missionId, itemId)

  const tabs: { id: IntelTab; label: string }[] = [
    { id: 'forecast', label: t('shopping.tabForecast') },
    { id: 'analytics', label: t('shopping.tabAnalytics') },
    { id: 'negotiate', label: t('shopping.tabNegotiate') },
    { id: 'market', label: t('shopping.tabMarket') },
    { id: 'reviews', label: t('shopping.tabReviews') },
  ]

  return (
    <div className={styles.intelPanel}>
      <div className={styles.tabBar} role="tablist" aria-label={t('shopping.intelPanelLabel')}>
        {tabs.map((tab) => (
          <button
            key={tab.id}
            role="tab"
            aria-selected={activeTab === tab.id}
            className={`${styles.tabBtn} ${activeTab === tab.id ? styles.tabBtnActive : ''}`}
            onClick={() => onTabChange(tab.id)}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className={styles.tabContent} role="tabpanel" aria-label={activeTab}>
        {activeTab === 'forecast' && (
          <>
            <PriceForecastCard missionId={missionId} itemId={itemId} currency={currency} />
            <DropPredictionCard missionId={missionId} itemId={itemId} />
            <TargetSuggestionCard missionId={missionId} itemId={itemId} currency={currency} />
            <PriceFloorCard missionId={missionId} itemId={itemId} />
          </>
        )}
        {activeTab === 'analytics' && (
          <>
            <PriceSparkline missionId={missionId} itemId={itemId} />
            {score && <DealBadges itemId={itemId} dealScore={score} />}
            <PriceAnalyticsPanel missionId={missionId} itemId={itemId} currency={currency} />
            <PriceInsightsCard missionId={missionId} itemId={itemId} currency={currency} />
            <PriceAnnotationList missionId={missionId} itemId={itemId} currency={currency} />
            <VolatilityCalendar missionId={missionId} itemId={itemId} currency={currency} />
            <ElasticityCard missionId={missionId} itemId={itemId} currency={currency} />
          </>
        )}
        {activeTab === 'negotiate' && (
          <>
            <NegotiationCoach itemId={itemId} />
            <NegotiationSimulator missionId={missionId} itemId={itemId} />
            <CouponFinderCard missionId={missionId} itemId={itemId} />
            <NegotiationScriptCard missionId={missionId} itemId={itemId} />
          </>
        )}
        {activeTab === 'market' && (
          <>
            <FrenchMarketInsight itemName={itemName} category={category} />
            <CompetitorPriceTable missionId={missionId} itemId={itemId} />
            <PriceComparisonTable missionId={missionId} itemId={itemId} currency={currency} />
            <DealAggregatorCard missionId={missionId} itemId={itemId} currency={currency} />
            <MerchantRecommenderCard missionId={missionId} itemId={itemId} />
          </>
        )}
        {activeTab === 'reviews' && (
          <>
            <ReviewSummaryCard itemId={itemId} itemName={itemName} />
            <PurchaseAdvisorCard itemId={itemId} itemName={itemName} />
            <ConditionPricingCard missionId={missionId} itemId={itemId} />
          </>
        )}
      </div>
    </div>
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
  const [activeTab, setActiveTab] = useState<IntelTab>('forecast')

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
              <ShoppingCart size={11} strokeWidth={2} />
              {' liens'}
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

        {/* Deal intelligence — tab panel loaded on demand */}
        <div className={styles.dealIntel}>
          <button
            onClick={() => setShowIntel((v) => !v)}
            className={styles.intelToggleBtn}
            title={showIntel ? 'Hide deal intelligence' : 'Show deal intelligence'}
            aria-label={showIntel ? 'Hide deal intelligence' : 'Show deal intelligence'}
            aria-expanded={showIntel}
          >
            <BarChart2 size={14} />
          </button>
          <PricePredictionBadge itemId={item.id} />
          <PriceBenchmarkBadge itemId={item.id} itemName={item.name} />
          <PriceAlertRules itemId={item.id} currentPrice={item.price} />
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
            <Pin size={14} />
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
        <IntelTabPanel
          missionId={missionId}
          itemId={item.id}
          currency={currency}
          itemName={item.name}
          category={item.costCategory}
          activeTab={activeTab}
          onTabChange={setActiveTab}
        />
      )}
    </>
  )
}

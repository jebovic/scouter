import { useRef, useState } from 'react'
import { Badge } from '../scouter'
import { TrendBadge } from './TrendBadge'
import { DealScoreBadge } from './DealScoreBadge'
import { DealExplainerPopover } from './DealExplainerPopover'
import { PriceSparkline } from './PriceSparkline'
import { PricePredictionBadge } from './PricePredictionBadge'
import { PriceAlertBadge } from './PriceAlertBadge'
import { PriceBenchmarkBadge } from './PriceBenchmarkBadge'
import { VATCalculator } from './VATCalculator'
import { RetailerLinks } from '../options/RetailerLinks'
import { useDealScore, useUpdateShoppingItem } from '../../hooks'
import { formatCurrency } from '../../utils/format'
import type { ShoppingItem, ItemStatus } from '../../types'
import styles from './ShoppingItemRow.module.css'

// Separate component so hooks only fire when the panel is actually rendered,
// avoiding N+1 requests when many rows are displayed simultaneously.
function DealIntelPanel({ missionId, itemId }: { missionId: string; itemId: string }) {
  const { score } = useDealScore(missionId, itemId)
  return (
    <>
      <PriceSparkline missionId={missionId} itemId={itemId} />
      {score && (
        <>
          <TrendBadge trend={score.trend} />
          <DealScoreBadge score={score} />
          <DealExplainerPopover
            itemId={itemId}
            dealScore={Math.round(Math.max(0, Math.min(100, score.pctBelowAvg)))}
            trend={score.trend}
          />
        </>
      )}
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
}

const STATUSES: ItemStatus[] = ['buy', 'watch', 'flash-sale', 'preorder', 'defer', 'crisis']

export function ShoppingItemRow({ item, missionId, currency = 'USD', onStatusChange, onPriceClick, onPin }: ShoppingItemRowProps) {
  const [editingTarget, setEditingTarget] = useState(false)
  const [targetInput, setTargetInput] = useState('')
  const [showIntel, setShowIntel] = useState(false)
  const [showRetailers, setShowRetailers] = useState(false)
  const [showVAT, setShowVAT] = useState(false)
  const cancelledRef = useRef(false)

  const { updateItem } = useUpdateShoppingItem(missionId)

  const priceDelta = item.originalEstimate != null
    ? item.price - item.originalEstimate
    : null

  const fmt = (n: number) => formatCurrency(n, currency)

  function handleTargetSubmit() {
    if (cancelledRef.current) {
      cancelledRef.current = false
      return
    }
    const val = parseFloat(targetInput)
    if (!isNaN(val) && val > 0) {
      updateItem({ itemId: item.id, req: { targetPrice: val } })
    }
    setEditingTarget(false)
  }

  function handleTargetKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') handleTargetSubmit()
    if (e.key === 'Escape') {
      cancelledRef.current = true
      setEditingTarget(false)
    }
  }

  return (
    <>
      <div className={styles.row}>
      {/* Name + merchant */}
      <div className={styles.info}>
        <div className={styles.name}>{item.name}</div>
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
          <RetailerLinks query={item.name} />
        )}
      </div>

      {/* Price + delta + target */}
      <div className={styles.priceCell}>
        <button
          onClick={onPriceClick}
          className={styles.priceBtn}
          aria-label={`View price history for ${item.name}: ${fmt(item.price)}`}
          disabled={!onPriceClick}
        >
          {fmt(item.price)}
        </button>
        {priceDelta !== null && (
          <div className={`${styles.priceDelta} ${priceDelta > 0 ? styles.priceDeltaUp : styles.priceDeltaDown}`}>
            {priceDelta > 0 ? '+' : ''}{fmt(priceDelta)}
          </div>
        )}
        {/* Price Alert Badge */}
        <PriceAlertBadge currentPrice={item.price} targetPrice={item.targetPrice} currency={currency} />
        {/* Target price */}
        {editingTarget ? (
          <input
            autoFocus
            type="number"
            value={targetInput}
            onChange={(e) => setTargetInput(e.target.value)}
            onBlur={handleTargetSubmit}
            onKeyDown={handleTargetKeyDown}
            placeholder="target"
            className={styles.targetInput}
          />
        ) : (
          <button
            onClick={() => { setTargetInput(item.targetPrice != null ? String(item.targetPrice) : ''); setEditingTarget(true) }}
            title="Set target price"
            className={`${styles.targetBtn} ${item.targetPrice != null ? styles.targetSet : styles.targetUnset}`}
          >
            {item.targetPrice != null ? `↯ ${fmt(item.targetPrice)}` : '+ target'}
          </button>
        )}
      </div>

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
    </>
  )
}

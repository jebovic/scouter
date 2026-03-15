import { useRef, useState } from 'react'
import { VoteBar } from './VoteBar'
import { CurrencyConverter } from './CurrencyConverter'
import { LoyaltyCalculator } from './LoyaltyCalculator'
import { PriceAlertBadge } from './PriceAlertBadge'
import { useUpdateShoppingItem } from '../../hooks'
import { formatCurrency } from '../../utils/format'
import type { ShoppingItem } from '../../types'
import styles from './ShoppingItemRow.module.css'

interface PriceCellProps {
  item: ShoppingItem
  missionId: string
  currency: string
  onPriceClick?: () => void
}

export function PriceCell({ item, missionId, currency, onPriceClick }: PriceCellProps) {
  const [editingTarget, setEditingTarget] = useState(false)
  const [targetInput, setTargetInput] = useState('')
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
      {/* Currency Converter */}
      {item.price > 0 && (
        <CurrencyConverter basePrice={item.price} baseCurrency={currency} />
      )}
      {/* Loyalty Points & Cashback Calculator */}
      {item.price > 0 && (
        <LoyaltyCalculator basePrice={item.price} merchant={item.merchant} />
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
      {/* Collaborative voting bar */}
      <VoteBar optionId={item.id} />
    </div>
  )
}

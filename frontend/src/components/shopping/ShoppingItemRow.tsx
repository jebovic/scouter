import { useRef, useState } from 'react'
import { Badge } from '../scouter'
import { TrendBadge } from './TrendBadge'
import { DealScoreBadge } from './DealScoreBadge'
import { PriceSparkline } from './PriceSparkline'
import { useDealScore, useUpdateShoppingItem } from '../../hooks'
import { formatCurrency } from '../../utils/format'
import type { ShoppingItem, ItemStatus } from '../../types'
import styles from './ShoppingItemRow.module.css'

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
  const cancelledRef = useRef(false)

  const { score } = useDealScore(missionId, item.id)
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
    <div className={styles.row}>
      {/* Name + merchant */}
      <div className={styles.info}>
        <div className={styles.name}>{item.name}</div>
        <div className={styles.meta}>
          {item.merchant} · {item.costCategory}
        </div>
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

      {/* Deal intelligence badges */}
      <div className={styles.dealIntel}>
        <PriceSparkline missionId={missionId} itemId={item.id} />
        {score && (
          <>
            <TrendBadge trend={score.trend} />
            <DealScoreBadge score={score} />
          </>
        )}
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
  )
}

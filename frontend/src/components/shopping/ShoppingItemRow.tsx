import { useState } from 'react'
import { Badge } from '../scouter'
import { TrendBadge } from './TrendBadge'
import { DealScoreBadge } from './DealScoreBadge'
import { PriceSparkline } from './PriceSparkline'
import { useDealScore, useUpdateShoppingItem } from '../../hooks'
import type { ShoppingItem, ItemStatus } from '../../types'

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

  const { score } = useDealScore(missionId, item.id)
  const { updateItem } = useUpdateShoppingItem(missionId)

  const priceDelta = item.originalEstimate != null
    ? item.price - item.originalEstimate
    : null

  const fmt = (n: number) =>
    new Intl.NumberFormat(undefined, { style: 'currency', currency, maximumFractionDigits: 0 }).format(n)

  function handleTargetSubmit() {
    const val = parseFloat(targetInput)
    if (!isNaN(val) && val > 0) {
      updateItem({ itemId: item.id, req: { targetPrice: val } })
    }
    setEditingTarget(false)
  }

  function handleTargetKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') handleTargetSubmit()
    if (e.key === 'Escape') setEditingTarget(false)
  }

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: '1fr auto auto auto auto auto',
        alignItems: 'center',
        gap: 12,
        padding: '10px 16px',
        borderBottom: '1px solid var(--border)',
        transition: 'background 0.15s',
        opacity: item.pinned ? 1 : undefined,
      }}
    >
      {/* Name + merchant */}
      <div>
        <div style={{ fontSize: '0.875rem', color: 'var(--text)' }}>{item.name}</div>
        <div style={{ fontSize: '0.7rem', color: 'var(--text-dim)', marginTop: 2 }}>
          {item.merchant} · {item.costCategory}
        </div>
      </div>

      {/* Price + delta + target */}
      <div style={{ textAlign: 'right', minWidth: 80 }}>
        <button
          onClick={onPriceClick}
          style={{
            background: 'none',
            border: 'none',
            cursor: onPriceClick ? 'pointer' : 'default',
            padding: 0,
            fontFamily: 'var(--font-mono)',
            fontSize: '0.9rem',
            color: 'var(--cyan)',
            fontWeight: 600,
          }}
        >
          {fmt(item.price)}
        </button>
        {priceDelta !== null && (
          <div
            style={{
              fontSize: '0.65rem',
              color: priceDelta > 0 ? 'var(--coral)' : 'var(--green)',
              fontFamily: 'var(--font-mono)',
            }}
          >
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
            style={{
              width: 72,
              background: 'var(--raised)',
              border: '1px solid var(--cyan)',
              borderRadius: 4,
              color: 'var(--text)',
              padding: '2px 4px',
              fontSize: '0.65rem',
              fontFamily: 'var(--font-mono)',
              marginTop: 2,
            }}
          />
        ) : (
          <button
            onClick={() => { setTargetInput(item.targetPrice != null ? String(item.targetPrice) : ''); setEditingTarget(true) }}
            title="Set target price"
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              padding: 0,
              fontSize: '0.65rem',
              fontFamily: 'var(--font-mono)',
              color: item.targetPrice != null ? 'var(--gold)' : 'var(--text-dim)',
              marginTop: 2,
              display: 'block',
            }}
          >
            {item.targetPrice != null ? `↯ ${fmt(item.targetPrice)}` : '+ target'}
          </button>
        )}
      </div>

      {/* Deal intelligence badges */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4, alignItems: 'flex-end' }}>
        <PriceSparkline missionId={missionId} itemId={item.id} />
        {score?.trend && <TrendBadge trend={score.trend} />}
        {score && <DealScoreBadge score={score} />}
      </div>

      <Badge variant={item.status} />

      {onPin && (
        <button
          onClick={() => onPin(item.id)}
          title={item.pinned ? 'Pinned — will survive re-run' : 'Pin this item'}
          style={{
            background: 'none',
            border: `1px solid ${item.pinned ? 'var(--cyan)' : 'var(--border)'}`,
            borderRadius: 4,
            color: item.pinned ? 'var(--cyan)' : 'var(--text-dim)',
            cursor: 'pointer',
            padding: '2px 6px',
            fontSize: '0.7rem',
            fontFamily: 'var(--font-mono)',
            transition: 'all 0.15s',
          }}
        >
          📌
        </button>
      )}

      {onStatusChange && (
        <select
          value={item.status}
          onChange={(e) => onStatusChange(e.target.value as ItemStatus)}
          style={{
            background: 'var(--raised)',
            border: '1px solid var(--border)',
            borderRadius: 4,
            color: 'var(--text-mid)',
            padding: '4px 6px',
            fontSize: '0.7rem',
            fontFamily: 'var(--font-mono)',
            cursor: 'pointer',
          }}
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      )}
    </div>
  )
}

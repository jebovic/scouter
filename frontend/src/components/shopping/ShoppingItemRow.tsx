import { Badge } from '../scouter'
import type { ShoppingItem, ItemStatus } from '../../types'

interface ShoppingItemRowProps {
  item: ShoppingItem
  currency?: string
  onStatusChange?: (status: ItemStatus) => void
  onPriceClick?: () => void
  onPin?: (itemId: string) => void
}

const STATUSES: ItemStatus[] = ['buy', 'watch', 'flash-sale', 'preorder', 'defer', 'crisis']

export function ShoppingItemRow({ item, currency = 'USD', onStatusChange, onPriceClick, onPin }: ShoppingItemRowProps) {
  const priceDelta = item.originalEstimate != null
    ? item.price - item.originalEstimate
    : null

  const fmt = (n: number) =>
    new Intl.NumberFormat(undefined, { style: 'currency', currency, maximumFractionDigits: 0 }).format(n)

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: '1fr auto auto auto auto',
        alignItems: 'center',
        gap: 12,
        padding: '10px 16px',
        borderBottom: '1px solid var(--border)',
        transition: 'background 0.15s',
        opacity: item.pinned ? 1 : undefined,
      }}
    >
      <div>
        <div style={{ fontSize: '0.875rem', color: 'var(--text)' }}>{item.name}</div>
        <div style={{ fontSize: '0.7rem', color: 'var(--text-dim)', marginTop: 2 }}>
          {item.merchant} · {item.costCategory}
        </div>
      </div>

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

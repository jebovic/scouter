import { ShoppingItemRow } from './ShoppingItemRow'
import type { ShoppingItem, ItemStatus } from '../../types'

interface MerchantGroupProps {
  merchant: string
  items: ShoppingItem[]
  currency?: string
  onStatusChange?: (itemId: string, status: ItemStatus) => void
  onPriceClick?: (item: ShoppingItem) => void
}

export function MerchantGroup({ merchant, items, currency, onStatusChange, onPriceClick }: MerchantGroupProps) {
  const total = items.reduce((sum, i) => sum + i.price, 0)
  const fmt = (n: number) =>
    new Intl.NumberFormat(undefined, { style: 'currency', currency, maximumFractionDigits: 0 }).format(n)

  return (
    <div
      style={{
        background: 'var(--surface)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--radius)',
        overflow: 'hidden',
        marginBottom: 12,
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          padding: '8px 16px',
          background: 'var(--raised)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <span
          style={{
            fontFamily: 'var(--font-display)',
            fontSize: '0.8rem',
            color: 'var(--text-mid)',
            letterSpacing: '0.05em',
          }}
        >
          {merchant}
        </span>
        <span
          style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '0.8rem',
            color: 'var(--gold)',
          }}
        >
          {fmt(total)}
        </span>
      </div>
      {items.map((item) => (
        <ShoppingItemRow
          key={item.id}
          item={item}
          currency={currency}
          onStatusChange={onStatusChange ? (s) => onStatusChange(item.id, s) : undefined}
          onPriceClick={onPriceClick ? () => onPriceClick(item) : undefined}
        />
      ))}
    </div>
  )
}

import styles from './StockBadge.module.css'

export type StockBadgeStatus = 'in_stock' | 'limited' | 'out_of_stock' | 'loading'

interface StockBadgeProps {
  status: StockBadgeStatus
}

const LABEL: Record<StockBadgeStatus, string> = {
  in_stock: '✓ En stock',
  limited: '⚡ Stock limité',
  out_of_stock: '✗ Rupture',
  loading: '',
}

const CLASS: Record<StockBadgeStatus, string> = {
  in_stock: styles.inStock,
  limited: styles.limited,
  out_of_stock: styles.outOfStock,
  loading: styles.loading,
}

export function StockBadge({ status }: StockBadgeProps) {
  return (
    <span
      role="status"
      aria-label={status === 'loading' ? 'Chargement du stock' : LABEL[status]}
      className={`${styles.badge} ${CLASS[status]}`}
    >
      {LABEL[status]}
    </span>
  )
}

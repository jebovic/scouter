import { useStockStatus } from '../../hooks/useStockStatus'
import styles from './StockStatusBadge.module.css'

interface StockStatusBadgeProps {
  missionId: string
  itemId: string
}

export function StockStatusBadge({ missionId, itemId }: StockStatusBadgeProps) {
  const { data, isLoading } = useStockStatus(missionId, itemId)

  if (isLoading || !data) return null

  const statusClass = {
    rupture: styles.rupture,
    stock_faible: styles.stock_faible,
    disponible: styles.disponible,
    'stock_élevé': styles.stock_élevé,
  }[data.stockStatus] ?? styles.disponible

  return (
    <div className={styles.wrapper}>
      <span className={`${styles.badge} ${statusClass}`}>
        {data.stockLabel}
      </span>
      {data.restockDays != null && (
        <span className={styles.restockNote}>
          Réappro dans ~{data.restockDays} j
        </span>
      )}
      <div className={styles.tooltip}>
        <div className={styles.tooltipAlert}>{data.alert}</div>
      </div>
    </div>
  )
}

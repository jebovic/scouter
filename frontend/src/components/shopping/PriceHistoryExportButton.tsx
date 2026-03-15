import { getPriceHistoryExportURL } from '../../api/shopping'
import styles from './PriceHistoryExportButton.module.css'

interface PriceHistoryExportButtonProps {
  missionId: string
  itemId: string
  itemName: string
}

export function PriceHistoryExportButton({ missionId, itemId, itemName }: PriceHistoryExportButtonProps) {
  const exportURL = getPriceHistoryExportURL(missionId, itemId)

  return (
    <a
      href={exportURL}
      download
      className={styles.button}
      title={`Export price history for ${itemName} as CSV`}
      aria-label={`Export price history for ${itemName}`}
    >
      📥 Exporter CSV
    </a>
  )
}

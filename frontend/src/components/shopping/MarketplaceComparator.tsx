import { useMarketplace } from '../../hooks/useMarketplace'
import styles from './MarketplaceComparator.module.css'

interface MarketplaceComparatorProps {
  itemName: string
}

export function MarketplaceComparator({ itemName }: MarketplaceComparatorProps) {
  const { data, isLoading } = useMarketplace(itemName)

  if (isLoading) return <div className={styles.loading}>🔍 Recherche…</div>
  if (!data) return null

  return (
    <div className={styles.wrap}>
      <div className={styles.title}>Comparer sur</div>
      <div className={styles.grid}>
        {data.results.map((r) => (
          <a
            key={r.site}
            href={r.url}
            target="_blank"
            rel="noopener noreferrer"
            className={styles.siteLink}
          >
            <span className={styles.siteName}>{r.site}</span>
            <span className={styles.siteLabel}>{r.label}</span>
          </a>
        ))}
      </div>
    </div>
  )
}

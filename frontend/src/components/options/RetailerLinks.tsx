import { buildRetailerLinks } from '../../utils/retailerLinks'
import { useStockCheck } from '../../hooks/useStockCheck'
import { StockBadge } from './StockBadge'
import type { StockBadgeStatus } from './StockBadge'
import styles from './RetailerLinks.module.css'

interface RetailerLinksProps {
  query: string
  category?: string
  showStock?: boolean
}

interface RetailerPillProps {
  query: string
  retailerId: string
  retailerName: string
  retailerLogo: string
  url: string
  showStock: boolean
}

function RetailerPill({ query, retailerId, retailerName, retailerLogo, url, showStock }: RetailerPillProps) {
  const { data, isLoading } = useStockCheck(showStock ? query : '', showStock ? retailerId : '')

  const badgeStatus: StockBadgeStatus = isLoading
    ? 'loading'
    : (data?.status ?? 'loading')

  return (
    <div className={styles.pillWrapper}>
      <a
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.pill}
        title={`Search "${query}" on ${retailerName}`}
      >
        <span className={styles.logo} aria-hidden="true">
          {retailerLogo}
        </span>
        {retailerName}
      </a>
      {showStock && <StockBadge status={badgeStatus} />}
    </div>
  )
}

export function RetailerLinks({ query, category, showStock = false }: RetailerLinksProps) {
  const links = buildRetailerLinks(query, category)

  if (links.length === 0 || !query.trim()) return null

  return (
    <div className={styles.container} aria-label="Search on French retailers">
      {links.map(({ retailer, url }) => (
        <RetailerPill
          key={retailer.id}
          query={query}
          retailerId={retailer.id}
          retailerName={retailer.name}
          retailerLogo={retailer.logo}
          url={url}
          showStock={showStock}
        />
      ))}
    </div>
  )
}

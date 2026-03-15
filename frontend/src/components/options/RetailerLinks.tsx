import { buildRetailerLinks } from '../../utils/retailerLinks'
import styles from './RetailerLinks.module.css'

interface RetailerLinksProps {
  query: string
  category?: string
}

export function RetailerLinks({ query, category }: RetailerLinksProps) {
  const links = buildRetailerLinks(query, category)

  if (links.length === 0 || !query.trim()) return null

  return (
    <div className={styles.container} aria-label="Search on French retailers">
      {links.map(({ retailer, url }) => (
        <a
          key={retailer.id}
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className={styles.pill}
          title={`Search "${query}" on ${retailer.name}`}
        >
          <span className={styles.logo} aria-hidden="true">
            {retailer.logo}
          </span>
          {retailer.name}
        </a>
      ))}
    </div>
  )
}

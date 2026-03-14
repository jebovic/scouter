import { Link } from 'react-router-dom'
import styles from './Breadcrumb.module.css'

interface BreadcrumbProps {
  missionName?: string
  missionSlug?: string
  subPage?: string
}

export function Breadcrumb({ missionName, missionSlug, subPage }: BreadcrumbProps) {
  return (
    <nav aria-label="breadcrumb" className={styles.breadcrumb}>
      <Link to="/" className={styles.crumb}>
        HQ
      </Link>
      {missionName && (
        <>
          <span className={styles.sep}>/</span>
          {subPage && missionSlug ? (
            <Link to={`/missions/${missionSlug}`} className={styles.crumb}>
              {missionName}
            </Link>
          ) : (
            <span className={styles.current}>{missionName}</span>
          )}
        </>
      )}
      {subPage && (
        <>
          <span className={styles.sep}>/</span>
          <span className={styles.current}>{subPage}</span>
        </>
      )}
    </nav>
  )
}

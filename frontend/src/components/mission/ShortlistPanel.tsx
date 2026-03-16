import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import type { Option } from '../../types'
import styles from './ShortlistPanel.module.css'

interface ShortlistPanelProps {
  options: Option[]
  missionSlug: string
  onSelect: (option: Option) => void
}

export function ShortlistPanel({ options, missionSlug, onSelect }: ShortlistPanelProps) {
  const { t } = useTranslation()
  const pinned = options.filter((o) => o.pinned)

  return (
    <div className={styles.panel}>
      <h3 className={styles.title}>{t('shortlist.title')}</h3>

      {pinned.length === 0 ? (
        <div className={styles.empty}>
          <p>{t('shortlist.empty')}</p>
          <Link to={`/missions/${missionSlug}/options`} className={styles.emptyLink}>
            {t('shortlist.emptyLink')}
          </Link>
        </div>
      ) : (
        <div className={styles.list}>
          {pinned.map((option) => (
            <div key={option.id} className={styles.card}>
              <div className={styles.cardInfo}>
                <span className={styles.optionName}>{option.name}</span>
                {option.priceRange && (
                  <span className={styles.priceRange}>
                    {option.priceRange.min}–{option.priceRange.max}
                  </span>
                )}
                {option.badge && (
                  <span className={`${styles.badge} ${styles[option.badge]}`}>
                    {t(`options.badge.${option.badge}`)}
                  </span>
                )}
              </div>
              <button
                className={styles.selectBtn}
                onClick={() => onSelect(option)}
              >
                {t('shortlist.select')}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

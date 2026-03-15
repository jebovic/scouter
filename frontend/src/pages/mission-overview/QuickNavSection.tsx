import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import type { Mission } from '../../types'
import styles from '../MissionOverview.module.css'

interface QuickNavSectionProps {
  mission: Mission
}

export function QuickNavSection({ mission }: QuickNavSectionProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const slug = mission.slug
  return (
    <div className={styles.quickNav}>
      <button
        onClick={() => navigate(`/missions/${slug}/options`)}
        className={styles.quickNavBtn}
      >
        <div className={styles.quickNavIcon}>🔍</div>
        <div className={`${styles.quickNavTitle} ${styles.quickNavTitleOptions}`}>{t('quickNav.optionsExplorer')}</div>
        <div className={styles.quickNavDesc}>{t('quickNav.optionsDesc')}</div>
      </button>
      <button
        onClick={() => navigate(`/missions/${slug}/shopping`)}
        className={styles.quickNavBtn}
      >
        <div className={styles.quickNavIcon}>🛒</div>
        <div className={`${styles.quickNavTitle} ${styles.quickNavTitleShopping}`}>{t('quickNav.shoppingTracker')}</div>
        <div className={styles.quickNavDesc}>{t('quickNav.shoppingDesc')}</div>
      </button>
    </div>
  )
}

import { useTranslation } from 'react-i18next'
import { Topnav } from '../components/scouter/Topnav'
import { LoyaltyTracker } from '../components/scouter/LoyaltyTracker'
import styles from './LoyaltyPage.module.css'

export default function LoyaltyPage() {
  const { t } = useTranslation()
  return (
    <>
      <Topnav />
      <main className={`page ${styles.page}`}>
        <div className={styles.header}>
          <h1 className={styles.title}>💎 {t('loyalty.title')}</h1>
          <p className={styles.subtitle}>{t('loyalty.subtitle')}</p>
        </div>
        <LoyaltyTracker />
      </main>
    </>
  )
}

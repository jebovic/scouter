import { Topnav } from '../components/scouter/Topnav'
import { LoyaltyTracker } from '../components/scouter/LoyaltyTracker'
import styles from './LoyaltyPage.module.css'

export default function LoyaltyPage() {
  return (
    <>
      <Topnav />
      <main className={`page ${styles.page}`}>
        <div className={styles.header}>
          <h1 className={styles.title}>💎 Mes Points Fidélité</h1>
          <p className={styles.subtitle}>Suivi de vos programmes de fidélité</p>
        </div>
        <LoyaltyTracker />
      </main>
    </>
  )
}

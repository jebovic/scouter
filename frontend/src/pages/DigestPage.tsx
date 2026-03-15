import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Topnav } from '../components/scouter'
import { useWeeklyDigest } from '../hooks/useDigest'
import type { PriceDropItem } from '../api/digest'
import styles from './DigestPage.module.css'

const PERIOD_OPTIONS: { label: string; days: number }[] = [
  { label: '7 jours', days: 7 },
  { label: '14 jours', days: 14 },
  { label: '30 jours', days: 30 },
]

function formatPrice(amount: number, currency: string): string {
  return new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount)
}

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('fr-FR', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(iso))
}

function DropRow({ item }: { item: PriceDropItem }) {
  const currency = item.currency || 'EUR'
  return (
    <div className={styles.row}>
      <div className={styles.rowLeft}>
        <Link
          to={`/missions/${item.missionSlug}/shopping`}
          className={styles.missionLink}
        >
          {item.missionName}
        </Link>
        <span className={styles.itemName}>{item.itemName}</span>
        <span className={styles.merchant}>{item.merchant}</span>
      </div>

      <div className={styles.rowRight}>
        <div className={styles.priceGroup}>
          <span className={styles.priceBefore}>
            {formatPrice(item.priceBefore, currency)}
          </span>
          <span className={styles.arrow}>↓</span>
          <span className={styles.priceNow}>
            {formatPrice(item.priceNow, currency)}
          </span>
        </div>
        <span className={styles.dropBadge}>
          -{item.dropPct.toFixed(1)}%
        </span>
      </div>
    </div>
  )
}

export default function DigestPage() {
  const [days, setDays] = useState(7)
  const { digest, isLoading } = useWeeklyDigest(days)

  const currency = digest?.currency ?? 'EUR'

  return (
    <>
      <Topnav />
      <main className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Résumé de la semaine</h1>
          <p className={styles.subtitle}>
            Baisses de prix détectées sur vos missions actives
            {digest && (
              <> — généré le {formatDate(digest.generatedAt)}</>
            )}
          </p>
        </div>

        {/* Period selector */}
        <div className={styles.periodTabs}>
          {PERIOD_OPTIONS.map(({ label, days: d }) => (
            <button
              key={d}
              className={
                days === d
                  ? `${styles.tabBtn} ${styles.tabBtnActive}`
                  : styles.tabBtn
              }
              onClick={() => setDays(d)}
            >
              {label}
            </button>
          ))}
        </div>

        {/* Summary bar */}
        {!isLoading && digest && (
          <div className={styles.summaryBar}>
            <div className={styles.summaryMetric}>
              <span className={styles.summaryLabel}>Baisses détectées</span>
              <span className={styles.summaryValue}>{digest.totalDrops}</span>
            </div>
            <div className={styles.summaryMetric}>
              <span className={styles.summaryLabel}>Économies totales</span>
              <span className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>
                {formatPrice(digest.totalSavings, currency)}
              </span>
            </div>
          </div>
        )}

        {/* Content */}
        {isLoading ? (
          <div className={styles.list}>
            {[1, 2, 3, 4].map((k) => (
              <div key={k} className={styles.skeletonRow} />
            ))}
          </div>
        ) : digest && digest.items.length === 0 ? (
          <div className={styles.empty}>
            <div className={styles.emptyIcon}>🎉</div>
            <div className={styles.emptyTitle}>
              Aucune baisse de prix cette semaine
            </div>
          </div>
        ) : (
          <div className={styles.list}>
            {(digest?.items ?? []).map((item) => (
              <DropRow key={item.itemId} item={item} />
            ))}
          </div>
        )}
      </main>
    </>
  )
}

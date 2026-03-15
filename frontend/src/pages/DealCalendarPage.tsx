import { useState } from 'react'
import { Topnav } from '../components/scouter/Topnav'
import { useDealCalendar } from '../hooks/useDealCalendar'
import type { DealEvent } from '../api/dealCalendar'
import styles from './DealCalendarPage.module.css'

const CATEGORY_FILTERS: { label: string; value: string }[] = [
  { label: 'Tous', value: 'all' },
  { label: 'Électronique', value: 'electronique' },
  { label: 'Informatique', value: 'informatique' },
  { label: 'Mode', value: 'mode' },
  { label: 'Voyage', value: 'voyage' },
]

function formatDateRange(startIso: string, endIso: string): string {
  const fmt = new Intl.DateTimeFormat('fr-FR', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
  const start = new Date(startIso)
  const end = new Date(endIso)

  const startStr = fmt.format(start)
  const endStr = fmt.format(end)
  return `${startStr} – ${endStr}`
}

function discountLabel(level: DealEvent['discountLevel']): string {
  switch (level) {
    case 'low':
      return 'Faibles remises'
    case 'medium':
      return 'Remises moyennes'
    case 'high':
      return 'Fortes remises'
  }
}

function discountBadgeClass(level: DealEvent['discountLevel']): string {
  switch (level) {
    case 'low':
      return styles.discountLow
    case 'medium':
      return styles.discountMedium
    case 'high':
      return styles.discountHigh
  }
}

function barClass(level: DealEvent['discountLevel']): string {
  switch (level) {
    case 'low':
      return styles.barLow
    case 'medium':
      return styles.barMedium
    case 'high':
      return styles.barHigh
  }
}

function isActive(event: DealEvent): boolean {
  const now = Date.now()
  return now >= new Date(event.startDate).getTime() && now <= new Date(event.endDate).getTime()
}

function isPast(event: DealEvent): boolean {
  return Date.now() > new Date(event.endDate).getTime()
}

function matchesCategory(event: DealEvent, selected: string): boolean {
  if (selected === 'all') return true
  return event.categoryHint === 'all' || event.categoryHint === selected
}

function EventCard({ event }: { event: DealEvent }) {
  const active = isActive(event)
  const past = isPast(event)

  let cardClass = styles.card
  if (active) cardClass += ` ${styles.cardActive}`
  if (past) cardClass += ` ${styles.cardPast}`

  return (
    <article className={cardClass}>
      <div className={`${styles.bar} ${barClass(event.discountLevel)}`} />
      <div className={styles.body}>
        <h2 className={styles.eventName}>{event.name}</h2>
        <div className={styles.meta}>
          <span className={styles.dateRange}>
            {formatDateRange(event.startDate, event.endDate)}
          </span>
          {event.categoryHint !== 'all' && (
            <span className={`${styles.badge} ${styles.categoryBadge}`}>
              {event.categoryHint}
            </span>
          )}
          <span className={`${styles.badge} ${discountBadgeClass(event.discountLevel)}`}>
            {discountLabel(event.discountLevel)}
          </span>
          {active && (
            <span className={`${styles.badge} ${styles.activeBadge}`}>En cours</span>
          )}
        </div>
        <p className={styles.description}>{event.description}</p>
      </div>
    </article>
  )
}

export default function DealCalendarPage() {
  const [selectedCategory, setSelectedCategory] = useState('all')
  const { events, isLoading } = useDealCalendar()

  const filtered = events.filter((e) => matchesCategory(e, selectedCategory))

  return (
    <>
      <Topnav />
      <main className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Calendrier des bons plans</h1>
          <p className={styles.subtitle}>
            Événements promotionnels du marché français
          </p>
        </div>

        <div className={styles.filters}>
          {CATEGORY_FILTERS.map(({ label, value }) => (
            <button
              key={value}
              className={
                selectedCategory === value
                  ? `${styles.filterBtn} ${styles.filterBtnActive}`
                  : styles.filterBtn
              }
              onClick={() => setSelectedCategory(value)}
            >
              {label}
            </button>
          ))}
        </div>

        {isLoading ? (
          <div className={styles.timeline}>
            {[1, 2, 3, 4].map((k) => (
              <div key={k} className={styles.skeletonCard} />
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <div className={styles.empty}>
            <div className={styles.emptyIcon}>📅</div>
            <div className={styles.emptyTitle}>
              AUCUN ÉVÉNEMENT POUR CETTE CATÉGORIE
            </div>
          </div>
        ) : (
          <div className={styles.timeline}>
            {filtered.map((event) => (
              <EventCard key={`${event.name}-${event.startDate}`} event={event} />
            ))}
          </div>
        )}
      </main>
    </>
  )
}

import { useTranslation } from 'react-i18next'
import { EmptyState } from '../scouter'
import { computePurchaseTimeline } from '../../utils/purchaseTimeline'
import type { ShoppingItem } from '../../types'
import styles from './PurchaseTimeline.module.css'

interface PurchaseTimelineProps {
  items: ShoppingItem[]
}

/**
 * Visual Gantt-style timeline showing recommended purchase windows for shopping items.
 * Based on price trends, deal periods, and target price status.
 */
export function PurchaseTimeline({ items }: PurchaseTimelineProps) {
  const { t } = useTranslation()

  if (items.length === 0) {
    return (
      <EmptyState
        icon="📅"
        title={t('timeline.empty').toUpperCase()}
        description={t('timeline.emptyDesc')}
      />
    )
  }

  // Compute timeline items from shopping items
  const timelineItems = computePurchaseTimeline(items)

  // Get current date and calculate date range
  const now = new Date()
  const timelineEnd = new Date(now)
  timelineEnd.setDate(timelineEnd.getDate() + 90)

  // Generate month labels for next 3 months
  const getMonthLabels = () => {
    const labels = []
    for (let i = 0; i < 3; i++) {
      const date = new Date(now)
      date.setMonth(date.getMonth() + i)
      const monthName = date.toLocaleString('en-US', { month: 'short', year: '2-digit' })
      labels.push(monthName)
    }
    return labels
  }

  const monthLabels = getMonthLabels()

  // Calculate bar position and width as percentages
  const getTotalDays = () => 90
  const totalDays = getTotalDays()

  const calculateBarPosition = (windowStart: Date): number => {
    const daysFromNow = Math.floor((windowStart.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    return Math.max(0, (daysFromNow / totalDays) * 100)
  }

  const calculateBarWidth = (windowStart: Date, windowEnd: Date): number => {
    const startDays = Math.floor((windowStart.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    const endDays = Math.floor((windowEnd.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    const duration = endDays - startDays
    return Math.min(100, (duration / totalDays) * 100)
  }

  return (
    <div className={styles.root}>
      <h2 className={styles.title}>📅 {t('timeline.title')}</h2>

      <div className={styles.ganttWrapper}>
        {/* Month header */}
        <div className={styles.monthHeader}>
          <div></div>
          <div className={styles.monthLabels}>
            {monthLabels.map((month) => (
              <div key={month} className={styles.monthLabel}>
                {month}
              </div>
            ))}
          </div>
        </div>

        {/* Timeline rows */}
        {timelineItems.map((item) => {
          const position = calculateBarPosition(item.windowStart)
          const width = calculateBarWidth(item.windowStart, item.windowEnd)

          return (
            <div key={item.itemId} className={styles.row}>
              <div className={styles.itemInfo}>
                <div className={styles.itemName} title={item.itemName}>
                  {item.itemName}
                </div>
                <div className={styles.merchant}>{item.merchant}</div>
              </div>

              <div className={styles.barContainer}>
                <div
                  className={`${styles.bar} ${styles[item.priority]}`}
                  style={{
                    left: `${position}%`,
                    width: `${width}%`,
                  }}
                  title={item.reason}
                >
                  {position + width < 90 ? item.reason : ''}
                </div>
                <div className={`${styles.badge} ${styles[item.priority]}`}>
                  {item.priority === 'now' && 'MAINTENANT'}
                  {item.priority === 'soon' && 'BIENTÔT'}
                  {item.priority === 'wait' && 'ATTENDRE'}
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Legend */}
      <div className={styles.legend}>
        <div className={styles.legendTitle}>{t('timeline.legend')}</div>
        <div className={styles.legendItems}>
          <div className={styles.legendItem}>
            <div className={`${styles.legendColor} ${styles.now}`}></div>
            <span>{t('timeline.priorityNow')}</span>
          </div>
          <div className={styles.legendItem}>
            <div className={`${styles.legendColor} ${styles.soon}`}></div>
            <span>{t('timeline.prioritySoon')}</span>
          </div>
          <div className={styles.legendItem}>
            <div className={`${styles.legendColor} ${styles.wait}`}></div>
            <span>{t('timeline.priorityWait')}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

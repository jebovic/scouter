import { useMemo } from 'react'
import { getUpcomingHolidays } from '../../utils/frenchHolidays'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import styles from './HolidayWidget.module.css'

interface HolidayWidgetProps {
  maxItems?: number
  compact?: boolean
}

export function HolidayWidget({ maxItems = 4, compact = false }: HolidayWidgetProps) {
  const { locale } = useFormatCurrency()
  const holidays = useMemo(() => {
    return getUpcomingHolidays(maxItems)
  }, [maxItems])

  if (holidays.length === 0) {
    return null
  }

  const formatter = new Intl.DateTimeFormat(locale, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })

  const formatDate = (dateStr: string): string => {
    const [year, month, day] = dateStr.split('-')
    return formatter.format(new Date(Number(year), Number(month) - 1, Number(day)))
  }

  if (compact) {
    return (
      <section className={styles.widgetCompact} role="region" aria-label="Jours fériés à venir">
        <ul className={styles.compactList}>
          {holidays.map((holiday) => (
            <li key={holiday.date} className={styles.compactItem}>
              <span className={styles.compactDate}>{formatDate(holiday.date)}</span>
              <span className={styles.compactName}>{holiday.name}</span>
              {holiday.storesLikelyClosed && (
                <span className={styles.badgeClosed}>Fermés</span>
              )}
            </li>
          ))}
        </ul>
      </section>
    )
  }

  return (
    <section className={styles.widget} role="region" aria-label="Jours fériés à venir">
      <div className={styles.header}>
        <h2 className={styles.title}>🗓️ Prochains jours fériés</h2>
      </div>

      <table className={styles.table}>
        <thead>
          <tr>
            <th className={styles.colDate}>Date</th>
            <th className={styles.colName}>Jour férié</th>
            <th className={styles.colStatus}>Magasins</th>
          </tr>
        </thead>
        <tbody>
          {holidays.map((holiday) => (
            <tr key={holiday.date} className={styles.row}>
              <td className={styles.cellDate}>
                <code className={styles.dateCode}>{holiday.date}</code>
              </td>
              <td className={styles.cellName}>{holiday.name}</td>
              <td className={styles.cellStatus}>
                {holiday.storesLikelyClosed ? (
                  <span className={styles.badgeClosed}>Magasins fermés</span>
                ) : (
                  <span className={styles.badgeOpen}>Ouvert</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}

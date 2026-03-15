import { useSmartAlerts } from '../../hooks/useSmartAlerts'
import type { Alert } from '../../api/notificationrules'
import styles from './SmartAlertsPanel.module.css'

interface Props {
  missionId: string | undefined
}

function AlertRow({ alert }: { alert: Alert }) {
  // Extract the leading emoji as the icon, remainder as text
  const emojiMatch = alert.message.match(/^(\p{Emoji_Presentation}|\p{Emoji}\uFE0F)\s*/u)
  const icon = emojiMatch ? emojiMatch[0].trim() : '•'
  const text = emojiMatch ? alert.message.slice(emojiMatch[0].length) : alert.message

  return (
    <div className={styles.alert} data-severity={alert.severity}>
      <span className={styles.alertIcon} aria-hidden="true">{icon}</span>
      <span className={styles.alertMessage}>{text}</span>
    </div>
  )
}

export function SmartAlertsPanel({ missionId }: Props) {
  const { data, isLoading } = useSmartAlerts(missionId)

  return (
    <section className={styles.card} aria-label="Alertes intelligentes">
      <div className={styles.header}>
        <span className={styles.title}>Alertes intelligentes</span>
        {data && data.criticalCount > 0 && (
          <span className={styles.badge} aria-label={`${data.criticalCount} alerte(s) critique(s)`}>
            {data.criticalCount}
          </span>
        )}
      </div>

      {isLoading && (
        <div className={styles.alertList}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={styles.skeleton} aria-hidden="true" />
          ))}
        </div>
      )}

      {!isLoading && (!data || data.alerts.length === 0) && (
        <p className={styles.empty}>Aucune alerte active</p>
      )}

      {!isLoading && data && data.alerts.length > 0 && (
        <div className={styles.alertList}>
          {data.alerts.map((alert, idx) => (
            <AlertRow key={`${alert.itemId}-${alert.ruleId}-${idx}`} alert={alert} />
          ))}
        </div>
      )}
    </section>
  )
}

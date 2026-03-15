import { usePriceAlertDigest } from '../../hooks/usePriceAlertDigest'
import type { AlertEntry } from '../../api/pricealertdigest'
import { formatCurrency } from '../../utils/format'
import styles from './PriceAlertDigestPanel.module.css'

interface Props {
  missionId: string
  currency: string
}

function AlertRow({ alert, currency }: { alert: AlertEntry; currency: string }) {
  // Extract the leading emoji as the icon, remainder as text
  const emojiMatch = alert.message.match(/^(\p{Emoji_Presentation}|\p{Emoji}\uFE0F)\s*/u)
  const icon = emojiMatch ? emojiMatch[0].trim() : '•'
  const text = emojiMatch ? alert.message.slice(emojiMatch[0].length) : alert.message

  return (
    <div className={styles.alert} data-type={alert.alertType}>
      <span className={styles.alertIcon} aria-hidden="true">{icon}</span>
      <div className={styles.alertContent}>
        <div className={styles.alertName}>{alert.name}</div>
        <div className={styles.alertMessage}>{text}</div>
        <div className={styles.alertPrice}>
          <span className={styles.currentPrice}>{formatCurrency(alert.price, currency)}</span>
          {alert.savings > 0 && (
            <span className={styles.savings}>
              Économies: {formatCurrency(alert.savings, currency)}
            </span>
          )}
        </div>
      </div>
    </div>
  )
}

export function PriceAlertDigestPanel({ missionId, currency }: Props) {
  const { digest, isLoading } = usePriceAlertDigest(missionId)

  return (
    <section className={styles.card} aria-label="Alertes de prix">
      <div className={styles.header}>
        <span className={styles.title}>Alertes de prix</span>
        {digest && digest.alertCount > 0 && (
          <span className={styles.badge} aria-label={`${digest.alertCount} alerte(s)`}>
            {digest.alertCount}
          </span>
        )}
      </div>

      <div className={styles.summary}>{digest?.summary}</div>

      {isLoading && (
        <div className={styles.alertList}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={styles.skeleton} aria-hidden="true" />
          ))}
        </div>
      )}

      {!isLoading && (!digest || digest.alerts.length === 0) && (
        <p className={styles.empty}>Aucune alerte de prix active</p>
      )}

      {!isLoading && digest && digest.alerts.length > 0 && (
        <div className={styles.alertList}>
          {digest.alerts.map((alert, idx) => (
            <AlertRow key={`${alert.itemId}-${alert.alertType}-${idx}`} alert={alert} currency={currency} />
          ))}
        </div>
      )}
    </section>
  )
}

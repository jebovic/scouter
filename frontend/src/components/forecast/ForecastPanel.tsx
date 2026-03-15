import { useTranslation } from 'react-i18next'
import { useForecast, useRunForecast } from '../../hooks/useForecast'
import type { ForecastResult } from '../../api/forecast'
import styles from './ForecastPanel.module.css'

interface ForecastPanelProps {
  missionId: string
}

function ConfidenceBadge({ level }: { level: ForecastResult['confidence'] }) {
  const { t } = useTranslation()
  return (
    <span className={`${styles.confidenceBadge} ${styles[`confidence_${level}`]}`}>
      {t(`forecast.confidence_${level}`)}
    </span>
  )
}

export function ForecastPanel({ missionId }: ForecastPanelProps) {
  const { t } = useTranslation()
  const { forecast, isLoading } = useForecast(missionId)
  const { runForecast, isPending } = useRunForecast(missionId)

  return (
    <div className={styles.panel}>
      <div className={styles.panelHeader}>
        <h3 className={styles.panelTitle}>🔮 {t('forecast.title').toUpperCase()}</h3>
        <button
          onClick={() => runForecast()}
          disabled={isPending}
          className={`${styles.runBtn} ${isPending ? styles.runBtnPending : styles.runBtnActive}`}
        >
          {isPending ? `⚙ ${t('forecast.running')}` : t('forecast.run')}
        </button>
      </div>

      {isLoading && (
        <div className={styles.emptyState}>{t('common.loading')}</div>
      )}

      {!isLoading && !forecast && (
        <div className={styles.emptyState}>
          <div className={styles.emptyIcon}>🔮</div>
          <div className={styles.emptyTitle}>{t('forecast.empty')}</div>
          <div className={styles.emptyDesc}>{t('forecast.emptyDesc')}</div>
        </div>
      )}

      {forecast && (
        <div className={styles.forecastBody}>
          {/* Estimated total */}
          {forecast.estimatedTotal !== null && (
            <div className={styles.totalSection}>
              <div className={styles.totalLabel}>{t('forecast.estimatedTotal').toUpperCase()}</div>
              <div className={styles.totalValue}>~{forecast.estimatedTotal.toLocaleString()} €</div>
            </div>
          )}

          {/* Confidence + meta chips */}
          <div className={styles.chipsRow}>
            <div className={styles.chip}>
              <span className={styles.chipLabel}>{t('forecast.confidence').toUpperCase()}</span>
              <ConfidenceBadge level={forecast.confidence} />
            </div>

            {forecast.monthsToSave !== null && (
              <div className={styles.chip}>
                <span className={styles.chipLabel}>{t('forecast.monthsToSave').toUpperCase()}</span>
                <span className={styles.chipValue}>{forecast.monthsToSave} mo</span>
              </div>
            )}

            {forecast.optimalBuyTime && (
              <div className={styles.chip}>
                <span className={styles.chipLabel}>{t('forecast.optimalBuyTime').toUpperCase()}</span>
                <span className={styles.chipValue}>{forecast.optimalBuyTime}</span>
              </div>
            )}
          </div>

          {/* Recommendations */}
          {forecast.recommendations.length > 0 && (
            <div className={styles.listSection}>
              <div className={styles.listLabel}>{t('forecast.recommendations').toUpperCase()}</div>
              <ul className={`${styles.list} ${styles.listGreen}`}>
                {forecast.recommendations.map((rec, i) => (
                  <li key={i} className={styles.listItem}>
                    <span className={styles.iconGreen}>✓</span>
                    <span>{rec}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Risk factors */}
          {forecast.riskFactors.length > 0 && (
            <div className={styles.listSection}>
              <div className={styles.listLabel}>{t('forecast.riskFactors').toUpperCase()}</div>
              <ul className={`${styles.list} ${styles.listCoral}`}>
                {forecast.riskFactors.map((risk, i) => (
                  <li key={i} className={styles.listItem}>
                    <span className={styles.iconCoral}>⚠</span>
                    <span>{risk}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

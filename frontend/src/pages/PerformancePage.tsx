import { useTranslation } from 'react-i18next'
import { usePerformanceMetrics } from '../hooks/usePerformanceMetrics'
import styles from './PerformancePage.module.css'

function formatBytes(bytes: number | null): string {
  if (bytes === null) return 'N/A'
  return (bytes / 1024 / 1024).toFixed(1)
}

function getColorClass(
  value: number | null,
  thresholds: { green: number; gold: number }
): string {
  if (value === null) return 'neutral'
  if (value < thresholds.green) return 'green'
  if (value < thresholds.gold) return 'gold'
  return 'coral'
}

export default function PerformancePage() {
  const { t } = useTranslation()
  const { metrics, refresh } = usePerformanceMetrics()

  const ttfbColor = getColorClass(metrics.ttfb, { green: 200, gold: 500 })
  const loadTimeColor = getColorClass(metrics.loadTime, { green: 1000, gold: 3000 })
  const heapColor = getColorClass(metrics.usedJSHeapSize, { green: 50000000, gold: 100000000 })

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.heading}>{t('performance.title', 'Perfs')}</h1>
        <button className={styles.refreshBtn} onClick={refresh} aria-label="Refresh metrics">
          ⟲ {t('common.refresh', 'Refresh')}
        </button>
      </header>

      <div className={styles.grid}>
        {/* TTFB Card */}
        <div className={`${styles.metricCard} ${styles[ttfbColor]}`}>
          <div className={styles.metricName}>TTFB</div>
          <div className={styles.metricValue}>
            {metrics.ttfb !== null ? metrics.ttfb : 'N/A'}
          </div>
          <div className={styles.metricUnit}>ms</div>
          <div className={styles.metricLabel}>Time to First Byte</div>
        </div>

        {/* DOM Content Loaded Card */}
        <div className={`${styles.metricCard} ${styles[getColorClass(metrics.domContentLoaded, { green: 1000, gold: 2000 })]}`}>
          <div className={styles.metricName}>DOM Content Loaded</div>
          <div className={styles.metricValue}>
            {metrics.domContentLoaded !== null ? metrics.domContentLoaded : 'N/A'}
          </div>
          <div className={styles.metricUnit}>ms</div>
          <div className={styles.metricLabel}>Time to DOM Ready</div>
        </div>

        {/* Load Time Card */}
        <div className={`${styles.metricCard} ${styles[loadTimeColor]}`}>
          <div className={styles.metricName}>Load Time</div>
          <div className={styles.metricValue}>
            {metrics.loadTime !== null ? metrics.loadTime : 'N/A'}
          </div>
          <div className={styles.metricUnit}>ms</div>
          <div className={styles.metricLabel}>Full Page Load</div>
        </div>

        {/* JS Heap Used Card */}
        <div className={`${styles.metricCard} ${styles[heapColor]}`}>
          <div className={styles.metricName}>JS Heap Used</div>
          <div className={styles.metricValue}>
            {formatBytes(metrics.usedJSHeapSize)}
          </div>
          <div className={styles.metricUnit}>MB</div>
          <div className={styles.metricLabel}>JavaScript Memory</div>
        </div>

        {/* Connection Type Card */}
        <div className={`${styles.metricCard} ${styles.neutral}`}>
          <div className={styles.metricName}>Connection</div>
          <div className={styles.metricValue} style={{ fontSize: '1.5rem' }}>
            {metrics.connectionType ? metrics.connectionType.toUpperCase() : 'N/A'}
          </div>
          <div className={styles.metricUnit}></div>
          <div className={styles.metricLabel}>Effective Type</div>
        </div>

        {/* Total Heap Card (Info only) */}
        <div className={`${styles.metricCard} ${styles.neutral}`}>
          <div className={styles.metricName}>Total Heap</div>
          <div className={styles.metricValue}>
            {formatBytes(metrics.totalJSHeapSize)}
          </div>
          <div className={styles.metricUnit}>MB</div>
          <div className={styles.metricLabel}>JS Heap Size</div>
        </div>
      </div>

      <section className={styles.info}>
        <p className={styles.disclaimer}>
          {t('performance.disclaimer', 'Métriques mesurées en temps réel dans votre navigateur')}
        </p>
        <p className={styles.note}>
          {t(
            'performance.note',
            'Les métriques de performance varient selon votre connexion, votre appareil et votre navigateur. Navigation Timing API, Memory API (Chrome/Edge) et Network Information API.'
          )}
        </p>
      </section>
    </main>
  )
}
